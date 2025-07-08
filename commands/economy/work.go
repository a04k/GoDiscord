package economy

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
)

func Work(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	reward := rand.Intn(650-65) + 65

	var lastDaily sql.NullTime // using nulltime to handle nulls in db
	// even though the user is supposed to be registered to the bot with a timestamp, when a db problem happens / buggy code things the timestamp could be lost leaving a null
	var balance int

	// Check if the user exists
	err := b.Db.QueryRow("SELECT last_daily, balance FROM users WHERE user_id = $1", m.Author.ID).Scan(&lastDaily, &balance)
	if err == sql.ErrNoRows {
		// User doesn't exist, insert a new row / reg user
		_, err = b.Db.Exec("INSERT INTO users (user_id, balance, last_daily) VALUES ($1, $2, NOW())", m.Author.ID, reward)
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
		_, err := b.Db.Exec("INSERT INTO users (user_id, balance, last_daily) VALUES ($1, $2, NOW()) ON CONFLICT (user_id) DO UPDATE SET balance = users.balance + $2, last_daily = NOW()", m.Author.ID, reward)
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
}
