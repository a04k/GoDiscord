package admin

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"DiscordBot/bot"
	"github.com/bwmarrin/discordgo"
	"DiscordBot/commands"
)

func init() {
	commands.RegisterCommand("enable", EnableCommand)
}

// EnableCommand allows server admins to re-enable specific commands or categories in their server
func EnableCommand(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// Check if the user is an admin (same as before)
	var isAdmin bool
	err := b.Db.QueryRow("SELECT is_admin FROM users WHERE guild_id = $1 AND user_id = $2", m.GuildID, m.Author.ID).Scan(&isAdmin)
	if err != nil {
		if err == sql.ErrNoRows {
			isAdmin = false
		} else {
			log.Printf("Error checking admin status for user %s: %v", m.Author.ID, err)
			s.ChannelMessageSend(m.ChannelID, "Error checking admin status.")
			return
		}
	}

	member, err := s.State.Member(m.GuildID, m.Author.ID)
	if err != nil {
		log.Printf("Error getting member %s in guild %s: %v", m.Author.ID, m.GuildID, err)
		s.ChannelMessageSend(m.ChannelID, "Error getting member information.")
		return
	}

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

	if !isAdmin && !isGuildAdmin {
		s.ChannelMessageSend(m.ChannelID, "You must be an administrator to use this command.")
		return
	}

	if len(args) < 3 {
		s.ChannelMessageSend(m.ChannelID, "Usage: `.enable command <command>` or `.enable category <category>`")
		return
	}

	enableType := strings.ToLower(args[1])
	name := strings.ToLower(args[2])

	if enableType != "command" && enableType != "category" {
		s.ChannelMessageSend(m.ChannelID, "Invalid type. Use 'command' or 'category'.")
		return
	}

	result, err := b.Db.Exec(`
		DELETE FROM disabled_commands 
		WHERE guild_id = $1 AND name = $2 AND type = $3`,
		m.GuildID, name, enableType)

	if err != nil {
		log.Printf("Error enabling %s %s for guild %s: %v", enableType, name, m.GuildID, err)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error enabling %s.", enableType))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s `%s` was not disabled.", strings.Title(enableType), name))
	} else {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Successfully enabled %s `%s`.", enableType, name))
	}
}