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
				Name:        "f1 results",
				Aliases:     []string{"f1results"},
				Description: "Shows the latest F1 race results",
				Usage:       ".f1 results",
				Category:    "Sports",
			},
			{
				Name:        "f1 standings",
				Aliases:     []string{"f1standings"},
				Description: "Shows the current F1 championship standings",
				Usage:       ".f1 standings",
				Category:    "Sports",
			},
			{
				Name:        "f1 wdc",
				Aliases:     []string{"f1wdc"},
				Description: "Shows the F1 Driver's Championship standings",
				Usage:       ".f1 wdc",
				Category:    "Sports",
			},
			{
				Name:        "f1 wcc",
				Aliases:     []string{"f1wcc"},
				Description: "Shows the F1 Constructor's Championship standings",
				Usage:       ".f1 wcc",
				Category:    "Sports",
			},
			{
				Name:        "f1 quali",
				Aliases:     []string{"qualiresults"},
				Description: "Shows the latest F1 qualifying results",
				Usage:       ".f1 quali",
				Category:    "Sports",
			},
			{
				Name:        "f1 next",
				Aliases:     []string{"nextf1session"},
				Description: "Shows information about the next F1 session",
				Usage:       ".f1 next",
				Category:    "Sports",
			},
			{
				Name:        "f1 sub",
				Aliases:     []string{"f1sub"},
				Description: "Subscribe/unsubscribe to F1 notifications",
				Usage:       ".f1 sub",
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
	commands.RegisterCommand("f1 results", F1Results, "f1results")
	commands.RegisterCommand("f1 standings", F1Standings, "f1standings")
	commands.RegisterCommand("f1 wdc", F1WDC, "f1wdc")
	commands.RegisterCommand("f1 wcc", F1WCC, "f1wcc")
	commands.RegisterCommand("f1 quali", QualiResults, "qualiresults")
	commands.RegisterCommand("f1 next", NextF1Session, "nextf1session")
	commands.RegisterCommand("f1 sub", F1Subscribe, "f1sub")
}
