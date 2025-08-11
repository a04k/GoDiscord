package admin

import (
	"fmt"
	"log"
	"strings"

	"DiscordBot/bot"
	"DiscordBot/commands"
	"DiscordBot/utils"
	"github.com/bwmarrin/discordgo"
)

func init() {
	commands.RegisterCommand("disable", DisableCommand)
}

// DisableCommand allows server admins to disable specific commands or categories in their server
func DisableCommand(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// Check if the user has administrator permissions
	hasAdmin, err := utils.CheckAdminPermission(s, m.GuildID, m.Author.ID)
	if err != nil {
		log.Printf("Error checking admin status: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
		return
	}

	if !hasAdmin {
		s.ChannelMessageSend(m.ChannelID, "You must be a bot administrator to use this command.")
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