package admin

import (
	"fmt"
	"log"
	"strings"

	"DiscordBot/bot"
	"DiscordBot/commands"
	"github.com/bwmarrin/discordgo"
)

func init() {
	commands.RegisterCommand("setmod", SetMod, "sm")
}

func SetMod(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Usage: .sm <@user>")
		return
	}

	isAdmin, err := b.IsAdmin(m.GuildID, m.Author.ID)
	if err != nil {
		log.Printf("Error checking admin status: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
		return
	}
	if !isAdmin {
		s.ChannelMessageSend(m.ChannelID, "You must be an admin or owner to use this command.")
		return
	}

	recipient := strings.TrimPrefix(strings.TrimSuffix(args[1], ">"), "<@")

	// Add the user as a moderator in our database
	_, err = b.Db.Exec("INSERT INTO permissions (guild_id, user_id, role) VALUES ($1, $2, 'moderator') ON CONFLICT (guild_id, user_id) DO UPDATE SET role = 'moderator'", m.GuildID, recipient)
	if err != nil {
		log.Printf("Error promoting user: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Promoted <@%s> to custom moderator. Note: Users with Manage Messages permission are also automatically considered moderators.", recipient))
}
