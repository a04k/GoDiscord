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

// parseDuration converts "1d 2h 30m" into time.Duration
func parseDuration(input string) (time.Duration, error) {
	// Case-insensitive regex, allows spaces between units
	re := regexp.MustCompile(`(?i)(\d+)\s*(s|sec|secs|m|min|mins|h|hr|hrs|d|day|days)`)
	matches := re.FindAllStringSubmatch(input, -1)
	if len(matches) == 0 {
		return 0, fmt.Errorf("invalid duration format")
	}

	var total time.Duration
	for _, match := range matches {
		value, _ := strconv.Atoi(match[1])
		switch strings.ToLower(match[2]) {
		case "s", "sec", "secs":
			total += time.Second * time.Duration(value)
		case "m", "min", "mins":
			total += time.Minute * time.Duration(value)
		case "h", "hr", "hrs":
			total += time.Hour * time.Duration(value)
		case "d", "day", "days":
			total += time.Hour * 24 * time.Duration(value)
		default:
			return 0, fmt.Errorf("unknown time unit: %s", match[2])
		}
	}

	return total, nil
}

// formatDuration pretty-prints a duration like "1 day, 2 hours, 5 minutes"
func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	parts := []string{}
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%d day(s)", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%d hour(s)", hours))
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%d minute(s)", minutes))
	}
	if seconds > 0 && len(parts) == 0 { // only show seconds if no bigger unit
		parts = append(parts, fmt.Sprintf("%d second(s)", seconds))
	}

	if len(parts) == 0 {
		return "0 seconds"
	}
	return strings.Join(parts, ", ")
}

func RemindMe(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 2 {
		s.ChannelMessageSend(m.ChannelID,
			"Usage: `.remindme <duration> <message>`\nExample: `.remindme 2h30m Take out the trash`")
		return
	}

	// First arg is the duration
	duration, err := parseDuration(args[0])
	if err != nil || duration <= 0 {
		s.ChannelMessageSend(m.ChannelID,
			"Invalid duration. Example: `10m`, `2h30m`, `1d 2h`")
		return
	}

	// Remaining args are the reminder message
	message := strings.TrimSpace(strings.Join(args[1:], " "))
	if message == "" {
		s.ChannelMessageSend(m.ChannelID, "Please provide a reminder message.")
		return
	}

	remindAt := time.Now().UTC().Add(duration)

	// Insert reminder into DB
	_, err = b.Db.Exec(`
		INSERT INTO reminders (user_id, guild_id, message, remind_at, source)
		VALUES ($1, $2, $3, $4, 'remindme')
	`, m.Author.ID, m.GuildID, message, remindAt)
	if err != nil {
		log.Printf("Error inserting reminder: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred while setting your reminder.")
		return
	}

	// Confirm to user
	s.ChannelMessageSend(m.ChannelID,
		fmt.Sprintf("⏰ Reminder set for **%s** from now: %s",
			formatDuration(duration), message))
}
