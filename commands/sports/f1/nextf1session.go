package f1

import (
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
)

func NextF1Session(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	events, err := FetchF1Events()
	if err != nil {
		log.Printf("Error fetching F1 events: %v", err)
		s.ChannelMessageSend(m.ChannelID, "Error fetching F1 schedule")
		return
	}

	now := time.Now().UTC()
	var nextSession *Session
	var parentEvent *Event

	// Find the next upcoming session
	for _, event := range events {
		for _, session := range event.Sessions {
			sessionTime, err := time.Parse(time.RFC3339, session.Date)
			if err != nil {
				continue
			}

			// Find the first session that hasn't started yet
			if sessionTime.After(now) {
				if nextSession == nil || sessionTime.Before(func() time.Time {
					t, _ := time.Parse(time.RFC3339, nextSession.Date)
					return t
				}()) {
					nextSession = &session
					parentEvent = &event
				}
			}
		}
	}

	if nextSession == nil || parentEvent == nil {
		s.ChannelMessageSend(m.ChannelID, "No upcoming F1 sessions found.")
		return
	}

	sessionTime, err := time.Parse(time.RFC3339, nextSession.Date)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error parsing session time.")
		return
	}

	// Create an embed with the session information
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Next F1 Session: %s", nextSession.Name),
		Description: fmt.Sprintf("**%s** at %s", parentEvent.Name, parentEvent.Location),
		Color:       0xFF0000, // Red color for F1
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Date & Time",
				Value:  fmt.Sprintf("<t:%d:F> (<t:%d:R>)", sessionTime.Unix(), sessionTime.Unix()),
				Inline: false,
			},
		},
	}

	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}
