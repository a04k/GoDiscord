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
    last_daily TIMESTAMP
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

func getUSDEGP() (float64, error) {
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

	case "daily":
		var lastDaily time.Time
		var balance int

		// Check if the user exists
		err := b.db.QueryRow("SELECT last_daily, balance FROM users WHERE user_id = $1", m.Author.ID).Scan(&lastDaily, &balance)
		if err == sql.ErrNoRows {
			// User doesn't exist, insert a new row
			_, err := b.db.Exec("INSERT INTO users (user_id, balance, last_daily) VALUES ($1, 100, NOW())", m.Author.ID)
			if err != nil {
				log.Printf("Error inserting new user: %v", err)
				s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
				return
			}
			s.ChannelMessageSend(m.ChannelID, "You received 100 coins!")
			return
		} else if err != nil {
			log.Printf("Error querying user: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			return
		}

		// Check if 24 hours have passed since the last daily
		if lastDaily.IsZero() || time.Since(lastDaily) >= 24*time.Hour {
			_, err := b.db.Exec("UPDATE users SET balance = balance + 100, last_daily = NOW() WHERE user_id = $1", m.Author.ID)
			if err != nil {
				log.Printf("Error updating user balance: %v", err)
				s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
				return
			}
			s.ChannelMessageSend(m.ChannelID, "You received 100 coins!")
		} else {
			s.ChannelMessageSend(m.ChannelID, "Wait 24h before next daily")
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
			s.ChannelMessageSend(m.ChannelID, "Usage: .flip <amount>")
			return
		}

		amount := 0
		fmt.Sscanf(args[1], "%d", &amount)

		var balance int
		b.db.QueryRow("SELECT balance FROM users WHERE user_id = $1", m.Author.ID).Scan(&balance)

		if balance < amount {
			s.ChannelMessageSend(m.ChannelID, "Not enough coins")
			return
		}

		if rand.Intn(2) == 0 {
			b.db.Exec("UPDATE users SET balance = balance + $1 WHERE user_id = $2", amount, m.Author.ID)
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("You won %d coins!", amount))
		} else {
			b.db.Exec("UPDATE users SET balance = balance - $1 WHERE user_id = $2", amount, m.Author.ID)
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("You lost %d coins!", amount))
		}
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
