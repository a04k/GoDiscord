package admin

import (
	"fmt"
	"log"
	"strings"

	"DiscordBot/bot"
	"github.com/bwmarrin/discordgo"
	"DiscordBot/utils"
	"DiscordBot/commands"
)

func init() {
	commands.RegisterCommand("enable", EnableCommand)
}

// EnableCommand allows server admins to re-enable specific commands or categories in their server
func EnableCommand(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
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