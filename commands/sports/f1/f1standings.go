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

	// Create an embed with the championship standings
	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("F1 Championship Standings %s", season),
		Color: 0xFF0000, // Red color for F1
	}

	// Add driver standings (top 10)
	driversAdded := 0
	for _, standing := range driverStandings {
		if driversAdded >= 10 {
			break
		}
		
		driverName := fmt.Sprintf("%s %s", standing.Driver.GivenName, standing.Driver.FamilyName)
		
		// Add position indicator for top 3
		positionStr := fmt.Sprintf("#%s", standing.Position)
		switch standing.Position {
		case "1":
			positionStr = "🥇 #" + standing.Position
		case "2":
			positionStr = "🥈 #" + standing.Position
		case "3":
			positionStr = "🥉 #" + standing.Position
		}
		
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%s %s (%s)", positionStr, driverName, standing.Driver.Code),
			Value:  fmt.Sprintf("Points: %s\nWins: %s", standing.Points, standing.Wins),
			Inline: false,
		})
		
		driversAdded++
	}

	// Add a separator
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "\u200B", // Zero-width space
		Value:  "\u200B",
		Inline: false,
	})

	// Add constructor standings (top 10)
	constructorsAdded := 0
	for _, standing := range constructorStandings {
		if constructorsAdded >= 10 {
			break
		}
		
		// Add position indicator for top 3
		positionStr := fmt.Sprintf("#%s", standing.Position)
		switch standing.Position {
		case "1":
			positionStr = "🥇 #" + standing.Position
		case "2":
			positionStr = "🥈 #" + standing.Position
		case "3":
			positionStr = "🥉 #" + standing.Position
		}
		
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%s %s", positionStr, standing.Constructor.Name),
			Value:  fmt.Sprintf("Points: %s\nWins: %s", standing.Points, standing.Wins),
			Inline: false,
		})
		
		constructorsAdded++
	}

	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}