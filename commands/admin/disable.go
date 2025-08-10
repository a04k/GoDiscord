package admin

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"DiscordBot/bot"
	"DiscordBot/commands"
	"github.com/bwmarrin/discordgo"
)

func init() {
	commands.RegisterCommand("disable", DisableCommand)
}

// DisableCommand allows server admins to disable specific commands or categories in their server
func DisableCommand(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// Check if the user is an admin (same as before)
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
		s.ChannelMessageSend(m.ChannelID, "Usage: `.disable command <command>` or `.disable category <category>`")
		return
	}

	disableType := strings.ToLower(args[1])
	name := strings.ToLower(args[2])

	if disableType != "command" && disableType != "category" {
		s.ChannelMessageSend(m.ChannelID, "Invalid type. Use 'command' or 'category'.")
		return
	}

	if disableType == "category" {
		// Check if it's a valid category
		if _, ok := commands.CommandCategories[strings.Title(name)]; !ok {
			s.ChannelMessageSend(m.ChannelID, "Invalid category.")
			return
		}
	}

	// Prevent disabling essential commands
	if disableType == "command" && (name == "help" || name == "enable" || name == "disable") {
		s.ChannelMessageSend(m.ChannelID, "You cannot disable this command.")
		return
	}

	_, err = b.Db.Exec(`
		INSERT INTO disabled_commands (guild_id, name, type) 
		VALUES ($1, $2, $3) 
		ON CONFLICT (guild_id, name) 
		DO UPDATE SET type = $3`,
		m.GuildID, name, disableType)

	if err != nil {
		log.Printf("Error disabling %s %s for guild %s: %v", disableType, name, m.GuildID, err)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error disabling %s.", disableType))
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Successfully disabled %s `%s`.", disableType, name))
}