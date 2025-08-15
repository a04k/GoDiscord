package finance

import (
	"DiscordBot/commands"
)

func init() {
	module := &commands.ModuleInfo{
		Name:         "Finance",
		Description:  "Financial market data including cryptocurrency prices",
		Version:      "1.0.0",
		Author:       "Bot Team",
		Category:     "Finance",
		Dependencies: []string{},
		Config: map[string]interface{}{
			"api_timeout":         10,
			"default_vs_currency": "USD",
		},
		Commands: []commands.CommandInfo{
			{
				Name:        "btc",
				Aliases:     []string{},
				Description: "Shows the current Bitcoin price",
				Usage:       ".btc",
				Category:    "Finance",
			},
		},
		SlashCommands: []commands.SlashCommandInfo{},
	}

	commands.RegisterModule(module)

	// Register command handlers
	commands.RegisterCommand("btc", BTC)
}
