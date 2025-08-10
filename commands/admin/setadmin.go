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

	isAdmin, err := utils.IsAdmin(b.Db, m.Author.ID)
	if err != nil {
		log.Printf("Error checking admin status: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
		return
	}
	if !isAdmin {
		s.ChannelMessageSend(m.ChannelID, "You are not authorized to use this command.")
		return
	}

	recipient := strings.TrimPrefix(strings.TrimSuffix(args[1], ">"), "<@")

	// Promote the user to admin
	_, err = b.Db.Exec("UPDATE users SET is_admin = TRUE WHERE user_id = $1", recipient)
	if err != nil {
		log.Printf("Error promoting user: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Promoted <@%s> to admin.", recipient))
}