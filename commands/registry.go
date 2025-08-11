package commands

import (
	"DiscordBot/bot"
	"github.com/bwmarrin/discordgo"
)

type CommandFunc func(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string)

var CommandMap = make(map[string]CommandFunc)
var CommandAliases = make(map[string]string)

func RegisterCommand(name string, handler CommandFunc, aliases ...string) {
	CommandMap[name] = handler
	for _, alias := range aliases {
		CommandAliases[alias] = name
	}
}
