package commands

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
)

func init() {
	RegisterCommand("remindme", RemindMe)
}

// parseDuration parses a duration string like "2h30m" into time.Duration
func parseDuration(durationStr string) (time.Duration, error) {
	// Regular expression to match duration components
	re := regexp.MustCompile(`(\d+)([a-zA-Z]+)`)
	matches := re.FindAllStringSubmatch(durationStr, -1)

	if len(matches) == 0 {
		return 0, fmt.Errorf("invalid duration format")
	}

	var totalDuration time.Duration
	for _, match := range matches {
		value, err := strconv.Atoi(match[1])
		if err != nil {
			return 0, fmt.Errorf("invalid duration value: %v", err)
		}

		unit := strings.ToLower(match[2])
		var duration time.Duration
		switch unit {
		case "s", "sec", "secs":
			duration = time.Duration(value) * time.Second
		case "m", "min", "mins":
			duration = time.Duration(value) * time.Minute
		case "h", "hr", "hrs":
			duration = time.Duration(value) * time.Hour
		case "d", "day", "days":
			duration = time.Duration(value) * 24 * time.Hour
		default:
			return 0, fmt.Errorf("unknown time unit: %s", unit)
		}

		totalDuration += duration
	}

	return totalDuration, nil
}

func RemindMe(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Usage: !remindme <duration> <message>\nExample: !remindme 2h30m Take out the trash")
		return
	}

	// Parse duration
	durationStr := args[0]
	duration, err := parseDuration(durationStr)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Invalid duration format: %v\nExample: 2h30m, 1d, 30m", err))
		return
	}

	// Validate duration
	if duration <= 0 {
		s.ChannelMessageSend(m.ChannelID, "Duration must be positive")
		return
	}

	// Get reminder message
	message := strings.Join(args[1:], " ")
	if message == "" {
		s.ChannelMessageSend(m.ChannelID, "Please provide a reminder message")
		return
	}

	// Calculate remind time
	remindAt := time.Now().Add(duration)

	// Insert reminder into database
	_, err = b.Db.Exec(`
		INSERT INTO reminders (user_id, guild_id, message, remind_at, source)
		VALUES ($1, $2, $3, $4, 'remindme')
	`, m.Author.ID, m.GuildID, message, remindAt)
	if err != nil {
		log.Printf("Error inserting reminder: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred while setting your reminder. Please try again.")
		return
	}

	// Format duration for response
	var durationParts []string
	days := int(duration.Hours()) / 24
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60

	if days > 0 {
		durationParts = append(durationParts, fmt.Sprintf("%d day(s)", days))
	}
	if hours > 0 {
		durationParts = append(durationParts, fmt.Sprintf("%d hour(s)", hours))
	}
	if minutes > 0 {
		durationParts = append(durationParts, fmt.Sprintf("%d minute(s)", minutes))
	}
	if seconds > 0 && len(durationParts) == 0 { // Only show seconds if no other units
		durationParts = append(durationParts, fmt.Sprintf("%d second(s)", seconds))
	}

	formattedDuration := strings.Join(durationParts, ", ")
	if formattedDuration == "" {
		formattedDuration = "0 seconds"
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Reminder set for %s from now: %s", formattedDuration, message))
}