package admin

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
	"DiscordBot/utils"
	"DiscordBot/commands"
)

func init() {
	commands.RegisterCommand("setadmin", SetAdmin, "sa")
}

func SetAdmin(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Usage: .sa <@user>")
		return
	}

	isOwner, err := utils.IsOwner(b.Db, m.GuildID, m.Author.ID)
	if err != nil {
		log.Printf("Error checking owner status: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
		return
	}
	if !isOwner {
		s.ChannelMessageSend(m.ChannelID, "You must be the server owner to use this command.")
		return
	}

	recipient := strings.TrimPrefix(strings.TrimSuffix(args[1], ">"), "<@")

	// Promote the user to admin
	_, err = b.Db.Exec("INSERT INTO users (guild_id, user_id, is_admin) VALUES ($1, $2, TRUE) ON CONFLICT (guild_id, user_id) DO UPDATE SET is_admin = TRUE", m.GuildID, recipient)
	if err != nil {
		log.Printf("Error promoting user: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Promoted <@%s> to admin.", recipient))
}