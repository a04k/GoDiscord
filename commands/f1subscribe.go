package commands

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
)

func F1Subscribe(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// Check if this is a DM
	channel, err := s.Channel(m.ChannelID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error checking channel type.")
		return
	}

	if channel.Type != discordgo.ChannelTypeDM {
		s.ChannelMessageSend(m.ChannelID, "This command can only be used in DMs. Please DM the bot to subscribe to F1 notifications.")
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