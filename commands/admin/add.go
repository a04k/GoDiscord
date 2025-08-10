package admin

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
	commands.RegisterCommand("add", Add)
}

func Add(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 3 {
		s.ChannelMessageSend(m.ChannelID, "Usage: .add <@user> <amount>")
		return
	}

	// Check if the user is an admin
	isAdmin, err := utils.IsAdmin(b.Db, m.GuildID, m.Author.ID)
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
			// Recipient doesn't exist, create a new row for them with a balance of 0
			_, err = b.Db.Exec("INSERT INTO users (user_id, balance) VALUES ($1, 0)", recipientID)
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
	_, err = b.Db.Exec("UPDATE users SET balance = balance + $1 WHERE user_id = $2", amount, recipientID)
	if err != nil {
		log.Printf("Error adding coins to user: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Added %d coins to <@%s>", amount, recipientID))
}