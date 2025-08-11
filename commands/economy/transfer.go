package economy

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
	"DiscordBot/utils"
	"DiscordBot/commands"
)

func init() {
	commands.RegisterCommand("transfer", Transfer)
}

func Transfer(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// Ensure command is used in a guild
	if m.GuildID == "" {
		return // Don't respond to DMs
	}

	if len(args) < 3 {
		s.ChannelMessageSend(m.ChannelID, "Usage: .transfer <recipient> <amount>")
		return
	}

	// Extract and validate the recipient mention
	recipientMention := args[1]
	recipientID, err := utils.ExtractUserID(recipientMention)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Invalid mention. Please use a proper mention (e.g., @username).")
		return
	}

	// Check if the recipient exists
	exists, err := utils.UserExists(s, recipientID)
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

	// Ensure sender exists in global users table
	_, err = b.Db.Exec(`
		INSERT INTO users (user_id, username, avatar, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			username = EXCLUDED.username,
			avatar = EXCLUDED.avatar,
			updated_at = NOW()
	`, m.Author.ID, m.Author.Username, m.Author.AvatarURL(""))
	if err != nil {
		log.Printf("Error upserting sender: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
		return
	}

	// Ensure sender exists in guild_members table
	_, err = b.Db.Exec(`
		INSERT INTO guild_members (guild_id, user_id, joined_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (guild_id, user_id) DO NOTHING
	`, m.GuildID, m.Author.ID)
	if err != nil {
		log.Printf("Error ensuring sender guild member: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
		return
	}

	// Ensure recipient exists in global users table
	recipient, err := s.User(recipientID)
	if err != nil {
		log.Printf("Error getting recipient user: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
		return
	}

	_, err = b.Db.Exec(`
		INSERT INTO users (user_id, username, avatar, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			username = EXCLUDED.username,
			avatar = EXCLUDED.avatar,
			updated_at = NOW()
	`, recipientID, recipient.Username, recipient.AvatarURL(""))
	if err != nil {
		log.Printf("Error upserting recipient: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
		return
	}

	// Ensure recipient exists in guild_members table
	_, err = b.Db.Exec(`
		INSERT INTO guild_members (guild_id, user_id, joined_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (guild_id, user_id) DO NOTHING
	`, m.GuildID, recipientID)
	if err != nil {
		log.Printf("Error ensuring recipient guild member: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
		return
	}

	// Get sender's balance
	var senderBalance int
	err = b.Db.QueryRow("SELECT balance FROM guild_members WHERE guild_id = $1 AND user_id = $2", m.GuildID, m.Author.ID).Scan(&senderBalance)
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
	err = b.Db.QueryRow("SELECT balance FROM guild_members WHERE guild_id = $1 AND user_id = $2", m.GuildID, recipientID).Scan(&recipientBalance)
	if err == sql.ErrNoRows {
		// This shouldn't happen since we just inserted the row, but just in case
		recipientBalance = 0
	} else if err != nil {
		log.Printf("Error querying recipient balance: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
		return
	}

	// Deduct from sender's balance
	_, err = b.Db.Exec("UPDATE guild_members SET balance = balance - $1 WHERE guild_id = $2 AND user_id = $3", amount, m.GuildID, m.Author.ID)
	if err != nil {
		log.Printf("Error updating sender balance: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
		return
	}

	// Add to recipient's balance
	_, err = b.Db.Exec("UPDATE guild_members SET balance = balance + $1 WHERE guild_id = $2 AND user_id = $3", amount, m.GuildID, recipientID)
	if err != nil {
		log.Printf("Error updating recipient balance: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
		return
	}

	

	// Notify sender and recipient
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("You transferred %d coins to <@%s>.", amount, recipientID))
	s.ChannelMessageSend(recipientID, fmt.Sprintf("You received %d coins from <@%s> in server %s.", amount, m.Author.ID, m.GuildID))
}