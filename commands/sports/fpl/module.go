package fpl

import (
	"DiscordBot/commands"
)

func init() {
	module := &commands.ModuleInfo{
		Name:         "FPL",
		Description:  "Fantasy Premier League standings and league management",
		Version:      "1.0.0",
		Author:       "Bot Team",
		Category:     "Sports",
		Dependencies: []string{"EPL"},
		Config: map[string]interface{}{
			"default_league_id": nil,
		},
		Commands: []commands.CommandInfo{
			{
				Name:        "fplstandings",
				Aliases:     []string{"fpl"},
				Description: "Shows the standings for your guild's Fantasy Premier League",
				Usage:       ".fplstandings",
				Category:    "Sports",
			},
			{
				Name:        "setfplleague",
				Aliases:     []string{},
				Description: "Sets the Fantasy Premier League league ID for your guild (Admin only)",
				Usage:       ".setfplleague <league_id>",
				Category:    "Sports",
			},
		},
		SlashCommands: []commands.SlashCommandInfo{},
	}

	commands.RegisterModule(module)

	// Register command handlers
	commands.RegisterCommand("fplstandings", FPLStandings, "fpl")
	commands.RegisterCommand("setfplleague", SetFPLLeague)
}
