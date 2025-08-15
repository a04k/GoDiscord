package general

import (
	"DiscordBot/commands"
)

func init() {
	module := &commands.ModuleInfo{
		Name:         "General",
		Description:  "General utility commands including reminders",
		Version:      "1.0.0",
		Author:       "Bot Team",
		Category:     "General",
		Dependencies: []string{},
		Config: map[string]interface{}{
			"default_reminder_limit": 10,
		},
		Commands: []commands.CommandInfo{
			{
				Name:        "remindme",
				Aliases:     []string{},
				Description: "Set a reminder for yourself",
				Usage:       ".remindme <duration> <message>",
				Category:    "General",
			},
		},
		SlashCommands: []commands.SlashCommandInfo{},
	}

	commands.RegisterModule(module)

	// Register command handlers
	commands.RegisterCommand("remindme", RemindMe)
}
