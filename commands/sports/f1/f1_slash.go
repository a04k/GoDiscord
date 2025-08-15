package f1

import (
	"log"

	"DiscordBot/bot"

	"github.com/bwmarrin/discordgo"
)

func F1SubscriptionToggle(b *bot.Bot, s *discordgo.Session, i *discordgo.InteractionCreate) {
	// For slash commands, we need to get the user ID differently
	userID := i.Member.User.ID
	if userID == "" {
		// If it's a DM, get the user ID from the user object
		userID = i.User.ID
	}

	var exists bool
	err := b.Db.QueryRow("SELECT EXISTS(SELECT 1 FROM f1_subscriptions WHERE user_id = $1)", userID).Scan(&exists)
	if err != nil {
		log.Printf("Error checking F1 subscription status for user %s: %v", userID, err)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "An error occurred while checking your subscription status.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
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

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: responseContent,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
