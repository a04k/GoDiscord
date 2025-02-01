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

func extractRoleID(input string) string {
	if strings.HasPrefix(input, "<@&") && strings.HasSuffix(input, ">") {
		return input[3 : len(input)-1]
	}
	return input // Return as is for ID/name validation
}

// function to validate and find role
func findRole(s *discordgo.Session, guildID string, roleInput string) (*discordgo.Role, error) {
	roles, err := s.GuildRoles(guildID)
	if err != nil {
		return nil, err
	}

	cleanedInput := extractRoleID(roleInput)

	// Check by ID first
	for _, role := range roles {
		if role.ID == cleanedInput {
			return role, nil
		}
	}

	// Check by name (case-insensitive)
	cleanedInput = strings.ToLower(strings.TrimSpace(cleanedInput))
	for _, role := range roles {
		if strings.ToLower(role.Name) == cleanedInput {
			return role, nil
		}
	}

	return nil, fmt.Errorf("role not found")
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

func (b *Bot) getMutedRole(guildID string) (*discordgo.Role, error) {
	// fetch all roles in server
	roles, err := b.client.GuildRoles(guildID)
	if err != nil {
		return nil, fmt.Errorf("error fetching guild roles: %v", err)
	}

	// look for muted role if it exists (which it should, if not then it is handled below)
	for _, role := range roles {
		if role.Name == "Muted" {
			return role, nil
		}
	}

	// creating a muted role if it doesn't exist : this will really only be used for servers that havent been set up yet / new servers
	mutedRole, err := b.client.GuildRoleCreate(guildID, &discordgo.RoleParams{
		Name: "Muted",
	})
	if err != nil {
		return nil, fmt.Errorf("error creating muted role: %v", err)
	}

	// Set the role name to "Muted" and permissions to 0 (no permissions)
	_, err = b.client.GuildRoleEdit(guildID, mutedRole.ID, &discordgo.RoleParams{
		Name:        "Muted",
		Permissions: new(int64),
	})
	if err != nil {
		return nil, fmt.Errorf("error editing muted role: %v", err)
	}

	// Update channel permissions to restrict the "Muted" role from sending messages
	channels, err := b.client.GuildChannels(guildID)
	if err != nil {
		return nil, fmt.Errorf("error fetching guild channels: %v", err)
	}
	for _, channel := range channels {
		// Deny SendMessages permission for the "Muted" role
		err = b.client.ChannelPermissionSet(channel.ID, mutedRole.ID, discordgo.PermissionOverwriteTypeRole, 0, discordgo.PermissionSendMessages)
		if err != nil {
			return nil, fmt.Errorf("error setting channel permissions for muted role: %v", err)
		}

		//Deny Speak permission on voice channels too
		if channel.Type == discordgo.ChannelTypeGuildVoice {
			err = b.client.ChannelPermissionSet(channel.ID, mutedRole.ID, discordgo.PermissionOverwriteTypeRole, 0, discordgo.PermissionVoiceSpeak)
			if err != nil {
				return nil, fmt.Errorf("error setting voice channel permissions for muted role: %v", err)
			}
		}
	}

	return mutedRole, nil
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

	case "help", "h":
		if len(args) < 2 {
			embed := &discordgo.MessageEmbed{
				Title:       "Usage:",
				Description: ".h / .help [commandname] \n Use .cl / .commandlist to view all commands",
				Color:       0x0000FF, //blue left bar for commands list
			}

			s.ChannelMessageSendEmbed(m.ChannelID, embed)
			return
		}

		// If a command is specified, display individual command help
		command := strings.ToLower(args[1])
		var response string
		noPermissionMessage := "You do not have permission to access this command."

		// Check if the user is an admin or mod
		isMod, _ := b.isModerator(m.Author.ID)
		isOwner, _ := b.isAdmin(m.Author.ID)

		switch command {
		case "bal", "balance":
			response = ".bal or .balance\nCheck your balance."

		case "work":
			response = ".work\nEarn coins (6h cooldown)."

		case "flip":
			response = ".flip <amount / all>\nGamble coins: (specified amount / all)."

		case "transfer":
			response = ".transfer <@user> <amount>\nSend coins to another user."

		case "usd":
			response = ".usd [amount]\nUSD to EGP exchange rate."

		case "btc":
			response = ".btc\nBitcoin price in USD."

		case "setup":
			response = "/setup <landline> <password>\nSave WE credentials."

		case "quota":
			response = "/quota\nCheck internet quota (after setup)."

		// Admin Commands
		case "add":
			if isOwner {
				response = ".add <@user> <amount>\nAdmin: Add coins to a user."
			} else {
				response = noPermissionMessage
			}

		case "sa":
			if isOwner {
				response = ".sa <@user>\nAdmin: Promote a user to admin."
			} else {
				response = noPermissionMessage
			}

		case "createrole":
			if isOwner {
				response = ".createrole <role name> [color] [permissions] [hoist]\nCreate a new role with options for color, permissions, and hoisting."
			} else {
				response = noPermissionMessage
			}

		case "setrole", "sr":
			if isOwner || isMod {
				response = ".setrole <@user> <role name>\nGives the user the role."
			} else {
				response = noPermissionMessage
			}

		case "inrole":
			if isOwner || isMod {
				response = ".inrole <role name or mention>\nView all users in a specific role."
			} else {
				response = noPermissionMessage
			}

		case "roleinfo", "ri":
			if isOwner || isMod {
				response = ".ri / roleinfo <role name or mention>\nView role information."
			} else {
				response = noPermissionMessage
			}

		case "ban":
			if isOwner {
				response = ".ban <@user> [reason] [days]\nBan a user with optional reason and days of message deletion."
			} else {
				response = noPermissionMessage
			}

		default:
			response = "Command not found. Use .help to get a list of available commands."
		}

		// Send command help in an embed
		embed := &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("Help: %s", command),
			Description: response,
			Color:       0xFF0000, // red left bar for command help
		}
		s.ChannelMessageSendEmbed(m.ChannelID, embed)

	case "commandlist", "cl":

		embed := &discordgo.MessageEmbed{
			Title:       "Available Commands",
			Description: "Here is a list of all available commands:",
			Color:       0x0000FF, //blue left bar for commands list

			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "User Commands",
					Value:  "`bal / balance` - Check your balance.\n`work` - Earn coins (6h cooldown).\n`flip <amount / all>` - Gamble coins.\n`transfer <@user> <amount>` - Send coins to another user.\n`usd [amount]` - USD to EGP exchange rate.\n`btc` - Bitcoin price in USD.\n`/setup <landline> <password>` - Save WE credentials.\n`/quota` - Check internet quota (after setup).",
					Inline: false,
				},
				{
					Name:   "Moderator Commands",
					Value:  "`mute <@user> [reason]` - Mute a user.\n`unmute <@user>` - Unmute a user.\n`voicemute/vm <@user> [reason]` - Mute a user's voice.\n`voiceunmute/vum <@user>` - Unmute a user's voice.",
					Inline: false,
				},
				{
					Name:   "Admin Commands",
					Value:  "`add <@user> <amount>` - add: Add coins to a user.\n`sa <@user>` - sa/setadmin: Promote a user to admin.\n`cr/createrole <role name> [color] [permissions] [hoist]` - Create a new role.\n`sr/setrole <@user> <role name>` - Assign role to user.\n`inrole <role name or mention>` - View users in a role.\n `ri/roleinfo <role name or mention>` - View role information.\n`ban <@user> [reason(OPTIONAL)] [days(OPTIONAL)]` - Ban a user with optional reason and days of message deletion.",
					Inline: false,
				},
			},
		}

		s.ChannelMessageSendEmbed(m.ChannelID, embed)

	case "usd":
		// default amount = 1
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
		// Check if a mention is provided & validate
		if len(args) >= 2 {
			// Extract and validate the target user mention
			recipientMention := args[1]
			var err error
			targetUserID, err = extractUserID(recipientMention)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Invalid mention / use. Please use a proper mention (e.g., @username).")
				return
			}

			// Check if mentioned user is in the guild
			member, err := s.GuildMember(m.GuildID, targetUserID)
			if err != nil || member == nil {
				s.ChannelMessageSend(m.ChannelID, "mentioned user is not in this server.")
				return
			}
		} else {
			// If no mention is provided, use the author's ID
			targetUserID = m.Author.ID
		}

		// register the user if they don't exist
		_, err := b.db.Exec(`
        INSERT INTO users (user_id, balance)
        VALUES ($1, $2)
        ON CONFLICT (user_id) DO NOTHING`,
			targetUserID, 0)
		if err != nil {
			log.Printf("Error in user registration: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			return
		}

		// Now query their balance (will exist due to prior insert)
		var balance int
		err = b.db.QueryRow("SELECT balance FROM users WHERE user_id = $1", targetUserID).Scan(&balance)
		if err != nil {
			log.Printf("Error querying balance: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			return
		}

		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("<@%s>'s balance: %d coins", targetUserID, balance))

	case "work":
		reward := rand.Intn(650-65) + 65

		var lastDaily sql.NullTime // using nulltime to handle nulls in db
		// even though the user is supposed to be registered to the bot with a timestamp, when a db problem happens / buggy code things the timestamp could be lost leaving a null
		var balance int

		// Check if the user exists
		err := b.db.QueryRow("SELECT last_daily, balance FROM users WHERE user_id = $1", m.Author.ID).Scan(&lastDaily, &balance)
		if err == sql.ErrNoRows {
			// User doesn't exist, insert a new row / reg user
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
		if !lastDaily.Valid || time.Since(lastDaily.Time) >= 6*time.Hour {
			_, err := b.db.Exec("INSERT INTO users (user_id, balance, last_daily) VALUES ($1, $2, NOW()) ON CONFLICT (user_id) DO UPDATE SET balance = users.balance + $2, last_daily = NOW()", m.Author.ID, reward)
			if err != nil {
				log.Printf("Error updating user balance: %v", err)
				s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
				return
			}
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("You received %d coins!", reward))
		} else {
			remaining := 6*time.Hour - time.Since(lastDaily.Time)
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

	case "kick":
		// Check if the user is an admin
		isOwner, err := b.isAdmin(m.Author.ID)
		if err != nil || !isOwner {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to use this command.")
			return
		}

		// Check if the user provided at least a target user
		if len(args) < 2 {
			s.ChannelMessageSend(m.ChannelID, "Usage: .kick <@user> [reason]")
			return
		}

		// Extract target user
		targetUser, err := extractUserID(args[1])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Invalid user / use. Please use a proper mention (e.g., @username).")
			return
		}

		// Set default reason
		reason := "No reason provided - .kick used"
		if len(args) >= 3 {
			reason = strings.Join(args[2:], " ")
		}

		// Kick the user
		err = s.GuildMemberDeleteWithReason(m.GuildID, targetUser, reason)
		if err != nil {
			log.Printf("Error kicking user: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred while kicking the user.")
			return
		}

		// kick confirmation message
		embed := &discordgo.MessageEmbed{
			Title:       "User Kicked",
			Description: fmt.Sprintf("Kicked user <@%s>", targetUser),
			Color:       0xFF0000, // Red bar on the left
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Reason",
					Value:  reason,
					Inline: false,
				},
			},
		}

		// send the embed msg
		s.ChannelMessageSendEmbed(m.ChannelID, embed)

	case "mute", "m":
		// Check if the user is an admin or moderator
		isMod, err := b.isModerator(m.Author.ID)
		if err != nil {
			log.Printf("Error checking mod status: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			return
		}

		isOwner, err := b.isAdmin(m.Author.ID)
		if err != nil {
			log.Printf("Error checking admin status: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			return
		}
		if !isMod && !isOwner {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to use this command.")
			return
		}

		// Check if the user provided at least a target user
		if len(args) < 2 {
			s.ChannelMessageSend(m.ChannelID, "Usage: .mute <@user> [reason]")
			return
		}

		// Extract target user
		targetUser, err := extractUserID(args[1])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Invalid user / use. Please use a proper mention (e.g., @username).")
			return
		}

		// Set default reason
		reason := "No reason provided - .mute used"
		if len(args) >= 3 {
			reason = strings.Join(args[2:], " ")
		}

		// Get the Muted role
		mutedRole, err := b.getMutedRole(m.GuildID)
		if err != nil {
			log.Printf("Error getting muted role: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred while getting the muted role.")
			return
		}

		// Add the Muted role to the user
		err = s.GuildMemberRoleAdd(m.GuildID, targetUser, mutedRole.ID)
		if err != nil {
			log.Printf("Error muting user: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred while muting the user.")
			return
		}

		// Mute confirmation message
		embed := &discordgo.MessageEmbed{
			Title:       "User Muted",
			Description: fmt.Sprintf("Muted user <@%s>", targetUser),
			Color:       0xFF0000, // Red bar on the left
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Reason",
					Value:  reason,
					Inline: false,
				},
			},
		}

		// Send the embed message
		s.ChannelMessageSendEmbed(m.ChannelID, embed)

	case "unmute", "um":
		// Check if the user is an admin or moderator
		isMod, err := b.isModerator(m.Author.ID)
		if err != nil {
			log.Printf("Error checking mod status: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			return
		}

		isOwner, err := b.isAdmin(m.Author.ID)
		if err != nil {
			log.Printf("Error checking admin status: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			return
		}

		if !isMod && !isOwner {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to use this command.")
			return
		}

		// Check if the user provided at least a target user
		if len(args) < 2 {
			s.ChannelMessageSend(m.ChannelID, "Usage: .unmute <@user>")
			return
		}

		// Extract target user
		targetUser, err := extractUserID(args[1])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Invalid user / use. Please use a proper mention (e.g., @username).")
			return
		}

		// Get the Muted role
		mutedRole, err := b.getMutedRole(m.GuildID)
		if err != nil {
			log.Printf("Error getting muted role: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred while getting the muted role.")
			return
		}

		// Remove the Muted role from the user
		err = s.GuildMemberRoleRemove(m.GuildID, targetUser, mutedRole.ID)
		if err != nil {
			log.Printf("Error unmuting user: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred while unmuting the user.")
			return
		}

		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Unmuted user <@%s>", targetUser))

	case "voicemute", "vm":
		// Check if the user is an admin or moderator
		isMod, err := b.isModerator(m.Author.ID)
		if err != nil {
			log.Printf("Error checking mod status: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			return
		}

		isOwner, err := b.isAdmin(m.Author.ID)
		if err != nil {
			log.Printf("Error checking admin status: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			return
		}

		if !isMod && !isOwner {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to use this command.")
			return
		}

		// Check if the user provided at least a target user
		if len(args) < 2 {
			s.ChannelMessageSend(m.ChannelID, "Usage: .voicemute/vm <@user> [reason]")
			return
		}

		// Extract target user
		targetUser, err := extractUserID(args[1])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Invalid user / use. Please use a proper mention (e.g., @username).")
			return
		}

		// Set default reason
		reason := "No reason provided - .mute used"
		if len(args) >= 3 {
			reason = strings.Join(args[2:], " ")
		}

		// Mute the user
		err = s.GuildMemberMute(m.GuildID, targetUser, true)
		if err != nil {
			log.Printf("Error muting user: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred while muting the user.")
			return
		}

		// mute confirmation message
		embed := &discordgo.MessageEmbed{
			Title:       "User Muted",
			Description: fmt.Sprintf("Muted user <@%s>", targetUser),
			Color:       0xFF0000, // Red bar on the left
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Reason",
					Value:  reason,
					Inline: false,
				},
			},
		}

		// send the embed msg
		s.ChannelMessageSendEmbed(m.ChannelID, embed)

	case "vunmute", "vum":
		// Check if the user is an admin or moderator
		isMod, err := b.isModerator(m.Author.ID)
		if err != nil {
			log.Printf("Error checking mod status: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			return
		}

		isOwner, err := b.isAdmin(m.Author.ID)
		if err != nil {
			log.Printf("Error checking admin status: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			return
		}

		if !isMod && !isOwner {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to use this command.")
			return
		}

		// Check if the user provided at least a target user
		if len(args) < 2 {
			s.ChannelMessageSend(m.ChannelID, "Usage: .voiceunmute/vum <@user>")
			return
		}

		// Extract target user
		targetUser, err := extractUserID(args[1])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Invalid user / use. Please use a proper mention (e.g., @username).")
			return
		}

		// Unmute the user
		err = s.GuildMemberMute(m.GuildID, targetUser, false)
		if err != nil {
			log.Printf("Error unmuting user: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred while unmuting the user.")
			return
		}
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Unmuted user <@%s>", targetUser))

	case "ban":
		// Check if the user is the owner
		isOwner, err := b.isAdmin(m.Author.ID) // Assuming b.isAdmin checks if the user is the owner
		if err != nil || !isOwner {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to use this command.")
			return
		}

		// Check if the user provided at least a target user
		if len(args) < 2 {
			s.ChannelMessageSend(m.ChannelID, "Usage: .ban @user [reason] [days_to_delete_messages]")
			return
		}

		// Extract target user
		targetUser, err := extractUserID(args[1])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Invalid user / use. Please use a proper mention (e.g., @username).")
			return
		}

		// Set default reason
		reason := "No reason provided - .ban used"
		if len(args) >= 3 {
			reason = strings.Join(args[2:], " ")
		}

		// Default message deletion days = zero
		msgDelDays := 0
		if len(args) >= 4 {
			days, err := strconv.Atoi(args[3])
			if err == nil {
				msgDelDays = days
			}
		}

		// Throw the ban hammer
		err = s.GuildBanCreateWithReason(m.GuildID, targetUser, reason, msgDelDays)
		if err != nil {
			log.Printf("Error banning user: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred while banning the user.")
			return
		}

		// ban confirmation message
		embed := &discordgo.MessageEmbed{
			Title:       "User Banned",
			Description: fmt.Sprintf("Banned user <@%s>", targetUser),
			Color:       0xFF0000, // Red bar on the left
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Reason",
					Value:  reason,
					Inline: false,
				},
				{
					Name:   "Message Deletion",
					Value:  fmt.Sprintf("%d days", msgDelDays),
					Inline: false,
				},
			},
		}

		// send the embed msg
		s.ChannelMessageSendEmbed(m.ChannelID, embed)

	case "unban":
		if len(args) >= 2 {
			targetUser := args[1]

			// exec remove ban => guildbandelete
			err := s.GuildBanDelete(m.GuildID, targetUser)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Failed to unban user: %s", err))
				return
			}
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Unbanned user <@%s>", targetUser))
		} else {
			s.ChannelMessageSend(m.ChannelID, "Please specify a user ID to unban.")
		}

	case "setadmin", "sa":
		if len(args) < 2 {
			s.ChannelMessageSend(m.ChannelID, "Usage: .sa <@user>")
			return
		}

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

		recipient := strings.TrimPrefix(strings.TrimSuffix(args[1], ">"), "<@")

		// Promote the user to admin
		_, err = b.db.Exec("UPDATE users SET is_admin = TRUE WHERE user_id = $1", recipient)
		if err != nil {
			log.Printf("Error promoting user: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			return
		}

		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Promoted <@%s> to admin.", recipient))

	case "ra", "removeadmin":
		if len(args) < 2 {
			s.ChannelMessageSend(m.ChannelID, "Usage: .ra <@user>")
			return
		}

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

		recipient := strings.TrimPrefix(strings.TrimSuffix(args[1], ">"), "<@")

		// Promote the user to admin
		_, err = b.db.Exec("UPDATE users SET is_admin = FALSE WHERE user_id = $1", recipient)
		if err != nil {
			log.Printf("Error promoting user: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			return
		}

		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Removed <@%s> from bot adminstrators.", recipient))

	case "ia", "isadmin":
		if len(args) < 2 {
			s.ChannelMessageSend(m.ChannelID, "Usage: .ia <@user> or .isadmin <@user>")
			return
		}

		// Admin check
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

		// Extract the target user ID
		targetUserID, err := extractUserID(args[1])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Invalid mention. Please use a proper mention (e.g., @username).")
			return
		}

		// Check if the target user is an admin
		targetIsAdmin, _ := b.isAdmin(targetUserID)

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
			s.ChannelMessageSend(m.ChannelID, "Usage: .createrole/cr <role name> [color hex] [permissions : (ε: default, mod/owner)] [hoist: (ε:false , hoist/true/1)]")
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
		mentionable := false

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
		if len(args) > 4 && (strings.ToLower(args[4]) == "hoist" || strings.ToLower(args[4]) == "true" || strings.ToLower(args[4]) == "1") {
			hoist = true // Enable hoisting if 'hoist' is provided as the 5th argument
		}
		if len(args) > 5 && (strings.ToLower(args[4]) == "hoist" || strings.ToLower(args[4]) == "true" || strings.ToLower(args[4]) == "1") {
			mentionable = true // makes role mentionable if provided as a 6th arg
		}

		// perms to int64
		permsInt64 := int64(perms)

		// executing the role creation call (GuildRoleCreate)
		newRole, err := s.GuildRoleCreate(m.GuildID, &discordgo.RoleParams{ // roleparams requiring perms val in int64 is stupid since perms max value admin = 8
			Name:        roleName,
			Color:       &roleColor,  // pointer to int for color
			Permissions: &permsInt64, // convert perms to int64 and pass a pointer
			Hoist:       &hoist,      // using a pointer to bool for hoist (hoisting: having the role appear on the member list)
			Mentionable: &mentionable,
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

	case "roleinfo", "ri":
		if len(args) < 2 {
			s.ChannelMessageSend(m.ChannelID, "Please specify a role name, ID, or mention.")
			return
		}

		roleInput := strings.Join(args[1:], " ")
		targetRole, err := findRole(s, m.GuildID, roleInput)

		if err != nil {
			log.Printf("Role lookup error: %v | Input: %s", err, roleInput)
			s.ChannelMessageSend(m.ChannelID, "Role not found or cannot be fetched.")
			return
		}

		createdTime, _ := discordgo.SnowflakeTimestamp(targetRole.ID)

		var keyPerms []string
		perms := targetRole.Permissions
		if perms&discordgo.PermissionAdministrator != 0 {
			keyPerms = append(keyPerms, "Administrator")
		}
		if perms&discordgo.PermissionManageRoles != 0 {
			keyPerms = append(keyPerms, "Manage Roles")
		}
		if perms&discordgo.PermissionManageMessages != 0 {
			keyPerms = append(keyPerms, "Manage Messages")
		}
		if perms&discordgo.PermissionBanMembers != 0 {
			keyPerms = append(keyPerms, "Ban Members")
		}
		if perms&discordgo.PermissionKickMembers != 0 {
			keyPerms = append(keyPerms, "Kick Members")
		}
		if perms&discordgo.PermissionModerateMembers != 0 {
			keyPerms = append(keyPerms, "Timeout Members")
		}

		permString := "No key permissions"
		if len(keyPerms) > 0 {
			permString = strings.Join(keyPerms, ", ")
		}

		embed := &discordgo.MessageEmbed{
			Title: "Role Information",
			Color: targetRole.Color,
			Fields: []*discordgo.MessageEmbedField{
				{Name: "ID", Value: targetRole.ID, Inline: true},
				{Name: "Name", Value: targetRole.Name, Inline: true},
				{Name: "Color", Value: fmt.Sprintf("#%06x", targetRole.Color), Inline: true},

				{Name: "Mention", Value: fmt.Sprintf("<@&%s>", targetRole.ID), Inline: true},
				{Name: "Hoisted", Value: map[bool]string{true: "Yes", false: "No"}[targetRole.Hoist], Inline: true},
				{Name: "Position", Value: strconv.Itoa(targetRole.Position), Inline: true},

				{Name: "Mentionable", Value: map[bool]string{true: "Yes", false: "No"}[targetRole.Mentionable], Inline: true},
				{Name: "Managed", Value: map[bool]string{true: "Yes", false: "No"}[targetRole.Managed], Inline: true},
				{Name: "\u200b", Value: "\u200b", Inline: true},

				{Name: "Key Permissions", Value: "```\n" + permString + "\n```", Inline: false},
				{Name: "Created At", Value: createdTime.Format("January 2, 2006"), Inline: false},
			},
		}

		if _, err := s.ChannelMessageSendEmbed(m.ChannelID, embed); err != nil {
			log.Printf("Failed to send roleinfo embed: %v", err)
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

		// test the credentials first
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

		// test if we can actually get the quota
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

	log.Println("GoBot is running. Press Ctrl+C to exit.")
	select {}

}
