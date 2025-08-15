package epl

import (
	"DiscordBot/commands"
)

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
				Name:        "epl table",
				Aliases:     []string{"epltable"},
				Description: "Shows the current Premier League table",
				Usage:       ".epl table",
				Category:    "Sports",
			},
			{
				Name:        "epl next",
				Aliases:     []string{"nextmatch"},
				Description: "Shows upcoming Premier League fixtures. Use with a club name to see their next match.",
				Usage:       ".epl next [club_name]",
				Category:    "Sports",
			},
		},
		SlashCommands: []commands.SlashCommandInfo{},
	}

	commands.RegisterModule(module)

	// Register command handlers
	commands.RegisterCommand("epl table", EPLTable, "epltable")
	commands.RegisterCommand("epl next", NextMatch, "nextmatch")
}
