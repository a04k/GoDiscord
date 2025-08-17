package f1

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
)

// F1WCC displays the constructors' championship standings
func F1WCC(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	getConstructorsChampionship(b, s, m)
}

// getConstructorsChampionship fetches and displays the full constructors' championship
func getConstructorsChampionship(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate) {
	data, err := FetchConstructorStandings()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error fetching constructor standings")
		return
	}

	if len(data.MRData.StandingsTable.StandingsLists) == 0 {
		s.ChannelMessageSend(m.ChannelID, "No constructor standings data available")
		return
	}

	standings := data.MRData.StandingsTable.StandingsLists[0].ConstructorStandings
	season := data.MRData.StandingsTable.Season

	// Create an embed with the constructors' championship standings
	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("🏎️ F1 Constructors' Championship %s", season),
		Color: 0xFF0000, // Red color for F1
	}

	// Add constructor standings (all)
	for _, standing := range standings {
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
	}

	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}