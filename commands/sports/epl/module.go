package epl

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
var eplSubcommands = map[string]SubcommandInfo{
	"table": {
		Handler:     EPLTable,
		Description: "Shows the current Premier League table",
		Usage:       ".epl table",
		Aliases:     []string{"epltable"},
	},
	"next": {
		Handler:     NextMatch,
		Description: "Shows upcoming Premier League fixtures. Use with a club name to see their next match.",
		Usage:       ".epl next [club_name]",
		Aliases:     []string{"nextmatch"},
	},
	"matchups": {
		Handler:     ShowMatchups,
		Description: "Shows previous matchups between two clubs",
		Usage:       ".epl matchups <club1> <club2> [season]",
		Aliases:     []string{},
	},
	"results": {
		Handler:     ShowResults,
		Description: "Shows results for a specific club or gameweek",
		Usage:       ".epl results <club_name> or .epl results <gameweek>",
		Aliases:     []string{},
	},
}

// EPL handles all EPL subcommands
func EPL(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 2 {
		// Show available subcommands
		usage := "Usage: .epl <subcommand>\nAvailable subcommands:\n"
		for name, subcmd := range eplSubcommands {
			// Only show primary commands, not aliases
			isAlias := false
			for _, aliasInfo := range eplSubcommands {
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
	for name, subcmd := range eplSubcommands {
		for _, alias := range subcmd.Aliases {
			if alias == subcommand {
				subcommand = name
				break
			}
		}
	}

	// Find and execute the handler
	if subcmdInfo, exists := eplSubcommands[subcommand]; exists {
		subcmdInfo.Handler(b, s, m, args[1:])
	} else {
		// Show available subcommands when invalid subcommand is used
		usage := "Unknown EPL subcommand: " + subcommand + "\nAvailable subcommands:\n"
		for name, subcmd := range eplSubcommands {
			// Only show primary commands, not aliases
			isAlias := false
			for _, aliasInfo := range eplSubcommands {
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
		Name:         "EPL",
		Description:  "English Premier League table and fixture information",
		Version:      "1.0.0",
		Author:       "Bot Team",
		Category:     "Sports",
		Dependencies: []string{},
		Config: map[string]interface{}{
			"default_season": "current",
		},
		Commands: []commands.CommandInfo{
			{
				Name:        "epl",
				Aliases:     []string{},
				Description: "English Premier League commands. Use .epl <subcommand>",
				Usage:       ".epl <subcommand>",
				Category:    "Sports",
			},
		},
		SlashCommands: []commands.SlashCommandInfo{},
	}

	commands.RegisterModule(module)

	// Register command handlers
	commands.RegisterCommand("epl", EPL)
}
