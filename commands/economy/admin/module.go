package economyadmin

import (
	"DiscordBot/commands"
)

func init() {
	module := &commands.ModuleInfo{
		Name:         "EconomyAdmin",
		Description:  "Administrative commands for economy management (add/take coins)",
		Version:      "1.0.0",
		Author:       "Bot Team",
		Category:     "Economy",
		Dependencies: []string{"Economy"},
		Config: map[string]interface{}{
			"require_confirmation": true,
			"log_transactions":     true,
		},
		Commands: []commands.CommandInfo{
			{
				Name:        "add",
				Aliases:     []string{},
				Description: "Adds coins to a user's balance",
				Usage:       ".add <user> <amount>",
				Category:    "Economy",
			},
			{
				Name:        "take",
				Aliases:     []string{},
				Description: "Takes coins from a user's balance",
				Usage:       ".take <user> <amount>",
				Category:    "Economy",
			},
		},
		SlashCommands: []commands.SlashCommandInfo{},
	}

	commands.RegisterModule(module)

	// Register command handlers
	commands.RegisterCommand("add", Add)
	commands.RegisterCommand("take", Take)
}
