package admin

import (
	"fmt"
	"log"

	"DiscordBot/bot"
	"DiscordBot/commands"
	"DiscordBot/utils"
	"github.com/bwmarrin/discordgo"
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
	isAdmin, err := b.IsAdmin(m.GuildID, m.Author.ID)
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

	// Add coins to the recipient
	_, err = b.Db.Exec("UPDATE users SET balance = balance + $1 WHERE guild_id = $2 AND user_id = $3", amount, m.GuildID, recipientID)
	if err != nil {
		log.Printf("Error adding coins to user: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Added %d coins to <@%s>", amount, recipientID))
}
