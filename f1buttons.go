package main

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
	"DiscordBot/commands/sports/f1"
)

func handleF1ResultsButton(b *bot.Bot, s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Extract session type from custom ID
	// Format: f1_results_{session_type}_{event_name}
	parts := strings.Split(i.MessageComponentData().CustomID, "_")
	if len(parts) < 4 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Invalid button interaction",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	sessionType := parts[2]
	// eventName := parts[3] // Not used in this implementation

	// Acknowledge the interaction
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		return
	}

	// Create a mock MessageCreate for the command handler
	m := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			ChannelID: i.ChannelID,
			Author:    i.Member.User,
		},
	}

	// Call the appropriate results function based on session type
	switch sessionType {
	case "free_practice_1", "fp1":
		f1.GetPracticeResults(b, s, m, "fp1", "")
	case "free_practice_2", "fp2":
		f1.GetPracticeResults(b, s, m, "fp2", "")
	case "free_practice_3", "fp3":
		f1.GetPracticeResults(b, s, m, "fp3", "")
	case "qualifying", "quali":
		f1.GetQualifyingResults(b, s, m, "")
	case "sprint_qualifying", "sprintquali":
		f1.GetSprintQualifyingResults(b, s, m, "")
	case "sprint":
		f1.GetSprintResults(b, s, m, "")
	case "race":
		f1.GetRaceResults(b, s, m, "")
	default:
		f1.GetRaceResults(b, s, m, "")
	}
}