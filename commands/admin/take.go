package admin

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
	"DiscordBot/utils"
)

func Take(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 3 {
		s.ChannelMessageSend(m.ChannelID, "Usage: .take <@user> <amount>")
		return
	}

	// Check if the user is an admin
	isAdmin, err := utils.IsAdmin(b.Db, m.Author.ID)
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
	recipientID, err := utils.ExtractUserID(recipientMention)
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
	err = b.Db.QueryRow("SELECT balance FROM users WHERE user_id = $1", recipientID).Scan(&recipientBalance)
	if err != nil {
		if err == sql.ErrNoRows {
			// Recipient doesn't exist
			s.ChannelMessageSend(m.ChannelID, "User not found in the database.")
			return
		} else {
			log.Printf("Error querying recipient balance: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			return
		}
	}

	// Check if the recipient has enough coins
	if recipientBalance < amount {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("User <@%s> only has %d coins, cannot take %d.", recipientID, recipientBalance, amount))
		return
	}

	// Remove coins from the recipient
	_, err = b.Db.Exec("UPDATE users SET balance = balance - $1 WHERE user_id = $2", amount, recipientID)
	if err != nil {
		log.Printf("Error removing coins from user: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Took %d coins from <@%s>", amount, recipientID))
}