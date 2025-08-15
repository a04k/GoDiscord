package f1

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
)

func F1Standings(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// Fetch real driver standings data from the Ergast API
	driverStandingsResponse, err := FetchDriverStandings()
	if err != nil {
		log.Printf("Error fetching F1 driver standings: %v", err)
		s.ChannelMessageSend(m.ChannelID, "Error fetching F1 driver standings")
		return
	}

	// Fetch real constructor standings data from the Ergast API
	constructorStandingsResponse, err := FetchConstructorStandings()
	if err != nil {
		log.Printf("Error fetching F1 constructor standings: %v", err)
		s.ChannelMessageSend(m.ChannelID, "Error fetching F1 constructor standings")
		return
	}

	// Check if we have standings data
	if len(driverStandingsResponse.MRData.StandingsTable.StandingsLists) == 0 ||
		len(constructorStandingsResponse.MRData.StandingsTable.StandingsLists) == 0 {
		s.ChannelMessageSend(m.ChannelID, "No F1 standings data available.")
		return
	}

	driverStandings := driverStandingsResponse.MRData.StandingsTable.StandingsLists[0].DriverStandings
	constructorStandings := constructorStandingsResponse.MRData.StandingsTable.StandingsLists[0].ConstructorStandings
	season := driverStandingsResponse.MRData.StandingsTable.Season

	// Format the driver standings into a table-like string
	driversStr := ""
	for _, standing := range driverStandings {
		driverName := fmt.Sprintf("%s %s", standing.Driver.GivenName, standing.Driver.FamilyName)
		driversStr += fmt.Sprintf("`%-2s` %-20s %-4s %-3s\n", 
			standing.Position, 
			TruncateString(driverName, 20), 
			standing.Points, 
			standing.Wins)
	}

	// Format the constructor standings into a table-like string
	constructorsStr := ""
	for _, standing := range constructorStandings {
		constructorsStr += fmt.Sprintf("`%-2s` %-15s %-4s %-3s\n", 
			standing.Position, 
			TruncateString(standing.Constructor.Name, 15), 
			standing.Points, 
			standing.Wins)
	}

	// Create an embed with the championship standings
	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("F1 Championship Standings %s", season),
		Color: 0xFF0000, // Red color for F1
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Drivers' Championship",
				Value:  fmt.Sprintf("```\nPos Driver               Pts  Wins\n%s```", driversStr),
				Inline: false,
			},
			{
				Name:   "Constructors' Championship",
				Value:  fmt.Sprintf("```\nPos Constructor     Pts  Wins\n%s```", constructorsStr),
				Inline: false,
			},
		},
	}

	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}