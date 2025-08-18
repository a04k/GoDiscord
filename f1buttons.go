package main

import (
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
	"DiscordBot/commands/sports/f1"
)

func handleF1ResultsButton(b *bot.Bot, s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Format: f1_results_{session_type}_{round}
	parts := strings.Split(i.MessageComponentData().CustomID, "_")
	if len(parts) != 4 {
		return // Invalid ID format, ignore.
	}

	sessionType := parts[2]
	round, err := strconv.Atoi(parts[3])
	if err != nil {
		return // Invalid round number, ignore.
	}

	// Acknowledge the interaction immediately to prevent timeout.
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	// Create a mock MessageCreate for the command handlers.
	// This is a bit of a workaround to reuse the existing command structure.
	m := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			ChannelID: i.ChannelID,
			Author:    i.Member.User,
		},
	}

	// Call the appropriate results function based on session type.
	switch sessionType {
	case "qualifying":
		f1.GetQualifyingResults(b, s, m, round)
	case "sprint":
		f1.GetSprintResults(b, s, m, round)
	case "race":
		f1.GetRaceResults(b, s, m, round)
	default:
		// For unsupported sessions like practice, we can send a follow-up message.
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: "Results for this session type are not available.",
		})
	}
}
