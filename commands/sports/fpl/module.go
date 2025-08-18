package fpl

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
var fplSubcommands = map[string]SubcommandInfo{
	"standings": {
		Handler:     FPLStandings,
		Description: "Shows the standings for your guild's Fantasy Premier League",
		Usage:       ".fpl standings",
		Aliases:     []string{"fplstandings", "fpl"},
	},
	"setleague": {
		Handler:     SetFPLLeague,
		Description: "Sets the Fantasy Premier League league ID for your guild (Admin only)",
		Usage:       ".fpl setleague <league_id>",
		Aliases:     []string{"setfplleague"},
	},
}

// FPL handles all FPL subcommands
func FPL(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 2 {
		// Show available subcommands
		usage := "Usage: .fpl <subcommand>\nAvailable subcommands:\n"
		for name, subcmd := range fplSubcommands {
			// Only show primary commands, not aliases
			isAlias := false
			for _, aliasInfo := range fplSubcommands {
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
	for name, subcmd := range fplSubcommands {
		for _, alias := range subcmd.Aliases {
			if alias == subcommand {
				subcommand = name
				break
			}
		}
	}

	// Find and execute the handler
	if subcmdInfo, exists := fplSubcommands[subcommand]; exists {
		subcmdInfo.Handler(b, s, m, args[1:])
	} else {
		// Show available subcommands when invalid subcommand is used
		usage := "Unknown FPL subcommand: " + subcommand + "\nAvailable subcommands:\n"
		for name, subcmd := range fplSubcommands {
			// Only show primary commands, not aliases
			isAlias := false
			for _, aliasInfo := range fplSubcommands {
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
		Name:         "FPL",
		Description:  "Fantasy Premier League standings and league management",
		Version:      "1.0.0",
		Author:       "Bot Team",
		Category:     "Sports",
		Dependencies: []string{"EPL"},
		Config: map[string]interface{}{
			"default_league_id": nil,
		},
		Commands: []commands.CommandInfo{
			{
				Name:        "fpl",
				Aliases:     []string{},
				Description: "Fantasy Premier League commands. Use .fpl <subcommand>",
				Usage:       ".fpl <subcommand>",
				Category:    "Sports",
			},
		},
		SlashCommands: []commands.SlashCommandInfo{},
	}

	commands.RegisterModule(module)

	// Register command handlers
	commands.RegisterCommand("fpl", FPL)
}
