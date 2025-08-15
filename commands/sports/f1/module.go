package f1

import (
	"DiscordBot/commands"

	"github.com/bwmarrin/discordgo"
)

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
				Description: "Shows current F1 information",
				Usage:       ".f1",
				Category:    "Sports",
			},
			{
				Name:        "f1results",
				Aliases:     []string{},
				Description: "Shows the latest F1 race results",
				Usage:       ".f1results",
				Category:    "Sports",
			},
			{
				Name:        "f1standings",
				Aliases:     []string{},
				Description: "Shows the current F1 championship standings",
				Usage:       ".f1standings",
				Category:    "Sports",
			},
			{
				Name:        "f1wdc",
				Aliases:     []string{},
				Description: "Shows the F1 Driver's Championship standings",
				Usage:       ".f1wdc",
				Category:    "Sports",
			},
			{
				Name:        "f1wcc",
				Aliases:     []string{},
				Description: "Shows the F1 Constructor's Championship standings",
				Usage:       ".f1wcc",
				Category:    "Sports",
			},
			{
				Name:        "qualiresults",
				Aliases:     []string{},
				Description: "Shows the latest F1 qualifying results",
				Usage:       ".qualiresults",
				Category:    "Sports",
			},
			{
				Name:        "nextf1session",
				Aliases:     []string{},
				Description: "Shows information about the next F1 session",
				Usage:       ".nextf1session",
				Category:    "Sports",
			},
			{
				Name:        "f1sub",
				Aliases:     []string{},
				Description: "Subscribe/unsubscribe to F1 notifications",
				Usage:       ".f1sub",
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
	commands.RegisterCommand("f1results", F1Results)
	commands.RegisterCommand("f1standings", F1Standings)
	commands.RegisterCommand("f1wdc", F1WDC)
	commands.RegisterCommand("f1wcc", F1WCC)
	commands.RegisterCommand("qualiresults", QualiResults)
	commands.RegisterCommand("nextf1session", NextF1Session)
	commands.RegisterCommand("f1sub", F1Subscribe)
}
