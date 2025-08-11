package economy

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"math/rand"
	"strconv"
	"time"

	"DiscordBot/bot"
	"DiscordBot/commands"
	"github.com/bwmarrin/discordgo"
	"github.com/lib/pq"
)

func init() {
	commands.RegisterCommand("work", Work)
}

// getUserDailyInfo retrieves user's last daily timestamp and base daily hours
func getUserDailyInfo(b *bot.Bot, guildID, userID string) (sql.NullTime, int, error) {
	var lastDaily sql.NullTime
	var baseDailyHours int
	err := b.Db.QueryRow(`
		SELECT last_daily, base_daily_hours
		FROM guild_members
		WHERE guild_id = $1 AND user_id = $2
	`, guildID, userID).Scan(&lastDaily, &baseDailyHours)
	return lastDaily, baseDailyHours, err
}

// getRoleModifiers calculates the minimum hours and maximum multiplier based on user's roles
func getRoleModifiers(b *bot.Bot, guildID, userID string, s *discordgo.Session) (int, float64, error) {
	member, err := s.GuildMember(guildID, userID)
	if err != nil {
		return 0, 1.0, err
	}

	var roleIDs []int64
	for _, role := range member.Roles {
		roleID, err := strconv.ParseInt(role, 10, 64)
		if err != nil {
			continue
		}
		roleIDs = append(roleIDs, roleID)
	}

	if len(roleIDs) == 0 {
		return 0, 1.0, nil
	}

	rows, err := b.Db.Query(`
		SELECT min_hours, multiplier
		FROM role_daily_modifiers
		WHERE guild_id = $1 AND role_id = ANY($2)
	`, guildID, pq.Array(roleIDs))
	if err != nil {
		return 0, 1.0, err
	}
	defer rows.Close()

	minHours := 0
	maxMultiplier := 1.0
	for rows.Next() {
		var roleMinHours int
		var roleMultiplier float64
		if err := rows.Scan(&roleMinHours, &roleMultiplier); err != nil {
			return 0, 1.0, err
		}
		if minHours == 0 || roleMinHours < minHours {
			minHours = roleMinHours
		}
		if roleMultiplier > maxMultiplier {
			maxMultiplier = roleMultiplier
		}
	}

	return minHours, maxMultiplier, rows.Err()
}

// calculateWaitTime calculates the wait time based on last daily and role modifiers
func calculateWaitTime(lastDaily sql.NullTime, baseDailyHours, minHours int) time.Duration {
	if !lastDaily.Valid {
		return 0
	}

	waitHours := baseDailyHours
	if minHours > 0 && minHours < baseDailyHours {
		waitHours = minHours
	}

	waitDuration := time.Duration(waitHours) * time.Hour
	elapsed := time.Since(lastDaily.Time)

	if elapsed >= waitDuration {
		return 0
	}

	return waitDuration - elapsed
}

func Work(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if m.GuildID == "" {
		return
	}

	_, err := b.Db.Exec(`
		INSERT INTO users (user_id, username, avatar)
		VALUES ($1, $2, $3)
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

	_, err = b.Db.Exec(`
		INSERT INTO guild_members (guild_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (guild_id, user_id) DO NOTHING
	`, m.GuildID, m.Author.ID)
	if err != nil {
		log.Printf("Error ensuring guild member: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
		return
	}

	lastDaily, baseDailyHours, err := getUserDailyInfo(b, m.GuildID, m.Author.ID)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error getting user daily info: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
		return
	}

	minHours, maxMultiplier, err := getRoleModifiers(b, m.GuildID, m.Author.ID, s)
	if err != nil {
		log.Printf("Error getting role modifiers: %v", err)
		minHours = 0
		maxMultiplier = 1.0
	}

	waitTime := calculateWaitTime(lastDaily, baseDailyHours, minHours)

	if waitTime <= 0 {
		baseReward := rand.Intn(650-65) + 65
		reward := int(math.Round(float64(baseReward) * maxMultiplier))

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

		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("You received %d coins!", reward))
	} else {
		hours := int(waitTime.Hours())
		minutes := int(waitTime.Minutes()) % 60
		formattedWaitTime := fmt.Sprintf("%d hours %d minutes", hours, minutes)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("You have to wait %s before you can work again", formattedWaitTime))
	}
}
