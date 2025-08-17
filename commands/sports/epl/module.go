package epl

import (
	"strings"

	"DiscordBot/bot"
	"DiscordBot/commands"

	"github.com/bwmarrin/discordgo"
)

// EPL handles all EPL subcommands
func EPL(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Usage: .epl <subcommand>\nAvailable subcommands: table, next")
		return
	}

	subcommand := strings.ToLower(args[1])
	switch subcommand {
	case "table", "epltable":
		EPLTable(b, s, m, args[1:])
	case "next", "nextmatch":
		NextMatch(b, s, m, args[1:])
	default:
		s.ChannelMessageSend(m.ChannelID, "Unknown EPL subcommand. Available subcommands: table, next")
	}
}

func init() {
	module := &commands.ModuleInfo{
		Name:         "EPL",
		Description:  "English Premier League table and fixture information",
		Version:      "1.0.0",
		Author:       "Bot Team",
		Category:     "Sports",
		Dependencies: []string{},
		Config: map[string]interface{}{
			"default_season": "current",
		},
		Commands: []commands.CommandInfo{
			{
				Name:        "epl",
				Aliases:     []string{},
				Description: "English Premier League commands. Use .epl <subcommand>",
				Usage:       ".epl <subcommand>",
				Category:    "Sports",
			},
		},
		SlashCommands: []commands.SlashCommandInfo{},
	}

	commands.RegisterModule(module)

	// Register command handlers
	commands.RegisterCommand("epl", EPL)
}
