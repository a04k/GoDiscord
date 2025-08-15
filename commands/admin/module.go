package admin

import (
	"DiscordBot/commands"
)

func init() {
	module := &commands.ModuleInfo{
		Name:         "Admin",
		Description:  "Administrative commands for bot management",
		Version:      "1.0.0",
		Author:       "Bot Team",
		Category:     "Admin",
		Dependencies: []string{},
		Config: map[string]interface{}{
			"require_confirmation": true,
		},
		Commands: []commands.CommandInfo{
			{
				Name:        "disable",
				Aliases:     []string{},
				Description: "Disables a command or category",
				Usage:       ".disable <command|category> <name>",
				Category:    "Admin",
			},
			{
				Name:        "enable",
				Aliases:     []string{},
				Description: "Enables a command or category",
				Usage:       ".enable <command|category> <name>",
				Category:    "Admin",
			},
		},
		SlashCommands: []commands.SlashCommandInfo{},
	}

	commands.RegisterModule(module)

	// Register command handlers
	commands.RegisterCommand("disable", DisableCommand)
	commands.RegisterCommand("enable", EnableCommand)
}
