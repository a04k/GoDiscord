package economyadmin

import (
	"fmt"
	"log"

	"DiscordBot/bot"
	"DiscordBot/utils"

	"github.com/bwmarrin/discordgo"
)

// Take function is registered via the module.go file

func Take(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 3 {
		s.ChannelMessageSend(m.ChannelID, "Usage: .take <@user> <amount|all>")
		return
	}

	// Check if the user has administrator permissions
	hasAdmin, err := utils.CheckAdminPermission(s, m.GuildID, m.Author.ID)
	if err != nil {
		log.Printf("Error checking admin status: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
		return
	}

	if !hasAdmin {
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

	// Get the recipient's balance
	var recipientBalance int
	err = b.Db.QueryRow("SELECT balance FROM users WHERE guild_id = $1 AND user_id = $2", m.GuildID, recipientID).Scan(&recipientBalance)
	if err != nil {
		log.Printf("Error querying recipient balance: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
		return
	}

	// Handle "all" case by setting amount to the recipient's total balance
	amount := 0
	if args[2] == "all" {
		amount = recipientBalance
	} else {
		// Parse the amount
		_, err = fmt.Sscanf(args[2], "%d", &amount)
		if err != nil || amount <= 0 {
			s.ChannelMessageSend(m.ChannelID, "Invalid amount. Usage: .take <@user> <amount|all>")
			return
		}
	}

	// Check if the recipient has enough coins (only needed for specific amounts, not "all")
	if args[2] != "all" && recipientBalance < amount {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("User <@%s> only has %d coins, cannot take %d.", recipientID, recipientBalance, amount))
		return
	}

	// Remove coins from the recipient
	_, err = b.Db.Exec("UPDATE users SET balance = balance - $1 WHERE guild_id = $2 AND user_id = $3", amount, m.GuildID, recipientID)
	if err != nil {
		log.Printf("Error removing coins from user: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
		return
	}

	if args[2] == "all" {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Took all %d coins from <@%s>", amount, recipientID))
	} else {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Took %d coins from <@%s>", amount, recipientID))
	}
}
