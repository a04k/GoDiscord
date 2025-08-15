package f1

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
)

// QualiResults displays the qualifying results for the last race
func QualiResults(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	getQualifyingResults(b, s, m)
}

// getQualifyingResults fetches and displays the qualifying results for the last race
func getQualifyingResults(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate) {
	data, err := FetchQualifyingResults()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error fetching qualifying results")
		return
	}

	if len(data.MRData.RaceTable.Races) == 0 {
		s.ChannelMessageSend(m.ChannelID, "No qualifying results data available")
		return
	}

	race := data.MRData.RaceTable.Races[0]
	qualifyingResults := race.QualifyingResults

	// Sort the results by position
	sort.Slice(qualifyingResults, func(i, j int) bool {
		posI, errI := strconv.Atoi(qualifyingResults[i].Position)
		posJ, errJ := strconv.Atoi(qualifyingResults[j].Position)
		if errI != nil || errJ != nil {
			return false // Or handle error appropriately
		}
		return posI < posJ
	})

	// Format the results into a table-like string
	resultsStr := ""
	for _, result := range qualifyingResults {
		driverName := fmt.Sprintf("%s %s", result.Driver.GivenName, result.Driver.FamilyName)
		resultsStr += fmt.Sprintf("`%-2s` %-20s %-10s %-10s %-10s %-10s\n",
			result.Position,
			TruncateString(driverName, 20),
			TruncateString(result.Driver.Code, 4),
			TruncateString(result.Constructor.Name, 10),
			formatTime(result.Q1),
			formatTime(result.Q2))
	}

	// Determine the pole sitter's team color
	var color int = 0xFF0000 // Default red color for F1
	if len(qualifyingResults) > 0 {
		poleSitter := qualifyingResults[0]
		driverCode := poleSitter.Driver.Code
		if team, exists := DriverInfo[driverCode]; exists {
			if teamColor, exists := TeamColors[team]; exists {
				color = teamColor
			}
		}
	}

	// Create an embed with the qualifying results
	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("🏁 %s Qualifying Results", race.RaceName),
		Description: fmt.Sprintf("%s, %s", race.Circuit.CircuitName, race.Circuit.Location.Country),
		Color: color,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Results",
				Value:  fmt.Sprintf("```\nPos Driver               Code Team       Q1        Q2        Q3      \n%s```", resultsStr),
				Inline: false,
			},
		},
	}

	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}

// formatTime formats a time string, returning "—" for empty times
func formatTime(timeStr string) string {
	if timeStr == "" {
		return "—"
	}
	return timeStr
}