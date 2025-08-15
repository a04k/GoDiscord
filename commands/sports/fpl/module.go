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
				Name:        "fpl standings",
				Aliases:     []string{"fplstandings", "fpl"},
				Description: "Shows the standings for your guild's Fantasy Premier League",
				Usage:       ".fpl standings",
				Category:    "Sports",
			},
			{
				Name:        "fpl setleague",
				Aliases:     []string{"setfplleague"},
				Description: "Sets the Fantasy Premier League league ID for your guild (Admin only)",
				Usage:       ".fpl setleague <league_id>",
				Category:    "Sports",
			},
		},
		SlashCommands: []commands.SlashCommandInfo{},
	}

	commands.RegisterModule(module)

	// Register command handlers
	commands.RegisterCommand("fpl standings", FPLStandings, "fplstandings", "fpl")
	commands.RegisterCommand("fpl setleague", SetFPLLeague, "setfplleague")
}
