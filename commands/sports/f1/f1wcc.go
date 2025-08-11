package f1

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
	"DiscordBot/commands"
)

func init() {
	commands.RegisterCommand("f1wcc", F1WCC)
}

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

	// Format the standings into a table-like string
	standingsStr := ""
	for _, standing := range standings {
		standingsStr += fmt.Sprintf("`%-2s` %-15s %-4s %-3s\n",
			standing.Position,
			TruncateString(standing.Constructor.Name, 15),
			standing.Points,
			standing.Wins)
	}

	// Create an embed with the constructors' championship standings
	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("🏎️ F1 Constructors' Championship %s", season),
		Color: 0xFF0000, // Red color for F1
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Standings",
				Value:  fmt.Sprintf("```\nPos Constructor     Pts  Wins\n%s```", standingsStr),
				Inline: false,
			},
		},
	}

	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}