package economy

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
	"DiscordBot/commands"
)

func init() {
	commands.RegisterCommand("setdailyrole", SetDailyRole)
	commands.RegisterCommand("removedailyrole", RemoveDailyRole)
	commands.RegisterCommand("listdailyroles", ListDailyRoles)
}

// SetDailyRole allows admins to set a role that modifies the daily cooldown and multiplier
func SetDailyRole(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// Check if user is admin
	isAdmin, err := b.IsAdmin(m.GuildID, m.Author.ID)
	if err != nil {
		log.Printf("Error checking admin status: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
		return
	}
	if !isAdmin {
		s.ChannelMessageSend(m.ChannelID, "You must be an admin to use this command.")
		return
	}

	if len(args) < 4 {
		s.ChannelMessageSend(m.ChannelID, "Usage: .setdailyrole <role> <hours> <multiplier>\nExample: .setdailyrole @Premium 12 1.5")
		return
	}

	// Parse role ID
	roleID := args[1]
	if strings.HasPrefix(roleID, "<@&") && strings.HasSuffix(roleID, ">") {
		roleID = roleID[3 : len(roleID)-1]
	}

	// Validate role ID
	if _, err := strconv.ParseInt(roleID, 10, 64); err != nil {
		s.ChannelMessageSend(m.ChannelID, "Invalid role ID or mention.")
		return
	}

	// Parse hours
	hours, err := strconv.Atoi(args[2])
	if err != nil || hours <= 0 {
		s.ChannelMessageSend(m.ChannelID, "Invalid hours. Please provide a positive number.")
		return
	}

	// Parse multiplier
	multiplier, err := strconv.ParseFloat(args[3], 64)
	if err != nil || multiplier <= 0 {
		s.ChannelMessageSend(m.ChannelID, "Invalid multiplier. Please provide a positive number.")
		return
	}

	// Insert or update the role modifier
	_, err = b.Db.Exec(`
		INSERT INTO role_daily_modifiers (guild_id, role_id, min_hours, multiplier)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (guild_id, role_id) DO UPDATE SET min_hours = $3, multiplier = $4
	`, m.GuildID, roleID, hours, multiplier)
	if err != nil {
		log.Printf("Error setting role daily modifier: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred while setting the role modifier.")
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Role daily modifier set successfully.")
}

// RemoveDailyRole allows admins to remove a role's daily cooldown modifier
func RemoveDailyRole(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// Check if user is admin
	isAdmin, err := b.IsAdmin(m.GuildID, m.Author.ID)
	if err != nil {
		log.Printf("Error checking admin status: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
		return
	}
	if !isAdmin {
		s.ChannelMessageSend(m.ChannelID, "You must be an admin to use this command.")
		return
	}

	if len(args) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Usage: .removedailyrole <role_id_or_mention>\nExample: .removedailyrole @Premium")
		return
	}

	// Parse role ID
	roleID := args[1]
	if strings.HasPrefix(roleID, "<@&") && strings.HasSuffix(roleID, ">") {
		roleID = roleID[3 : len(roleID)-1]
	}

	// Validate role ID
	if _, err := strconv.ParseInt(roleID, 10, 64); err != nil {
		s.ChannelMessageSend(m.ChannelID, "Invalid role ID or mention.")
		return
	}

	// Delete the role modifier
	result, err := b.Db.Exec(`
		DELETE FROM role_daily_modifiers
		WHERE guild_id = $1 AND role_id = $2
	`, m.GuildID, roleID)
	if err != nil {
		log.Printf("Error removing role daily modifier: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred while removing the role modifier.")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		s.ChannelMessageSend(m.ChannelID, "No daily modifier found for that role.")
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Role daily modifier removed successfully.")
}

// ListDailyRoles shows all roles with daily cooldown modifiers
func ListDailyRoles(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// Query all role modifiers for this guild
	rows, err := b.Db.Query(`
		SELECT role_id, min_hours, multiplier
		FROM role_daily_modifiers
		WHERE guild_id = $1
		ORDER BY min_hours
	`, m.GuildID)
	if err != nil {
		log.Printf("Error querying role daily modifiers: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred while retrieving role modifiers.")
		return
	}
	defer rows.Close()

	var fields []*discordgo.MessageEmbedField
	for rows.Next() {
		var roleID string
		var minHours int
		var multiplier float64
		if err := rows.Scan(&roleID, &minHours, &multiplier); err != nil {
			log.Printf("Error scanning role daily modifier: %v", err)
			continue
		}
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("<@&%s>", roleID),
			Value:  fmt.Sprintf("Cooldown: %d hours\nMultiplier: %.2fx", minHours, multiplier),
			Inline: true,
		})
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating role daily modifiers: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred while retrieving role modifiers.")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Role Daily Modifiers",
		Description: "Roles with special daily bonuses in this server.",
		Fields:      fields,
		Color:       0x00ff00, // Green
	}

	if len(fields) == 0 {
		embed.Description = "No role daily cooldown modifiers have been set."
	}

	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}
