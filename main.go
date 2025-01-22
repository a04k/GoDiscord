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
    is_admin BOOLEAN DEFAULT FALSE,
    is_mod BOOLEAN DEFAULT FALSE
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

// ============================================ HELPER FUNCTIONS =========================================== //

func userExists(s *discordgo.Session, userID string) (bool, error) {
	_, err := s.User(userID)
	if err != nil {
		if strings.Contains(err.Error(), "Unknown User") {
			return false, nil // User does not exist
		}
		return false, err // Other error
	}
	return true, nil // User exists
}

func extractUserID(mention string) (string, error) {
	// Check if the mention is properly formatted
	if !strings.HasPrefix(mention, "<@") || !strings.HasSuffix(mention, ">") {
		return "", fmt.Errorf("invalid mention format")
	}

	// Extract the user ID
	userID := strings.TrimPrefix(strings.TrimSuffix(mention, ">"), "<@")

	// Validate that the user ID is a valid Snowflake (Discord ID)
	if _, err := strconv.ParseUint(userID, 10, 64); err != nil {
		return "", fmt.Errorf("invalid user ID")
	}

	return userID, nil
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

func (b *Bot) isModerator(userID string) (bool, error) {
	var isMod bool
	err := b.db.QueryRow("SELECT is_mod FROM users WHERE user_id = $1", userID).Scan(&isMod)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil // user not found, cant be mod
		}
		return false, err // db err
	}
	return isMod, nil
}

func (b *Bot) hasPermission(userID string) bool {
	isAdmin, err := b.isAdmin(userID)
	if err != nil {
		return false
	}

	isMod, err := b.isModerator(userID)
	if err != nil {
		return false
	}

	return isAdmin || isMod
}

// permission parsing function
func parsePermissions(permVal string) int {
	perms := 0
	switch strings.ToUpper(permVal) {
	case "MOD":
		// Mod bundle
		perms |= discordgo.PermissionManageRoles
		perms |= discordgo.PermissionChangeNickname
		perms |= discordgo.PermissionVoiceMuteMembers
		perms |= discordgo.PermissionManageMessages
	case "OWNER":
		// Owner
		perms |= discordgo.PermissionBanMembers
		perms |= discordgo.PermissionManageChannels
		perms |= discordgo.PermissionManageRoles
		perms |= discordgo.PermissionChangeNickname
		perms |= discordgo.PermissionVoiceMuteMembers
		perms |= discordgo.PermissionManageMessages
	case "DEFAULT":
		// Default user permissions (no extra permissions)
		// just return 0 for no permissions (this is the default behavior)
		return 0
	default:
		return 0 // If the permission is not recognized, return 0
	}
	return perms
}

// this fetches the highest role a mod can assign, cannot be above the users/mod's current role.
func getHighestRole(member *discordgo.Member, roles []*discordgo.Role) *discordgo.Role {
	var highestRole *discordgo.Role
	for _, roleID := range member.Roles {
		for _, role := range roles {
			if role.ID == roleID {
				if highestRole == nil || role.Position > highestRole.Position {
					highestRole = role
				}
			}
		}
	}
	return highestRole
}

// ============================================ BOT FUNCTIONALITY =========================================== //

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

// ================================ DOT COMMAND HANDLING =========================================

func (b *Bot) handleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if !strings.HasPrefix(m.Content, ".") {
		return
	}

	// Process commands that start with "."
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

	case "bal", "balance":
		var targetUserID string

		// Check if a mention is provided
		if len(args) >= 2 {
			// Extract and validate the target user mention
			recipientMention := args[1]
			var err error
			targetUserID, err = extractUserID(recipientMention)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Invalid mention / use. Please use a proper mention (e.g., @username).")
				return
			}
		} else {
			// If no mention is provided, use the author's ID
			targetUserID = m.Author.ID
		}

		// Query the balance of the target user
		var balance int
		err := b.db.QueryRow("SELECT balance FROM users WHERE user_id = $1", targetUserID).Scan(&balance)
		if err != nil {
			if err == sql.ErrNoRows {
				s.ChannelMessageSend(m.ChannelID, "User not found.")
			} else {
				log.Printf("Error querying balance: %v", err)
				s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			}
			return
		}

		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("<@%s>'s balance: %d coins", targetUserID, balance))

	case "work":
		reward := rand.Intn(650-65) + 65

		var lastDaily time.Time
		var balance int

		// Check if the user exists
		err := b.db.QueryRow("SELECT last_daily, balance FROM users WHERE user_id = $1", m.Author.ID).Scan(&lastDaily, &balance)
		if err == sql.ErrNoRows {
			// User doesn't exist, insert a new row
			_, err = b.db.Exec("INSERT INTO users (user_id, balance, last_daily) VALUES ($1, $2, NOW())", m.Author.ID, reward)
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
			// Remove seconds from the remaining time
			formattedRemaining := fmt.Sprintf("%d hours %d minutes", int(remaining.Hours()), int(remaining.Minutes())%60)
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("You have to wait %s before you can work again", formattedRemaining))
		}

	case "transfer":
		if len(args) < 3 {
			s.ChannelMessageSend(m.ChannelID, "Usage: .transfer <recipient> <amount>")
			return
		}

		// Extract and validate the recipient mention
		recipientMention := args[1]
		recipientID, err := extractUserID(recipientMention)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Invalid mention. Please use a proper mention (e.g., @username).")
			return
		}

		// Check if the recipient exists
		exists, err := userExists(s, recipientID)
		if err != nil {
			log.Printf("Error checking user existence: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			return
		}
		if !exists {
			s.ChannelMessageSend(m.ChannelID, "User not found. Please check the mention.")
			return
		}

		// Extract the amount
		amount := 0
		fmt.Sscanf(args[2], "%d", &amount)

		// Validate amount
		if amount <= 0 {
			s.ChannelMessageSend(m.ChannelID, "Amount must be greater than 0.")
			return
		}

		// Get sender's balance
		var senderBalance int
		err = b.db.QueryRow("SELECT balance FROM users WHERE user_id = $1", m.Author.ID).Scan(&senderBalance)
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
		err = b.db.QueryRow("SELECT balance FROM users WHERE user_id = $1", recipientID).Scan(&recipientBalance)
		if err == sql.ErrNoRows {
			// Recipient doesn't exist, create a new row for them
			_, err := b.db.Exec("INSERT INTO users (user_id, balance) VALUES ($1, 0)", recipientID)
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
		_, err = b.db.Exec("UPDATE users SET balance = balance + $1 WHERE user_id = $2", amount, recipientID)
		if err != nil {
			log.Printf("Error updating recipient balance: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			return
		}

		// Notify sender and recipient
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("You transferred %d coins to <@%s>.", amount, recipientID))
		s.ChannelMessageSend(recipientID, fmt.Sprintf("You received %d coins from <@%s>.", amount, m.Author.ID))

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

	case "lb":
		var users []struct {
			UserID  string
			Balance int
		}

		rows, err := b.db.Query("SELECT user_id, balance FROM users ORDER BY balance DESC")
		if err != nil {
			log.Printf("Error querying leaderboard: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			return
		}
		defer rows.Close()

		for rows.Next() {
			var user struct {
				UserID  string
				Balance int
			}
			if err := rows.Scan(&user.UserID, &user.Balance); err != nil {
				log.Printf("Error scanning leaderboard rows: %v", err)
				s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
				return
			}
			users = append(users, user)
		}
		if err := rows.Err(); err != nil {
			log.Printf("Error during leaderboard rows iteration: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			return
		}
		if len(users) == 0 {
			s.ChannelMessageSend(m.ChannelID, "No users found.")
			return
		}

		const pageSize = 10
		totalPages := (len(users) + pageSize - 1) / pageSize

		currentPage := 0

		sendLeaderboard := func(page int) (*discordgo.Message, error) {
			start := page * pageSize
			end := start + pageSize
			if end > len(users) {
				end = len(users)
			}

			embed := &discordgo.MessageEmbed{
				Title:       "Coin Leaderboard",
				Description: "Top users by coin balance:",
				Color:       0x00ff00, // Green color
				Fields:      make([]*discordgo.MessageEmbedField, 0),
				Footer: &discordgo.MessageEmbedFooter{
					Text: fmt.Sprintf("Page %d of %d", page+1, totalPages),
				},
			}

			for i := start; i < end; i++ {
				user := users[i]
				discordUser, err := s.User(user.UserID)
				if err != nil {
					log.Printf("Error fetching user: %v", err)
					continue
				}
				embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
					Name:  fmt.Sprintf("%d. %s", i+1, discordUser.Username),
					Value: fmt.Sprintf("%d coins", user.Balance),
				})
			}

			msg, err := s.ChannelMessageSendEmbed(m.ChannelID, embed)
			if err != nil {
				return nil, err
			}

			if totalPages > 1 {
				err = s.MessageReactionAdd(m.ChannelID, msg.ID, "⬅️")
				if err != nil {
					log.Printf("Error adding reaction: %v", err)
					return nil, err
				}
				err = s.MessageReactionAdd(m.ChannelID, msg.ID, "➡️")
				if err != nil {
					log.Printf("Error adding reaction: %v", err)
					return nil, err
				}
			}

			return msg, nil
		}

		msg, err := sendLeaderboard(currentPage)
		if err != nil {
			log.Printf("Error sending leaderboard message: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			return
		}

		// Reaction handling
		if totalPages > 1 {
			b.client.AddHandlerOnce(func(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
				if r.MessageID != msg.ID || r.UserID == s.State.User.ID {
					return
				}

				if r.Emoji.Name == "⬅️" {
					currentPage--
					if currentPage < 0 {
						currentPage = totalPages - 1
					}
				} else if r.Emoji.Name == "➡️" {
					currentPage++
					if currentPage >= totalPages {
						currentPage = 0
					}
				} else {
					return
				}

				err := s.MessageReactionsRemoveAll(m.ChannelID, msg.ID)
				if err != nil {
					log.Printf("Error removing reactions: %v", err)
					return
				}

				updatedMsg, err := sendLeaderboard(currentPage)
				if err != nil {
					log.Printf("Error updating leaderboard: %v", err)
				}

				s.ChannelMessageDelete(m.ChannelID, msg.ID)
				msg = updatedMsg
			})
		}

		// ============================= ADMIN ONLY ======================================

	case "ra", "removeadmin":
		if len(args) < 2 {
			s.ChannelMessageSend(m.ChannelID, "Usage: .ra <@user>")
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
		_, err = b.db.Exec("UPDATE users SET is_admin = FALSE WHERE user_id = $1", recipient)
		if err != nil {
			log.Printf("Error promoting user: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			return
		}

		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Removed <@%s> from bot adminstrators.", recipient))

	case "setadmin", "sa":
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

	case "ia", "isadmin":
		if len(args) < 2 {
			s.ChannelMessageSend(m.ChannelID, "Usage: .ia <@user> or .isadmin <@user>")
			return
		}

		// Check if the user executing the command is an admin
		var isAdmin bool
		err := b.db.QueryRow("SELECT is_admin FROM users WHERE user_id = $1", m.Author.ID).Scan(&isAdmin)
		if err != nil || !isAdmin {
			s.ChannelMessageSend(m.ChannelID, "You are not authorized to use this command.")
			return
		}

		// Extract the target user ID
		targetUserID, err := extractUserID(args[1])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Invalid mention. Please use a proper mention (e.g., @username).")
			return
		}

		// Check if the target user is an admin
		var targetIsAdmin bool
		err = b.db.QueryRow("SELECT is_admin FROM users WHERE user_id = $1", targetUserID).Scan(&targetIsAdmin)
		if err != nil {
			if err == sql.ErrNoRows {
				s.ChannelMessageSend(m.ChannelID, "User not found.")
			} else {
				log.Printf("Error querying target user admin status: %v", err)
				s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			}
			return
		}

		// Send the result
		if targetIsAdmin {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("<@%s> is an admin.", targetUserID))
		} else {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("<@%s> is not an admin.", targetUserID))
		}

	case "take":
		if len(args) < 3 {
			s.ChannelMessageSend(m.ChannelID, "Usage: .take <@user> <amount|all>")
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

		// Extract and validate the recipient mention
		recipientMention := args[1]
		recipientID, err := extractUserID(recipientMention)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Invalid mention. Please use a proper mention (e.g., @username).")
			return
		}

		// Check if the recipient exists
		exists, err := userExists(s, recipientID)
		if err != nil {
			log.Printf("Error checking user existence: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			return
		}
		if !exists {
			s.ChannelMessageSend(m.ChannelID, "User not found. Please check the mention.")
			return
		}

		// Get recipient's balance
		var recipientBalance int
		err = b.db.QueryRow("SELECT balance FROM users WHERE user_id = $1", recipientID).Scan(&recipientBalance)
		if err != nil {
			log.Printf("Error querying recipient balance: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			return
		}

		// Handle "all" case
		var amount int
		if args[2] == "all" {
			amount = recipientBalance // Take all coins
		} else {
			// Extract the amount
			_, err = fmt.Sscanf(args[2], "%d", &amount)
			if err != nil || amount <= 0 {
				s.ChannelMessageSend(m.ChannelID, "Invalid amount. Usage: .take <@user> <amount|all>")
				return
			}

			// Check if user has enough coins
			if recipientBalance < amount {
				s.ChannelMessageSend(m.ChannelID, "User does not have enough coins")
				return
			}
		}

		// Remove coins from recipient
		_, err = b.db.Exec("UPDATE users SET balance = balance - $1 WHERE user_id = $2", amount, recipientID)
		if err != nil {
			log.Printf("Error removing coins from user: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			return
		}

		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Removed %d coins from <@%s>", amount, recipientID))
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

		// Extract and validate the recipient mention
		recipientMention := args[1]
		recipientID, err := extractUserID(recipientMention)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Invalid mention. Please use a proper mention (e.g., @username).")
			return
		}

		// Extract the amount
		amount := 0
		fmt.Sscanf(args[2], "%d", &amount)

		// Validate amount
		if amount <= 0 {
			s.ChannelMessageSend(m.ChannelID, "Amount must be greater than 0.")
			return
		}

		// Check if the recipient exists
		var recipientBalance int
		err = b.db.QueryRow("SELECT balance FROM users WHERE user_id = $1", recipientID).Scan(&recipientBalance)
		if err != nil {
			if err == sql.ErrNoRows {
				// Recipient doesn't exist, create a new row for them with a balance of 0
				_, err = b.db.Exec("INSERT INTO users (user_id, balance) VALUES ($1, 0)", recipientID)
				if err != nil {
					log.Printf("Error creating user: %v", err)
					s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
					return
				}
				recipientBalance = 0 // Set balance to 0 for the new user
			} else {
				log.Printf("Error querying recipient balance: %v", err)
				s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
				return
			}
		}

		// Add coins to the recipient
		_, err = b.db.Exec("UPDATE users SET balance = balance + $1 WHERE user_id = $2", amount, recipientID)
		if err != nil {
			log.Printf("Error adding coins to user: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			return
		}

		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Added %d coins to <@%s>", amount, recipientID))

	case "createrole", "cr":
		if len(args) < 2 {
			s.ChannelMessageSend(m.ChannelID, "Usage: .createrole/cr <role name> [color hex] [permissions] [hoist]")
			return
		}

		// Fetch guild member details
		member, err := s.GuildMember(m.GuildID, m.Author.ID)
		if err != nil {
			log.Printf("Error fetching member: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred while verifying permissions.")
			return
		}

		// Check if the user has administrative permissions
		guild, err := s.Guild(m.GuildID)
		if err != nil {
			log.Printf("Error fetching guild: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred while fetching server information.")
			return
		}

		isAdmin := false
		for _, roleID := range member.Roles {
			for _, role := range guild.Roles {
				if role.ID == roleID && role.Permissions&discordgo.PermissionAdministrator != 0 {
					isAdmin = true
					break
				}
			}
		}

		if !isAdmin {
			s.ChannelMessageSend(m.ChannelID, "You must be an administrator to use this command.")
			return
		}

		// Extract role name
		roleName := args[1]
		roleColor := 0 // Default color (no color)
		hoist := false // Default hoisting (not visible)

		// Check for optional color hex
		if len(args) > 2 && strings.HasPrefix(args[2], "#") {
			colorHex := args[2]
			parsedColor, err := strconv.ParseInt(colorHex[1:], 16, 32)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Invalid color hex. Example: #FF5733")
				return
			}
			roleColor = int(parsedColor)
		}

		perms := 0
		if len(args) > 3 {
			permVal := args[3]
			perms = parsePermissions(permVal)
			if perms == 0 {
				s.ChannelMessageSend(m.ChannelID, "Invalid permissions input. Example: 'mod' / 'owner'")
				return
			}
		}

		// Check for hoist flag (optional)
		if len(args) > 4 && strings.ToLower(args[4]) == "hoist" {
			hoist = true // Enable hoisting if 'hoist' is provided as the 5th argument
		}

		// Prevent admin role creation
		if strings.Contains(roleName, "admin") {
			s.ChannelMessageSend(m.ChannelID, "You cannot create admin roles through this command. Admin roles must be added manually.")
			return
		}

		// Convert perms to int64 for the Permissions field
		permsInt64 := int64(perms)

		// Create the role
		newRole, err := s.GuildRoleCreate(m.GuildID, &discordgo.RoleParams{
			Name:        roleName,
			Color:       &roleColor,  // Use a pointer to int for color
			Permissions: &permsInt64, // Convert perms to int64 and pass a pointer
			Hoist:       &hoist,      // Use a pointer to bool for hoist
		})
		if err != nil {
			log.Printf("Error creating role: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred while creating the role.")
			return
		}

		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Role '%s' created successfully!", newRole.Name))

	// ==================================== MOD & ABOVE COMMANDS =======================================

	case "setrole", "sr":
		if len(args) < 3 {
			s.ChannelMessageSend(m.ChannelID, "Usage: .setrole/sr <@user> <role name or mention>")
			return
		}

		// Extract user mention and role name
		userMention := args[1]
		roleName := strings.Join(args[2:], " ")

		// Get user ID from mention
		if !strings.HasPrefix(userMention, "<@") || !strings.HasSuffix(userMention, ">") {
			s.ChannelMessageSend(m.ChannelID, "Invalid user mention. Example: .addrole @User RoleName")
			return
		}
		userID := userMention[2 : len(userMention)-1]
		if strings.HasPrefix(userID, "!") {
			userID = userID[1:] // Remove nickname exclamation mark
		}

		// Fetch guild and roles
		guild, err := s.Guild(m.GuildID)
		if err != nil {
			log.Printf("Error fetching guild: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred while fetching server information.")
			return
		}

		// Find the target role
		var targetRole *discordgo.Role
		if strings.HasPrefix(roleName, "<@&") && strings.HasSuffix(roleName, ">") {
			// Handle role mention
			roleID := roleName[3 : len(roleName)-1]
			for _, r := range guild.Roles {
				if r.ID == roleID {
					targetRole = r
					break
				}
			}
		} else {
			// Handle role name
			for _, r := range guild.Roles {
				if strings.EqualFold(r.Name, roleName) {
					targetRole = r
					break
				}
			}
		}

		if targetRole == nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Role '%s' not found.", roleName))
			return
		}

		// Fetch member executing the command
		executor, err := s.GuildMember(m.GuildID, m.Author.ID)
		if err != nil {
			log.Printf("Error fetching executor: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred while verifying permissions.")
			return
		}

		// Find the executor's highest role position
		var executorHighestPosition int
		for _, execRoleID := range executor.Roles {
			for _, role := range guild.Roles {
				if role.ID == execRoleID && role.Position > executorHighestPosition {
					executorHighestPosition = role.Position
				}
			}
		}

		// Ensure executor can assign the target role
		if targetRole.Position >= executorHighestPosition {
			s.ChannelMessageSend(m.ChannelID, "You cannot assign a role equal to or higher than your highest role.")
			return
		}

		// Add the role to the target user
		err = s.GuildMemberRoleAdd(m.GuildID, userID, targetRole.ID)
		if err != nil {
			log.Printf("Error adding role: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred while adding the role to the user.")
			return
		}

		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Role '%s' added to user <@%s> successfully!", targetRole.Name, userID))

	case "inrole":
		if len(args) < 2 {
			s.ChannelMessageSend(m.ChannelID, "Usage: .inrole <role name or mention>")
			return
		}

		// Combine arguments into a role name or mention
		input := strings.TrimSpace(strings.Join(args[1:], " "))
		log.Printf("Input received: '%s'", input)

		// Fetch guild information
		guild, err := s.Guild(m.GuildID)
		if err != nil {
			log.Printf("Error fetching guild: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred while fetching server information.")
			return
		}

		// Find the target role (by mention or name)
		var targetRole *discordgo.Role
		if strings.HasPrefix(input, "<@&") && strings.HasSuffix(input, ">") {
			// Handle role mention
			roleID := input[3 : len(input)-1] // Extract role ID
			for _, r := range guild.Roles {
				if r.ID == roleID {
					targetRole = r
					break
				}
			}
		} else {
			// Handle role name
			for _, r := range guild.Roles {
				if strings.EqualFold(r.Name, input) {
					targetRole = r
					break
				}
			}
		}

		if targetRole == nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Role '%s' not found.", input))
			return
		}

		log.Printf("Role found: '%s' (ID: %s)", targetRole.Name, targetRole.ID)

		// Fetch members in the guild
		var allMembers []*discordgo.Member
		after := ""
		for {
			members, err := s.GuildMembers(m.GuildID, after, 1000)
			if err != nil {
				log.Printf("Error fetching members: %v", err)
				s.ChannelMessageSend(m.ChannelID, "An error occurred while fetching members.")
				return
			}
			if len(members) == 0 {
				break
			}
			allMembers = append(allMembers, members...)
			after = members[len(members)-1].User.ID
		}

		// Filter members with the target role
		var membersInRole []string
		for _, member := range allMembers {
			for _, roleID := range member.Roles {
				if roleID == targetRole.ID {
					// Add nickname or username
					if member.Nick != "" {
						membersInRole = append(membersInRole, member.Nick)
					} else {
						membersInRole = append(membersInRole, member.User.Username)
					}
					break
				}
			}
		}

		if len(membersInRole) == 0 {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("No users found in role '%s'.", targetRole.Name))
			return
		}

		// Create and send the embed message
		embed := &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("Users in role: %s", targetRole.Name),
			Description: strings.Join(membersInRole, "\n"),
			Color:       targetRole.Color,
			Footer: &discordgo.MessageEmbedFooter{
				Text: fmt.Sprintf("Total members: %d", len(membersInRole)),
			},
		}

		s.ChannelMessageSendEmbed(m.ChannelID, embed)
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
