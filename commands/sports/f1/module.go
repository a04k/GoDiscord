package f1

import (
	"strings"

	"DiscordBot/bot"
	"DiscordBot/commands"

	"github.com/bwmarrin/discordgo"
)

// F1 handles all F1 subcommands
func F1(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Usage: .f1 <subcommand>\nAvailable subcommands: results, standings, wdc, wcc, quali, nextevent, next, sub")
		return
	}

	subcommand := strings.ToLower(args[1])
	switch subcommand {
	case "results", "f1results":
		F1Results(b, s, m, args[1:])
	case "standings", "f1standings":
		F1Standings(b, s, m, args[1:])
	case "wdc", "f1wdc":
		F1WDC(b, s, m, args[1:])
	case "wcc", "f1wcc":
		F1WCC(b, s, m, args[1:])
	case "quali", "qualiresults":
		QualiResults(b, s, m, args[1:])
	case "nextevent":
		F1NextEvent(b, s, m, args[1:])
	case "next", "nextf1session":
		NextF1Session(b, s, m, args[1:])
	case "sub", "f1sub":
		F1Subscribe(b, s, m, args[1:])
	default:
		s.ChannelMessageSend(m.ChannelID, "Unknown F1 subcommand. Available subcommands: results, standings, wdc, wcc, quali, nextevent, next, sub")
	}
}

func init() {
	module := &commands.ModuleInfo{
		Name:         "F1",
		Description:  "Formula 1 race results, standings, and live notifications",
		Version:      "1.2.0",
		Author:       "Bot Team",
		Category:     "Sports",
		Dependencies: []string{}, // No dependencies
		Config: map[string]interface{}{
			"notifications_enabled": true,
			"default_timezone":      "UTC",
			"notification_channels": []string{},
		},
		Commands: []commands.CommandInfo{
			{
				Name:        "f1",
				Aliases:     []string{},
				Description: "Formula 1 commands. Use .f1 <subcommand>",
				Usage:       ".f1 <subcommand>",
				Category:    "Sports",
			},
		},
		SlashCommands: []commands.SlashCommandInfo{
			{
				Name:        "f1",
				Description: "Subscribe/unsubscribe to F1 notifications",
				Options:     []*discordgo.ApplicationCommandOption{},
				Handler:     F1SubscriptionToggle,
			},
		},
	}

	commands.RegisterModule(module)

	// Register command handlers
	commands.RegisterCommand("f1", F1)
}
