package moderation

import (
	"DiscordBot/commands"
)

func init() {
	module := &commands.ModuleInfo{
		Name:         "Moderation",
		Description:  "Server moderation commands for managing users",
		Version:      "1.0.0",
		Author:       "Bot Team",
		Category:     "Moderation",
		Dependencies: []string{},
		Config: map[string]interface{}{
			"require_reason": true,
			"log_actions":    true,
		},
		Commands: []commands.CommandInfo{
			{
				Name:        "mute",
				Aliases:     []string{},
				Description: "Mutes a user",
				Usage:       ".mute <user>",
				Category:    "Moderation",
			},
			{
				Name:        "unmute",
				Aliases:     []string{},
				Description: "Unmutes a user",
				Usage:       ".unmute <user>",
				Category:    "Moderation",
			},
			{
				Name:        "voicemute",
				Aliases:     []string{},
				Description: "Mutes a user in voice channels",
				Usage:       ".voicemute <user>",
				Category:    "Moderation",
			},
			{
				Name:        "vunmute",
				Aliases:     []string{},
				Description: "Unmutes a user in voice channels",
				Usage:       ".vunmute <user>",
				Category:    "Moderation",
			},
		},
		SlashCommands: []commands.SlashCommandInfo{},
	}

	commands.RegisterModule(module)

	// Register command handlers
	commands.RegisterCommand("mute", Mute)
	commands.RegisterCommand("unmute", Unmute)
	commands.RegisterCommand("voicemute", VoiceMute)
	commands.RegisterCommand("vunmute", VoiceUnmute)
}
