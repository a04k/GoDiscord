package admin

import (
	"database/sql"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
)

// DisableCommand allows server admins to disable specific commands in their server
func DisableCommand(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// Check if the user is an admin
	var isAdmin bool
	err := b.Db.QueryRow("SELECT is_admin FROM users WHERE user_id = $1", m.Author.ID).Scan(&isAdmin)
	if err != nil {
		if err == sql.ErrNoRows {
			isAdmin = false
		} else {
			log.Printf("Error checking admin status for user %s: %v", m.Author.ID, err)
			s.ChannelMessageSend(m.ChannelID, "Error checking admin status.")
			return
		}
	}

	// Also check if the user has the \"Administrator\" permission in the guild
	guild, err := s.State.Guild(m.GuildID)
	if err != nil {
		log.Printf("Error getting guild %s: %v", m.GuildID, err)
		s.ChannelMessageSend(m.ChannelID, "Error getting guild information.")
		return
	}

	// Find the user in the guild members
	member, err := s.State.Member(m.GuildID, m.Author.ID)
	if err != nil {
		log.Printf("Error getting member %s in guild %s: %v", m.Author.ID, m.GuildID, err)
		s.ChannelMessageSend(m.ChannelID, "Error getting member information.")
		return
	}

	// Check if the user has administrator permissions
	isGuildAdmin := false
	for _, roleID := range member.Roles {
		role, err := s.State.Role(m.GuildID, roleID)
		if err != nil {
			continue
		}
		if role.Permissions&discordgo.PermissionAdministrator != 0 {
			isGuildAdmin = true
			break
		}
	}

	// If the user is not an admin in our system and doesn't have administrator permissions in the guild, deny access
	if !isAdmin && !isGuildAdmin {
		s.ChannelMessageSend(m.ChannelID, \"You must be an administrator to use this command.\")
		return
	}

	// Check if a command name was provided
	if len(args) < 2 {
		s.ChannelMessageSend(m.ChannelID, \"Please specify a command to disable.\\nUsage: `disablecommand <command>`\")
		return
	}

	commandName := strings.ToLower(args[1])

	// Insert or update the disabled command in the database
	_, err = b.Db.Exec(`
		INSERT INTO disabled_commands (guild_id, command_name) 
		VALUES ($1, $2) 
		ON CONFLICT (guild_id, command_name) 
		DO NOTHING`,
		m.GuildID, commandName)
	
	if err != nil {
		log.Printf(\"Error disabling command %s for guild %s: %v\", commandName, m.GuildID, err)
		s.ChannelMessageSend(m.ChannelID, \"Error disabling command.\")
		return
	}

	s.ChannelMessageSend(m.ChannelID, \"Command `\"+commandName+\"` has been disabled in this server.\")
}

// EnableCommand allows server admins to re-enable specific commands in their server
func EnableCommand(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// Check if the user is an admin
	var isAdmin bool
	err := b.Db.QueryRow(\"SELECT is_admin FROM users WHERE user_id = $1\", m.Author.ID).Scan(&isAdmin)
	if err != nil {
		if err == sql.ErrNoRows {
			isAdmin = false
		} else {
			log.Printf(\"Error checking admin status for user %s: %v\", m.Author.ID, err)
			s.ChannelMessageSend(m.ChannelID, \"Error checking admin status.\")
			return
		}
	}

	// Also check if the user has the \"Administrator\" permission in the guild
	guild, err := s.State.Guild(m.GuildID)
	if err != nil {
		log.Printf(\"Error getting guild %s: %v\", m.GuildID, err)
		s.ChannelMessageSend(m.ChannelID, \"Error getting guild information.\")
		return
	}

	// Find the user in the guild members
	member, err := s.State.Member(m.GuildID, m.Author.ID)
	if err != nil {
		log.Printf(\"Error getting member %s in guild %s: %v\", m.Author.ID, m.GuildID, err)
		s.ChannelMessageSend(m.ChannelID, \"Error getting member information.\")
		return
	}

	// Check if the user has administrator permissions
	isGuildAdmin := false
	for _, roleID := range member.Roles {
		role, err := s.State.Role(m.GuildID, roleID)
		if err != nil {
			continue
		}
		if role.Permissions&discordgo.PermissionAdministrator != 0 {
			isGuildAdmin = true
			break
		}
	}

	// If the user is not an admin in our system and doesn't have administrator permissions in the guild, deny access
	if !isAdmin && !isGuildAdmin {
		s.ChannelMessageSend(m.ChannelID, \"You must be an administrator to use this command.\")
		return
	}

	// Check if a command name was provided
	if len(args) < 2 {
		s.ChannelMessageSend(m.ChannelID, \"Please specify a command to enable.\\nUsage: `enablecommand <command>`\")
		return
	}

	commandName := strings.ToLower(args[1])

	// Delete the disabled command from the database
	result, err := b.Db.Exec(`
		DELETE FROM disabled_commands 
		WHERE guild_id = $1 AND command_name = $2`,
		m.GuildID, commandName)
	
	if err != nil {
		log.Printf(\"Error enabling command %s for guild %s: %v\", commandName, m.GuildID, err)
		s.ChannelMessageSend(m.ChannelID, \"Error enabling command.\")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		s.ChannelMessageSend(m.ChannelID, \"Command `\"+commandName+\"` was not disabled in this server.\")
	} else {
		s.ChannelMessageSend(m.ChannelID, \"Command `\"+commandName+\"` has been enabled in this server.\")
	}
}