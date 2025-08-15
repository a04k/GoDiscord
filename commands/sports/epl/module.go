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
				Name:        "epltable",
				Aliases:     []string{},
				Description: "Shows the current Premier League table",
				Usage:       ".epltable",
				Category:    "Sports",
			},
			{
				Name:        "nextmatch",
				Aliases:     []string{},
				Description: "Shows upcoming Premier League fixtures. Use with a club name to see their next match.",
				Usage:       ".nextmatch [club_name]",
				Category:    "Sports",
			},
		},
		SlashCommands: []commands.SlashCommandInfo{},
	}

	commands.RegisterModule(module)

	// Register command handlers
	commands.RegisterCommand("epltable", EPLTable)
	commands.RegisterCommand("nextmatch", NextMatch)
}
