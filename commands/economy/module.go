package economy

import (
	"DiscordBot/commands"
)

func init() {
	module := &commands.ModuleInfo{
		Name:         "Economy",
		Description:  "Virtual economy system with balance, work, and gambling commands",
		Version:      "1.0.0",
		Author:       "Bot Team",
		Category:     "Economy",
		Dependencies: []string{},
		Config: map[string]interface{}{
			"starting_balance": 100,
			"work_min_reward":  50,
			"work_max_reward":  200,
			"daily_reward":     500,
		},
		Commands: []commands.CommandInfo{
			{
				Name:        "balance",
				Aliases:     []string{"bal", "$"},
				Description: "Check your current balance",
				Usage:       ".balance [@user]",
				Category:    "Economy",
			},
			{
				Name:        "work",
				Aliases:     []string{"w"},
				Description: "Work to earn money",
				Usage:       ".work",
				Category:    "Economy",
			},
			{
				Name:        "transfer",
				Aliases:     []string{"pay", "give"},
				Description: "Transfer money to another user",
				Usage:       ".transfer @user <amount>",
				Category:    "Economy",
			},
			{
				Name:        "flip",
				Aliases:     []string{},
				Description: "Flip a coin and gamble your money",
				Usage:       ".flip <amount>",
				Category:    "Economy",
			},
			{
				Name:        "setdailyrole",
				Aliases:     []string{},
				Description: "Set a role that modifies the daily cooldown",
				Usage:       ".setdailyrole <role> <hours>",
				Category:    "Economy",
			},
			{
				Name:        "removedailyrole",
				Aliases:     []string{},
				Description: "Remove a role's daily cooldown modifier",
				Usage:       ".removedailyrole <role>",
				Category:    "Economy",
			},
			{
				Name:        "listdailyroles",
				Aliases:     []string{},
				Description: "List all roles with daily cooldown modifiers",
				Usage:       ".listdailyroles",
				Category:    "Economy",
			},
		},
		SlashCommands: []commands.SlashCommandInfo{},
	}

	commands.RegisterModule(module)

	// Register command handlers
	commands.RegisterCommand("balance", Balance, "bal", "$")
	commands.RegisterCommand("work", Work, "w")
	commands.RegisterCommand("transfer", Transfer, "pay", "give")
	commands.RegisterCommand("flip", Flip)
	commands.RegisterCommand("setdailyrole", SetDailyRole)
	commands.RegisterCommand("removedailyrole", RemoveDailyRole)
	commands.RegisterCommand("listdailyroles", ListDailyRoles)
}
