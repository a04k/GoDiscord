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

	// Get sender's balance
	var senderBalance int
	err = b.Db.QueryRow("SELECT balance FROM users WHERE guild_id = $1 AND user_id = $2", m.GuildID, m.Author.ID).Scan(&senderBalance)
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
	err = b.Db.QueryRow("SELECT balance FROM users WHERE guild_id = $1 AND user_id = $2", m.GuildID, recipientID).Scan(&recipientBalance)
	if err == sql.ErrNoRows {
		// Recipient doesn't exist, create a new row for them
		_, err := b.Db.Exec("INSERT INTO users (guild_id, user_id, balance) VALUES ($1, $2, 0)", m.GuildID, recipientID)
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
	_, err = b.Db.Exec("UPDATE users SET balance = balance - $1 WHERE guild_id = $2 AND user_id = $3", amount, m.GuildID, m.Author.ID)
	if err != nil {
		log.Printf("Error updating sender balance: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
		return
	}

	// Add to recipient's balance
	_, err = b.Db.Exec("UPDATE users SET balance = balance + $1 WHERE guild_id = $2 AND user_id = $3", amount, m.GuildID, recipientID)
	if err != nil {
		log.Printf("Error updating recipient balance: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
		return
	}

	// Notify sender and recipient
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("You transferred %d coins to <@%s>.", amount, recipientID))
	s.ChannelMessageSend(recipientID, fmt.Sprintf("You received %d coins from <@%s> in server %s.", amount, m.Author.ID, m.GuildID))
}