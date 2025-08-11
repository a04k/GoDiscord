package f1

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
	"DiscordBot/commands"
)

func init() {
	commands.RegisterCommand("f1sub", F1Subscribe)
}

func F1Subscribe(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// Ensure user exists in global users table
	_, err := b.Db.Exec(`
		INSERT INTO users (user_id, username, avatar, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			username = EXCLUDED.username,
			avatar = EXCLUDED.avatar,
			updated_at = NOW()
	`, m.Author.ID, m.Author.Username, m.Author.AvatarURL(""))
	if err != nil {
		log.Printf("Error upserting user: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
		return
	}

	userID := m.Author.ID

	var exists bool
	err = b.Db.QueryRow("SELECT EXISTS(SELECT 1 FROM f1_subscriptions WHERE user_id = $1)", userID).Scan(&exists)
	if err != nil {
		log.Printf("Error checking F1 subscription status for user %s: %v", userID, err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred while checking your subscription status.")
		return
	}

	var responseContent string
	if exists {
		// User is subscribed, so unsubscribe them
		_, err = b.Db.Exec("DELETE FROM f1_subscriptions WHERE user_id = $1", userID)
		if err != nil {
			log.Printf("Error unsubscribing user %s from F1 notifications: %v", userID, err)
			responseContent = "An error occurred while unsubscribing you from F1 notifications."
		} else {
			responseContent = "You have successfully unsubscribed from F1 notifications."
		}
	} else {
		// User is not subscribed, so subscribe them
		_, err = b.Db.Exec("INSERT INTO f1_subscriptions (user_id) VALUES ($1)", userID)
		if err != nil {
			log.Printf("Error subscribing user %s to F1 notifications: %v", userID, err)
			responseContent = "An error occurred while subscribing you to F1 notifications."
		} else {
			responseContent = "You have successfully subscribed to F1 notifications! I will DM you before each F1 event and 30 minutes before each session."
		}
	}

	s.ChannelMessageSend(m.ChannelID, responseContent)
}
