package economy

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"

	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
	"DiscordBot/commands"
)

func init() {
	commands.RegisterCommand("flip", Flip)
}

func Flip(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// Ensure command is used in a guild
	if m.GuildID == "" {
		return // Don't respond to DMs
	}

	if len(args) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Usage: .flip <amount|all>")
		return
	}
	// Get the user's balance
	var balance int
	err := b.Db.QueryRow("SELECT balance FROM users WHERE guild_id = $1 AND user_id = $2", m.GuildID, m.Author.ID).Scan(&balance)
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
		_, err := b.Db.Exec("UPDATE users SET balance = balance + $1 WHERE guild_id = $2 AND user_id = $3", amount, m.GuildID, m.Author.ID)
		if err != nil {
			log.Printf("Error updating balance: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			return
		}
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("You won %d coins! Your new balance is %d.", amount, balance+amount))
	} else {
		// User loses
		_, err := b.Db.Exec("UPDATE users SET balance = balance - $1 WHERE guild_id = $2 AND user_id = $3", amount, m.GuildID, m.Author.ID)
		if err != nil {
			log.Printf("Error updating balance: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			return
		}
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("You lost %d coins! Your new balance is %d.", amount, balance-amount))
	}
}