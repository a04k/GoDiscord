package f1

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
	"DiscordBot/commands"
)

func init() {
	commands.RegisterCommand("f1wdc", F1WDC)
}

// F1WDC displays the drivers' championship standings or a specific driver's position
func F1WDC(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// If there are arguments, it means we're looking for a specific driver
	if len(args) > 1 {
		driverQuery := strings.Join(args[1:], " ")
		getSpecificDriverStanding(b, s, m, driverQuery)
		return
	}

	// Otherwise, show the full drivers' championship
	getDriversChampionship(b, s, m)
}

// getDriversChampionship fetches and displays the full drivers' championship
func getDriversChampionship(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate) {
	data, err := FetchDriverStandings()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error fetching driver standings")
		return
	}

	if len(data.MRData.StandingsTable.StandingsLists) == 0 {
		s.ChannelMessageSend(m.ChannelID, "No driver standings data available")
		return
	}

	standings := data.MRData.StandingsTable.StandingsLists[0].DriverStandings
	season := data.MRData.StandingsTable.Season

	// Format the standings into a table-like string
	standingsStr := ""
	for _, standing := range standings {
		driverName := fmt.Sprintf("%s %s", standing.Driver.GivenName, standing.Driver.FamilyName)
		standingsStr += fmt.Sprintf("`%-2s` %-20s %-4s %-3s\n",
			standing.Position,
			TruncateString(driverName, 20),
			standing.Points,
			standing.Wins)
	}

	// Create an embed with the drivers' championship standings
	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("🏎️ F1 Drivers' Championship %s", season),
		Color: 0xFF0000, // Red color for F1
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Standings",
				Value:  fmt.Sprintf("```\nPos Driver               Pts  Wins\n%s```", standingsStr),
				Inline: false,
			},
		},
	}

	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}

// getSpecificDriverStanding fetches and displays a specific driver's championship position
func getSpecificDriverStanding(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, driverQuery string) {
	data, err := FetchDriverStandings()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error fetching driver standings")
		return
	}

	if len(data.MRData.StandingsTable.StandingsLists) == 0 {
		s.ChannelMessageSend(m.ChannelID, "No driver standings data available")
		return
	}

	standings := data.MRData.StandingsTable.StandingsLists[0].DriverStandings

	// Find the driver matching the query (name or number)
	var foundStanding *struct {
		Position  string `json:"position"`
		Points    string `json:"points"`
		Wins      string `json:"wins"`
		Driver    struct {
			GivenName  string `json:"givenName"`
			FamilyName string `json:"familyName"`
			Code       string `json:"code"`
		} `json:"Driver"`
	}

	driverQuery = strings.ToLower(driverQuery)

	for _, standing := range standings {
		driverName := fmt.Sprintf("%s %s", standing.Driver.GivenName, standing.Driver.FamilyName)
		driverCode := standing.Driver.Code

		// Check if the query matches the driver's name or code
		if strings.Contains(strings.ToLower(driverName), driverQuery) ||
			strings.Contains(strings.ToLower(driverCode), driverQuery) {
			foundStanding = &standing
			break
		}
	}

	if foundStanding == nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Driver '%s' not found in the championship standings", driverQuery))
		return
	}

	driverName := fmt.Sprintf("%s %s", foundStanding.Driver.GivenName, foundStanding.Driver.FamilyName)

	// Create an embed with the driver's championship position
	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("🏎️ %s - %s", driverName, foundStanding.Driver.Code),
		Color: 0xFF0000, // Red color for F1
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Position",
				Value:  foundStanding.Position,
				Inline: true,
			},
			{
				Name:   "Points",
				Value:  foundStanding.Points,
				Inline: true,
			},
			{
				Name:   "Wins",
				Value:  foundStanding.Wins,
				Inline: true,
			},
		},
	}

	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}