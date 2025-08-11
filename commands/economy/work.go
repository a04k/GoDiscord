package economy

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
	"DiscordBot/commands"
)

func init() {
	commands.RegisterCommand("work", Work)
}

// getUserDailyInfo retrieves user's last daily timestamp and base daily hours
func getUserDailyInfo(b *bot.Bot, guildID, userID string) (lastDaily sql.NullTime, baseDailyHours int, err error) {
	err = b.Db.QueryRow(`
		SELECT gm.last_daily, COALESCE(g.base_daily_hours, 24)
		FROM guild_members gm
		JOIN guilds g ON gm.guild_id = g.guild_id
		WHERE gm.guild_id = $1 AND gm.user_id = $2
	`, guildID, userID).Scan(&lastDaily, &baseDailyHours)
	return
}

// getMinHoursForRoles calculates the minimum hours based on user's roles
func getMinHoursForRoles(b *bot.Bot, guildID, userID string, s *discordgo.Session) (minHours int, err error) {
	// Get user's roles
	member, err := s.GuildMember(guildID, userID)
	if err != nil {
		return 0, err
	}

	// Convert role IDs to integers for the query
	var roleIDs []int64
	for _, role := range member.Roles {
		roleID, err := strconv.ParseInt(role, 10, 64)
		if err != nil {
			continue // Skip invalid role IDs
		}
		roleIDs = append(roleIDs, roleID)
	}

	// If no valid roles, return 0
	if len(roleIDs) == 0 {
		return 0, nil
	}

	// Get all role modifiers for this guild
	rows, err := b.Db.Query(`
		SELECT min_hours
		FROM role_daily_modifiers
		WHERE guild_id = $1 AND role_id = ANY($2)
	`, guildID, roleIDs)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	minHours = 0 // Default to no reduction
	for rows.Next() {
		var roleMinHours int
		if err := rows.Scan(&roleMinHours); err != nil {
			return 0, err
		}
		// Use the minimum hours from all applicable roles
		if minHours == 0 || roleMinHours < minHours {
			minHours = roleMinHours
		}
	}

	return minHours, rows.Err()
}

// calculateWaitTime calculates the wait time based on last daily and role modifiers
func calculateWaitTime(lastDaily sql.NullTime, baseDailyHours, minHours int) time.Duration {
	if !lastDaily.Valid {
		return 0 // No wait time if never worked before
	}

	waitHours := baseDailyHours
	if minHours > 0 && minHours < baseDailyHours {
		waitHours = minHours
	}

	waitDuration := time.Duration(waitHours) * time.Hour
	elapsed := time.Since(lastDaily.Time)

	if elapsed >= waitDuration {
		return 0 // No wait time
	}

	return waitDuration - elapsed
}

func Work(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// Ensure command is used in a guild
	if m.GuildID == "" {
		return // Don't respond to DMs
	}

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

	// Ensure user exists in guild_members table
	_, err = b.Db.Exec(`
		INSERT INTO guild_members (guild_id, user_id, joined_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (guild_id, user_id) DO NOTHING
	`, m.GuildID, m.Author.ID)
	if err != nil {
		log.Printf("Error ensuring guild member: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
		return
	}

	// Get user's daily info
	lastDaily, baseDailyHours, err := getUserDailyInfo(b, m.GuildID, m.Author.ID)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error getting user daily info: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
		return
	}

	// Get minimum hours from roles
	minHours, err := getMinHoursForRoles(b, m.GuildID, m.Author.ID, s)
	if err != nil {
		log.Printf("Error getting role modifiers: %v", err)
		// Continue with default behavior if role check fails
		minHours = 0
	}

	// Calculate wait time
	waitTime := calculateWaitTime(lastDaily, baseDailyHours, minHours)

	if waitTime <= 0 {
		// User can work now
		reward := rand.Intn(650-65) + 65

		// Update user's balance and last daily
		_, err := b.Db.Exec(`
			UPDATE guild_members 
			SET balance = balance + $1, last_daily = NOW() 
			WHERE guild_id = $2 AND user_id = $3
		`, reward, m.GuildID, m.Author.ID)
		if err != nil {
			log.Printf("Error updating user balance: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
			return
		}

		// Log transaction
		_, err = b.Db.Exec(`
			INSERT INTO transactions (guild_id, user_id, amount, balance_before, balance_after, reason)
			VALUES (
				$1, $2, $3,
				(SELECT balance - $3 FROM guild_members WHERE guild_id = $1 AND user_id = $2),
				(SELECT balance FROM guild_members WHERE guild_id = $1 AND user_id = $2),
				'work'
			)
		`, m.GuildID, m.Author.ID, reward)
		if err != nil {
			log.Printf("Error logging transaction: %v", err)
			// Don't fail the command if transaction logging fails
		}

		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("You received %d coins!", reward))
	} else {
		// User needs to wait
		hours := int(waitTime.Hours())
		minutes := int(waitTime.Minutes()) % 60
		formattedWaitTime := fmt.Sprintf("%d hours %d minutes", hours, minutes)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("You have to wait %s before you can work again", formattedWaitTime))
	}
}