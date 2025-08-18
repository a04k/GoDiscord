package f1

import (
	"strings"

	"DiscordBot/bot"
	"DiscordBot/commands"

	"github.com/bwmarrin/discordgo"
)

// SubcommandInfo holds information about a subcommand
type SubcommandInfo struct {
	Handler     func(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string)
	Description string
	Usage       string
	Aliases     []string
}

// Define subcommands with their information
var f1Subcommands = map[string]SubcommandInfo{
	"wdc": {
		Handler:     F1WDC,
		Description: "Shows the F1 Driver's Championship standings",
		Usage:       ".f1 wdc",
		Aliases:     []string{"f1wdc"},
	},
	"wcc": {
		Handler:     F1WCC,
		Description: "Shows the F1 Constructor's Championship standings",
		Usage:       ".f1 wcc",
		Aliases:     []string{"f1wcc"},
	},
	"results": {
		Handler:     F1Results,
		Description: "Shows F1 race results. Use with session type and/or race name",
		Usage:       ".f1 results [quali|race|fp1|fp2|fp3|sprint|sprintquali] [country/trackname/venue]",
		Aliases:     []string{"f1results"},
	},
	"nextevent": {
		Handler:     F1NextEvent,
		Description: "Shows information about the next F1 event",
		Usage:       ".f1 nextevent",
		Aliases:     []string{},
	},
	"next": {
		Handler:     NextF1Session,
		Description: "Shows information about the next F1 session",
		Usage:       ".f1 next",
		Aliases:     []string{"nextf1session"},
	},
	"sub": {
		Handler:     F1Subscribe,
		Description: "Subscribe/unsubscribe to F1 notifications",
		Usage:       ".f1 sub",
		Aliases:     []string{"f1sub"},
	},
}

// F1 handles all F1 subcommands
func F1(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 2 {
		// Show available subcommands
		usage := "Usage: .f1 <subcommand>\nAvailable subcommands:\n"
		for name, subcmd := range f1Subcommands {
			// Only show primary commands, not aliases
			isAlias := false
			for _, aliasInfo := range f1Subcommands {
				for _, alias := range aliasInfo.Aliases {
					if alias == name {
						isAlias = true
						break
					}
				}
				if isAlias {
					break
				}
			}
			if !isAlias {
				usage += "- " + name + ": " + subcmd.Description + "\n"
			}
		}
		s.ChannelMessageSend(m.ChannelID, usage)
		return
	}

	subcommand := strings.ToLower(args[1])
	
	// Resolve alias if needed
	for name, subcmd := range f1Subcommands {
		for _, alias := range subcmd.Aliases {
			if alias == subcommand {
				subcommand = name
				break
			}
		}
	}

	// Find and execute the handler
	if subcmdInfo, exists := f1Subcommands[subcommand]; exists {
		subcmdInfo.Handler(b, s, m, args[1:])
	} else {
		// Show available subcommands when invalid subcommand is used
		usage := "Unknown F1 subcommand: " + subcommand + "\nAvailable subcommands:\n"
		for name, subcmd := range f1Subcommands {
			// Only show primary commands, not aliases
			isAlias := false
			for _, aliasInfo := range f1Subcommands {
				for _, alias := range aliasInfo.Aliases {
					if alias == name {
						isAlias = true
						break
					}
				}
				if isAlias {
					break
				}
			}
			if !isAlias {
				usage += "- " + name + ": " + subcmd.Description + "\n"
			}
		}
		s.ChannelMessageSend(m.ChannelID, usage)
	}
}

func init() {
	module := &commands.ModuleInfo{
		Name:         "F1",
		Description:  "Formula 1 race results, standings, and live notifications",
		Version:      "1.2.0",
		Author:       "Bot Team",
		Category:     "Sports",
		Dependencies: []string{}, // No dependencies
		Config: map[string]interface{}{
			"notifications_enabled": true,
			"default_timezone":      "UTC",
			"notification_channels": []string{},
		},
		Commands: []commands.CommandInfo{
			{
				Name:        "f1",
				Aliases:     []string{},
				Description: "Formula 1 commands. Use .f1 <subcommand>",
				Usage:       ".f1 <subcommand>",
				Category:    "Sports",
			},
		},
		SlashCommands: []commands.SlashCommandInfo{},
	}

	commands.RegisterModule(module)

	// Register command handlers
	commands.RegisterCommand("f1", F1)
}
