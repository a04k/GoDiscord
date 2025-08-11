package admin

import (
	"DiscordBot/bot"
	"DiscordBot/commands"
	"github.com/bwmarrin/discordgo"
)

func init() {
	commands.RegisterCommand("setmod", SetMod, "sm")
}

func SetMod(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	s.ChannelMessageSend(m.ChannelID, "This command is deprecated. Users with Manage Messages permission are automatically considered moderators.")
}
