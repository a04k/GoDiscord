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

// SetDailyRole allows admins to set a role that modifies the daily cooldown
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

	if len(args) < 3 {
		s.ChannelMessageSend(m.ChannelID, "Usage: !setdailyrole <role_id_or_mention> <hours>\nExample: !setdailyrole @Premium 12")
		return
	}

	// Parse role ID
	roleID := args[1]
	if strings.HasPrefix(roleID, "<@&") && strings.HasSuffix(roleID, ">") {
		// It's a role mention, extract the ID
		roleID = roleID[3 : len(roleID)-1]
	}

	// Validate role ID
	_, err = strconv.ParseInt(roleID, 10, 64)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Invalid role ID or mention.")
		return
	}

	// Parse hours
	hours, err := strconv.Atoi(args[2])
	if err != nil || hours <= 0 {
		s.ChannelMessageSend(m.ChannelID, "Invalid hours. Please provide a positive number.")
		return
	}

	// Insert or update the role modifier
	_, err = b.Db.Exec(`
		INSERT INTO role_daily_modifiers (guild_id, role_id, min_hours)
		VALUES ($1, $2, $3)
		ON CONFLICT (guild_id, role_id) DO UPDATE SET min_hours = $3
	`, m.GuildID, roleID, hours)
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
		s.ChannelMessageSend(m.ChannelID, "Usage: !removedailyrole <role_id_or_mention>\nExample: !removedailyrole @Premium")
		return
	}

	// Parse role ID
	roleID := args[1]
	if strings.HasPrefix(roleID, "<@&") && strings.HasSuffix(roleID, ">") {
		// It's a role mention, extract the ID
		roleID = roleID[3 : len(roleID)-1]
	}

	// Validate role ID
	_, err = strconv.ParseInt(roleID, 10, 64)
	if err != nil {
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

	// Query all role modifiers for this guild
	rows, err := b.Db.Query(`
		SELECT role_id, min_hours
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

	// Build response message
	response := "Role Daily Cooldown Modifiers:\n"
	hasModifiers := false

	for rows.Next() {
		var roleID string
		var minHours int
		if err := rows.Scan(&roleID, &minHours); err != nil {
			log.Printf("Error scanning role daily modifier: %v", err)
			continue
		}
		response += fmt.Sprintf("<@&%s>: %d hours\n", roleID, minHours)
		hasModifiers = true
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating role daily modifiers: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred while retrieving role modifiers.")
		return
	}

	if !hasModifiers {
		response = "No role daily cooldown modifiers have been set."
	}

	s.ChannelMessageSend(m.ChannelID, response)
}