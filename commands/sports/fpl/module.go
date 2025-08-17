package fpl

import (
	"strings"

	"DiscordBot/bot"
	"DiscordBot/commands"

	"github.com/bwmarrin/discordgo"
)

// FPL handles all FPL subcommands
func FPL(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Usage: .fpl <subcommand>\nAvailable subcommands: standings, setleague")
		return
	}

	subcommand := strings.ToLower(args[1])
	switch subcommand {
	case "standings", "fplstandings", "fpl":
		FPLStandings(b, s, m, args[1:])
	case "setleague", "setfplleague":
		SetFPLLeague(b, s, m, args[1:])
	default:
		s.ChannelMessageSend(m.ChannelID, "Unknown FPL subcommand. Available subcommands: standings, setleague")
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
