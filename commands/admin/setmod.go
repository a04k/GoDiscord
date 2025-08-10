package admin

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
	"DiscordBot/commands"
	"DiscordBot/utils"
)

func init() {
	commands.RegisterCommand("setmod", SetMod, "sm")
}

func SetMod(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Usage: .sm <@user>")
		return
	}

	isAdmin, err := utils.IsAdmin(b.Db, m.GuildID, m.Author.ID)
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

	// Promote the user to mod
	_, err = b.Db.Exec("INSERT INTO users (guild_id, user_id, is_mod) VALUES ($1, $2, TRUE) ON CONFLICT (guild_id, user_id) DO UPDATE SET is_mod = TRUE", m.GuildID, recipient)
	if err != nil {
		log.Printf("Error promoting user: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Promoted <@%s> to mod.", recipient))
}