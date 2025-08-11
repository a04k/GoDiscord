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
	commands.RegisterCommand("setadmin", SetAdmin, "sa")
}

func SetAdmin(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Usage: .sa <@user>")
		return
	}

	isOwner, err := b.IsOwner(m.GuildID, m.Author.ID)
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

	// With the new permission system, users with Administrator permission are automatically admins
	// This command now serves as a way to explain this to the user
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("In the updated system, users with Administrator permission are automatically considered admins. You can grant <@%s> the Administrator permission in Discord role settings.\n\nAlternatively, you can use custom roles for moderators with the `.sm` command.", recipient))
}
