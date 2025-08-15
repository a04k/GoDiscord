package currency

import (
	"DiscordBot/commands"
)

func init() {
	module := &commands.ModuleInfo{
		Name:         "EgyptCurrency",
		Description:  "USD to EGP currency conversion for Egypt",
		Version:      "1.0.0",
		Author:       "Bot Team",
		Category:     "Egypt",
		Dependencies: []string{},
		Config: map[string]interface{}{
			"default_currency": "EGP",
			"api_timeout":      10,
		},
		Commands: []commands.CommandInfo{
			{
				Name:        "usd",
				Aliases:     []string{},
				Description: "Converts USD to local currency (EGP)",
				Usage:       ".usd <amount>",
				Category:    "Egypt",
			},
		},
		SlashCommands: []commands.SlashCommandInfo{},
	}

	commands.RegisterModule(module)

	// Register command handlers
	commands.RegisterCommand("usd", USD)
}
