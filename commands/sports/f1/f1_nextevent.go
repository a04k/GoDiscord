package f1

import (
	"fmt"
	"log"
	"time"

	"DiscordBot/bot"

	"github.com/bwmarrin/discordgo"
)

func F1NextEvent(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
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

	events, err := FetchF1Events()
	if err != nil {
		log.Printf("Error fetching F1 events: %v", err)
		s.ChannelMessageSend(m.ChannelID, "Error fetching F1 schedule")
		return
	}

	now := time.Now().UTC()
	var nextEvent *Event

	// Find the next upcoming event
	for _, event := range events {
		if len(event.Sessions) == 0 {
			continue
		}

		// Get the first session of the event (typically Practice 1)
		firstSession := event.Sessions[0]
		eventStartTime, err := time.Parse(time.RFC3339, firstSession.Date)
		if err != nil {
			continue
		}

		// Find the first event that hasn't started yet
		if eventStartTime.After(now) {
			if nextEvent == nil || eventStartTime.Before(func() time.Time {
				firstSession := nextEvent.Sessions[0]
				t, _ := time.Parse(time.RFC3339, firstSession.Date)
				return t
			}()) {
				nextEvent = &event
			}
		}
	}

	if nextEvent == nil {
		s.ChannelMessageSend(m.ChannelID, "No upcoming F1 events found.")
		return
	}

	if len(nextEvent.Sessions) == 0 {
		s.ChannelMessageSend(m.ChannelID, "No sessions found for the next F1 event.")
		return
	}

	// Get the first session of the event (typically Practice 1)
	firstSession := nextEvent.Sessions[0]
	eventStartTime, err := time.Parse(time.RFC3339, firstSession.Date)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error parsing event time.")
		return
	}

	// Get the last session of the event (typically Race)
	lastSession := nextEvent.Sessions[len(nextEvent.Sessions)-1]
	eventEndTime, err := time.Parse(time.RFC3339, lastSession.Date)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error parsing event end time.")
		return
	}

	// Create an embed with the event information
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Next F1 Event: %s", nextEvent.Name),
		Description: fmt.Sprintf("%s, %s", nextEvent.Location, nextEvent.Circuit),
		Color:       0xFF0000, // Red color for F1
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Start Date",
				Value:  fmt.Sprintf("<t:%d:F> (<t:%d:R>)", eventStartTime.Unix(), eventStartTime.Unix()),
				Inline: true,
			},
			{
				Name:   "End Date",
				Value:  fmt.Sprintf("<t:%d:F> (<t:%d:R>)", eventEndTime.Unix(), eventEndTime.Unix()),
				Inline: true,
			},
			{
				Name:   "\u200B", // Zero-width space for spacing
				Value:  "\u200B",
				Inline: true,
			},
		},
	}

	// Add session information if available
	if len(nextEvent.Sessions) > 0 {
		sessionInfo := ""
		for _, session := range nextEvent.Sessions {
			sessionTime, err := time.Parse(time.RFC3339, session.Date)
			if err != nil {
				continue
			}

			sessionInfo += fmt.Sprintf("**%s**: <t:%d:F>\n", session.Name, sessionTime.Unix())
		}

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Sessions",
			Value:  sessionInfo,
			Inline: false,
		})
	}

	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}
