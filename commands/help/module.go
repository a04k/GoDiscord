package help

import (
	"DiscordBot/commands"
)

func init() {
	module := &commands.ModuleInfo{
		Name:         "Help",
		Description:  "Help system with command documentation and paginated lists",
		Version:      "1.0.0",
		Author:       "Bot Team",
		Category:     "General",
		Dependencies: []string{},
		Config: map[string]interface{}{
			"commands_per_page": 10,
			"show_aliases":      true,
		},
		Commands: []commands.CommandInfo{
			{
				Name:        "help",
				Aliases:     []string{"h"},
				Description: "Displays help information for commands",
				Usage:       ".help [command]",
				Category:    "General",
			},
			{
				Name:        "commandlist",
				Aliases:     []string{"cl"},
				Description: "Lists all available commands",
				Usage:       ".commandlist",
				Category:    "General",
			},
		},
		SlashCommands: []commands.SlashCommandInfo{},
	}

	commands.RegisterModule(module)

	// Register command handlers
	commands.RegisterCommand("help", Help, "h")
	commands.RegisterCommand("commandlist", CommandList, "cl")
}
