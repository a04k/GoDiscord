package fpl

import (
	"strconv"

	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
)

func SetFPLLeague(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// Check if user is admin
	isAdmin, err := b.IsAdmin(m.GuildID, m.Author.ID)
	if err != nil || !isAdmin {
		s.ChannelMessageSend(m.ChannelID, "You must be an admin to set the FPL league ID.")
		return
	}

	// Check if league ID is provided
	if len(args) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Please provide a league ID. Usage: `.setfplleague <league_id>`")
		return
	}

	// Parse league ID
	leagueID, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Invalid league ID. Please provide a valid number.")
		return
	}

	// Update the database
	_, err = b.Db.Exec("UPDATE guilds SET fpl_league_id = $1 WHERE guild_id = $2", leagueID, m.GuildID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error setting FPL league ID. Please try again later.")
		return
	}

	s.ChannelMessageSend(m.ChannelID, "FPL league ID has been set successfully!")
}