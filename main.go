package main

import (
	"QCheckWE"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Bot struct {
	db     *sql.DB
	client *discordgo.Session
}

type ExchangeResponse struct {
	Result float64 `json:"result"`
}

type BTCResponse struct {
	Bitcoin struct {
		USD float64 `json:"usd"`
	} `json:"bitcoin"`
}

const schema = `
CREATE TABLE IF NOT EXISTS users (
    user_id TEXT PRIMARY KEY,
    balance INTEGER DEFAULT 0,
    last_daily TIMESTAMP,
    is_admin BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS credentials (
    user_id TEXT PRIMARY KEY,
    landline TEXT NOT NULL,
    password TEXT NOT NULL
);`

func NewBot(token string, dbURL string) (*Bot, error) {
	client, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}
	client.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(schema); err != nil {
		return nil, err
	}

	return &Bot{db: db, client: client}, nil
}

func (b *Bot) isAdmin(userID string) (bool, error) {
	var isAdmin bool
	err := b.db.QueryRow("SELECT is_admin FROM users WHERE user_id = $1", userID).Scan(&isAdmin)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil // user not found, cant be admin
		}
		return false, err // db err
	}
	return isAdmin, nil
}

func getUSDEGP() (float64, error) {
	// scrape town
	// URL for USD to EGP on Google Finance
	url := "https://www.google.com/finance/quote/USD-EGP"

	// Fetch the HTML content

	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to fetch URL: status code %d", resp.StatusCode)
	}

	// Parse the HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to parse HTML: %v", err)
	}

	// Find the exchange rate element
	rateText := doc.Find(".YMlKec.fxKbKc").First().Text()
	if rateText == "" {
		return 0, fmt.Errorf("failed to find exchange rate element")
	}

	// Clean and convert the rate text to a float
	rateText = strings.TrimSpace(rateText)
	rate, err := strconv.ParseFloat(rateText, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse exchange rate: %v", err)
	}

	return rate, nil
}

func getBTCPrice() (float64, error) {
	resp, err := http.Get("https://api.coingecko.com/api/v3/simple/price?ids=bitcoin&vs_currencies=usd")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result BTCResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}
	return result.Bitcoin.USD, nil
}

func (b *Bot) handleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if !strings.HasPrefix(m.Content, ".") {
		return
	}

	args := strings.Fields(m.Content)
	cmd := strings.ToLower(args[0][1:])

	switch cmd {
	case "usd":
		// Default amount is 1
		amount := 1.0

		// Check if the user provided an amount
		if len(args) > 1 {
			var err error
			amount, err = strconv.ParseFloat(args[1], 64)
			if err != nil || amount <= 0 {
				s.ChannelMessageSend(m.ChannelID, "Invalid amount. Usage: .usd [amount]")
				return
			}
		}

		rate, err := getUSDEGP()
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error fetching exchange rate")
			return
		}

		equivalent := amount * rate

		if amount == 1 {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("1 USD = %.2f EGP", rate))
		} else {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%.2f USD = %.2f EGP", amount, equivalent))
		}

	case "btc":
		price, err := getBTCPrice()
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error fetching BTC price")
			return
		}
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("BTC Price: $%.2f", price))

	case "bal":
		var balance int
		b.db.QueryRow("SELECT balance FROM users WHERE user_id = $1", m.Author.ID).Scan(&balance)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Balance: %d coins", balance))

	case "work":
		reward := rand.Intn(650-65) + 65

		var lastDaily time.Time
		var balance int

		// Check if the user exists
		err := b.db.QueryRow("SELECT last_daily, balance FROM users WHERE user_id = $1", m.Author.ID).Scan(&lastDaily, &balance)
		if err == sql.ErrNoRows {
			// User doesn't exist, insert a new row
			if err != nil {
				log.Printf("Error inserting new user: %v", err)
				s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
				return
			}
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("You received %d coins!", reward))
			return
		} else if err != nil {
			log.Printf("Error querying user: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			return
		}

		// Check if 6 hours have passed since the last work
		if lastDaily.IsZero() || time.Since(lastDaily) >= 6*time.Hour {
			_, err := b.db.Exec("INSERT INTO users (user_id, balance, last_daily) VALUES ($1, $2, NOW()) ON CONFLICT (user_id) DO UPDATE SET balance = users.balance + $2, last_daily = NOW()", m.Author.ID, reward)
			if err != nil {
				log.Printf("Error updating user balance: %v", err)
				s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
				return
			}
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("You received %d coins!", reward))
		} else {
			remaining := 6*time.Hour - time.Since(lastDaily)
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("You have to wait %s before you can work again", remaining.Round(time.Minute).String()))
		}

	case "transfer":
		if len(args) < 3 {
			s.ChannelMessageSend(m.ChannelID, "Usage: .transfer <recipient> <amount>")
			return
		}

		// Extract recipient and amount
		recipient := args[1]
		amount := 0
		fmt.Sscanf(args[2], "%d", &amount)

		// Validate amount
		if amount <= 0 {
			s.ChannelMessageSend(m.ChannelID, "Amount must be greater than 0.")
			return
		}

		// Get sender's balance
		var senderBalance int
		err := b.db.QueryRow("SELECT balance FROM users WHERE user_id = $1", m.Author.ID).Scan(&senderBalance)
		if err == sql.ErrNoRows {
			s.ChannelMessageSend(m.ChannelID, "You have 0 coins.")
			return
		} else if err != nil {
			log.Printf("Error querying sender balance: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			return
		}

		// Check if sender has enough coins
		if senderBalance < amount {
			s.ChannelMessageSend(m.ChannelID, "Not enough coins to transfer.")
			return
		}

		// Get recipient's balance
		var recipientBalance int
		err = b.db.QueryRow("SELECT balance FROM users WHERE user_id = $1", recipient).Scan(&recipientBalance)
		if err == sql.ErrNoRows {
			// Recipient doesn't exist, create a new row for them
			_, err := b.db.Exec("INSERT INTO users (user_id, balance) VALUES ($1, 0)", recipient)
			if err != nil {
				log.Printf("Error creating recipient: %v", err)
				s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
				return
			}
			recipientBalance = 0
		} else if err != nil {
			log.Printf("Error querying recipient balance: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			return
		}

		// Deduct from sender's balance
		_, err = b.db.Exec("UPDATE users SET balance = balance - $1 WHERE user_id = $2", amount, m.Author.ID)
		if err != nil {
			log.Printf("Error updating sender balance: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			return
		}

		// Add to recipient's balance
		_, err = b.db.Exec("UPDATE users SET balance = balance + $1 WHERE user_id = $2", amount, recipient)
		if err != nil {
			log.Printf("Error updating recipient balance: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			return
		}

		// Notify sender and recipient
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("You transferred %d coins to <@%s>.", amount, recipient))
		s.ChannelMessageSend(recipient, fmt.Sprintf("You received %d coins from <@%s>.", amount, m.Author.ID))

	case "flip":
		if len(args) < 2 {
			s.ChannelMessageSend(m.ChannelID, "Usage: .flip <amount|all>")
			return
		}
		// Get the user's balance
		var balance int
		err := b.db.QueryRow("SELECT balance FROM users WHERE user_id = $1", m.Author.ID).Scan(&balance)
		if err == sql.ErrNoRows {
			s.ChannelMessageSend(m.ChannelID, "You have 0 coins.")
			return
		} else if err != nil {
			log.Printf("Error querying balance: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			return
		}

		// Handle "all" case by setting amount to the user's total balance
		amount := 0
		if args[1] == "all" {
			amount = balance
		} else {
			// Parse the amount
			_, err = fmt.Sscanf(args[1], "%d", &amount)
			if err != nil || amount <= 0 {
				s.ChannelMessageSend(m.ChannelID, "Invalid amount. Usage: .flip <amount|all>")
				return
			}
		}

		// Check if the user has enough coins
		if balance < amount {
			s.ChannelMessageSend(m.ChannelID, "Not enough coins.")
			return
		}

		// Flip the coins
		if rand.Intn(2) == 0 {
			// User wins
			_, err := b.db.Exec("UPDATE users SET balance = balance + $1 WHERE user_id = $2", amount, m.Author.ID)
			if err != nil {
				log.Printf("Error updating balance: %v", err)
				s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
				return
			}
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("You won %d coins! Your new balance is %d.", amount, balance+amount))
		} else {
			// User loses
			_, err := b.db.Exec("UPDATE users SET balance = balance - $1 WHERE user_id = $2", amount, m.Author.ID)
			if err != nil {
				log.Printf("Error updating balance: %v", err)
				s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
				return
			}
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("You lost %d coins! Your new balance is %d.", amount, balance-amount))
		}

	case "sa":
		if len(args) < 2 {
			s.ChannelMessageSend(m.ChannelID, "Usage: .sa <@user>")
			return
		}

		// Check if the user is an admin
		var isAdmin bool
		err := b.db.QueryRow("SELECT is_admin FROM users WHERE user_id = $1", m.Author.ID).Scan(&isAdmin)
		if err != nil || !isAdmin {
			s.ChannelMessageSend(m.ChannelID, "You are not authorized to use this command.")
			return
		}

		recipient := strings.TrimPrefix(strings.TrimSuffix(args[1], ">"), "<@")

		// Promote the user to admin
		_, err = b.db.Exec("UPDATE users SET is_admin = TRUE WHERE user_id = $1", recipient)
		if err != nil {
			log.Printf("Error promoting user: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			return
		}

		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Promoted <@%s> to admin.", recipient))

		//ADMIN ONLY
	case "take":
		if len(args) < 3 {
			s.ChannelMessageSend(m.ChannelID, "Usage: .take <@user> <amount>")
			return
		}

		// Check if the user is an admin
		isAdmin, err := b.isAdmin(m.Author.ID)
		if err != nil {
			log.Printf("Error checking admin status: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			return
		}
		if !isAdmin {
			s.ChannelMessageSend(m.ChannelID, "You are not authorized to use this command.")
			return
		}

		// Extract recipient and amount
		recipient := strings.TrimPrefix(strings.TrimSuffix(args[1], ">"), "<@")
		amount := 0
		fmt.Sscanf(args[2], "%d", &amount)

		// Validate amount
		if amount <= 0 {
			s.ChannelMessageSend(m.ChannelID, "Amount must be greater than 0.")
			return
		}

		// Check if the recipient exists
		var userExists bool
		err = b.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1)", recipient).Scan(&userExists)
		if err != nil {
			log.Printf("Error checking user existence: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			return
		}

		if !userExists {
			s.ChannelMessageSend(m.ChannelID, "User not found.")
			return
		}

		// Get recipient's balance
		var recipientBalance int
		err = b.db.QueryRow("SELECT balance FROM users WHERE user_id = $1", recipient).Scan(&recipientBalance)
		if err != nil {
			log.Printf("Error querying recipient balance: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			return
		}

		// Check if user has enough coins
		if recipientBalance < amount {
			s.ChannelMessageSend(m.ChannelID, "User does not have enough coins")
			return
		}

		// Remove coins from recipient
		_, err = b.db.Exec("UPDATE users SET balance = balance - $1 WHERE user_id = $2", amount, recipient)
		if err != nil {
			log.Printf("Error removing coins from user: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			return
		}

		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Removed %d coins from <@%s>", amount, recipient))

	case "remove":
		if len(args) < 3 {
			s.ChannelMessageSend(m.ChannelID, "Usage: .remove <@user> <amount>")
			return
		}

		// Check if the user is an admin
		isAdmin, err := b.isAdmin(m.Author.ID)
		if err != nil {
			log.Printf("Error checking admin status: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			return
		}
		if !isAdmin {
			s.ChannelMessageSend(m.ChannelID, "You are not authorized to use this command.")
			return
		}

		// Extract recipient and amount
		recipient := strings.TrimPrefix(strings.TrimSuffix(args[1], ">"), "<@")
		amount := 0
		fmt.Sscanf(args[2], "%d", &amount)

		// Validate amount
		if amount <= 0 {
			s.ChannelMessageSend(m.ChannelID, "Amount must be greater than 0.")
			return
		}

		// Check if the recipient exists
		var userExists bool
		err = b.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1)", recipient).Scan(&userExists)
		if err != nil {
			log.Printf("Error checking user existence: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			return
		}

		if !userExists {
			s.ChannelMessageSend(m.ChannelID, "User not found.")
			return
		}

		// Get recipient's balance
		var recipientBalance int
		err = b.db.QueryRow("SELECT balance FROM users WHERE user_id = $1", recipient).Scan(&recipientBalance)
		if err != nil {
			log.Printf("Error querying recipient balance: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			return
		}

		// If the amount is greater than the users balance set amount to balance
		if amount > recipientBalance {
			amount = recipientBalance
		}

		// Remove coins from recipient
		_, err = b.db.Exec("UPDATE users SET balance = balance - $1 WHERE user_id = $2", amount, recipient)
		if err != nil {
			log.Printf("Error removing coins from user: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			return
		}

		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("stole %d coins from <@%s> ", amount, recipient))

	case "add":
		if len(args) < 3 {
			s.ChannelMessageSend(m.ChannelID, "Usage: .add <@user> <amount>")
			return
		}

		// Check if the user is an admin
		isAdmin, err := b.isAdmin(m.Author.ID)
		if err != nil {
			log.Printf("Error checking admin status: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			return
		}
		if !isAdmin {
			s.ChannelMessageSend(m.ChannelID, "You are not authorized to use this command.")
			return
		}

		// Extract recipient and amount
		recipient := strings.TrimPrefix(strings.TrimSuffix(args[1], ">"), "<@")
		amount := 0
		fmt.Sscanf(args[2], "%d", &amount)

		// Validate amount
		if amount <= 0 {
			s.ChannelMessageSend(m.ChannelID, "Amount must be greater than 0.")
			return
		}

		// Check if the recipient exists
		var userExists bool
		err = b.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1)", recipient).Scan(&userExists)
		if err != nil {
			log.Printf("Error checking user existence: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			return
		}

		// Add coins to the recipient
		if !userExists {
			_, err := b.db.Exec("INSERT INTO users (user_id, balance) VALUES ($1, $2)", recipient, amount)
			if err != nil {
				log.Printf("Error creating user: %v", err)
				s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
				return
			}
		} else {
			_, err = b.db.Exec("UPDATE users SET balance = balance + $1 WHERE user_id = $2", amount, recipient)
			if err != nil {
				log.Printf("Error adding coins to user: %v", err)
				s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
				return
			}
		}

		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Added %d coins to <@%s>", amount, recipient))
	}
}

func (b *Bot) handleSlashCommands(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	switch i.ApplicationCommandData().Name {
	case "quota":
		// First try to get saved credentials
		var landline, password string
		err := b.db.QueryRow("SELECT landline, password FROM credentials WHERE user_id = $1", i.Member.User.ID).Scan(&landline, &password)

		if err == sql.ErrNoRows {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Please set up your credentials first using /setup",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		} else if err != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Error retrieving credentials",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		checker, err := QCheckWE.NewWeQuotaChecker(landline, password)
		if err != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Error: %v", err),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		quota, err := checker.CheckQuota()
		if err != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Error: %v", err),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		msg := fmt.Sprintf(`
			Customer: %s
			Plan: %s
			Remaining: %.2f / %.2f (%s%% Used)
			Renewed: %s
			Expires: %s (%s)`,
			quota["name"],
			quota["offerName"],
			quota["remaining"],
			quota["total"],
			quota["usagePercentage"],
			quota["renewalDate"],
			quota["expiryDate"],
			quota["expiryIn"])

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: msg,
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})

	case "setup":
		options := i.ApplicationCommandData().Options
		landline := options[0].StringValue()
		password := options[1].StringValue()

		// Test the credentials first
		checker, err := QCheckWE.NewWeQuotaChecker(landline, password)
		if err != nil {
			log.Printf("Error creating WE Quota Checker: %v", err)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Invalid credentials",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		// Test if we can actually get the quota
		_, err = checker.CheckQuota()
		if err != nil {
			log.Printf("Error checking quota: %v", err)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Invalid credentials",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		// Save the credentials
		_, err = b.db.Exec(`
                INSERT INTO credentials (user_id, landline, password)
                VALUES ($1, $2, $3)
                ON CONFLICT (user_id)
                DO UPDATE SET landline = $2, password = $3
            `, i.Member.User.ID, landline, password)

		if err != nil {
			log.Printf("Error saving credentials: %v", err)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Error saving credentials",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Credentials saved successfully! You can now use /quota",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}
}

func main() {
	godotenv.Load()

	bot, err := NewBot(os.Getenv("DISCORD_TOKEN"), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}

	bot.client.AddHandler(bot.handleMessage)
	bot.client.AddHandler(bot.handleSlashCommands)

	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "quota",
			Description: "Check your internet quota using saved credentials",
		},
		{
			Name:        "setup",
			Description: "Save your WE credentials",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "landline",
					Description: "Your landline number",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "password",
					Description: "Your password",
					Required:    true,
				},
			},
		},
	}

	if err := bot.client.Open(); err != nil {
		log.Fatal(err)
	}
	defer bot.client.Close()

	// Register all slash commands
	for _, cmd := range commands {
		_, err := bot.client.ApplicationCommandCreate(bot.client.State.User.ID, "", cmd)
		if err != nil {
			log.Printf("Error creating command %s: %v", cmd.Name, err)
		}
	}

	log.Println("Bot is running. Press Ctrl+C to exit.")
	select {}
}
