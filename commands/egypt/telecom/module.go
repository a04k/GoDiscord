package telecom

import (
	"DiscordBot/commands"

	"github.com/bwmarrin/discordgo"
)

func init() {
	module := &commands.ModuleInfo{
		Name:         "EgyptTelecom",
		Description:  "WE (Egypt telecom) quota checking and account management",
		Version:      "1.0.0",
		Author:       "Bot Team",
		Category:     "Egypt",
		Dependencies: []string{},
		Config: map[string]interface{}{
			"default_timeout": 30,
		},
		Commands: []commands.CommandInfo{}, // No regular commands, only slash commands
		SlashCommands: []commands.SlashCommandInfo{
			{
				Name:        "wequota",
				Description: "Check your WE internet quota",
				Options:     []*discordgo.ApplicationCommandOption{},
				Handler:     WEQuota,
			},
			{
				Name:        "wesetup",
				Description: "Setup your WE account credentials",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "landline",
						Description: "Your WE landline number",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "password",
						Description: "Your WE account password",
						Required:    true,
					},
				},
				Handler: WEAccountSetup,
			},
		},
	}

	commands.RegisterModule(module)
}
