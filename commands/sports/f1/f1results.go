package f1

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
	"DiscordBot/commands"
)

func init() {
	commands.RegisterCommand("f1results", F1Results)
}

func F1Results(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// Fetch real data from the Ergast API
	apiResponse, err := FetchLatestRaceResults()
	if err != nil {
		log.Printf("Error fetching F1 race results: %v", err)
		s.ChannelMessageSend(m.ChannelID, "Error fetching F1 race results")
		return
	}

	// Check if we have race data
	if len(apiResponse.MRData.RaceTable.Races) == 0 {
		s.ChannelMessageSend(m.ChannelID, "No recent F1 race data available.")
		return
	}

	race := apiResponse.MRData.RaceTable.Races[0]

	// Get current time to find the last event from our schedule
	events, err := FetchF1Events()
	if err != nil {
		log.Printf("Error fetching F1 events: %v", err)
		// Continue with API data even if we can't get the schedule
	}

	now := time.Now().UTC()
	var lastEvent *Event

	// Find the most recent event that has finished
	if events != nil {
		for i := len(events) - 1; i >= 0; i-- {
			event := events[i]
			if len(event.Sessions) == 0 {
				continue
			}

			// Get the last session of the event (typically Race)
			lastSession := event.Sessions[len(event.Sessions)-1]
			eventEndTime, err := time.Parse(time.RFC3339, lastSession.Date)
			if err != nil {
				continue
			}

			// Find the last event that has ended
			if eventEndTime.Before(now) {
				lastEvent = &event
				break
			}
		}
	}

	title := fmt.Sprintf("Latest F1 Race: %s", race.RaceName)
	if lastEvent != nil {
		title = fmt.Sprintf("Latest F1 Race: %s", lastEvent.Name)
	}

	description := fmt.Sprintf("%s, %s", race.Circuit.CircuitName, race.Circuit.Location.Country)
	if lastEvent != nil {
		description = fmt.Sprintf("%s, %s", lastEvent.Location, lastEvent.Circuit)
	}

	// Calculate gaps to leader
	var leaderTimeMillis int64
	resultsWithGaps := make([]struct {
		Position     string
		DriverName   string
		DriverCode   string
		Constructor  string
		Laps         string
		TimeOrStatus string
		Points       string
		Gap          string
	}, len(race.Results))

	for i, result := range race.Results {
		driverName := fmt.Sprintf("%s %s", result.Driver.GivenName, result.Driver.FamilyName)
		timeOrStatus := result.Time.Time
		if timeOrStatus == "" {
			timeOrStatus = result.Status
		}

		resultsWithGaps[i].Position = result.Position
		resultsWithGaps[i].DriverName = driverName
		resultsWithGaps[i].DriverCode = result.Driver.Code
		resultsWithGaps[i].Constructor = result.Constructor.Name
		resultsWithGaps[i].Laps = result.Laps
		resultsWithGaps[i].TimeOrStatus = timeOrStatus
		resultsWithGaps[i].Points = result.Points

		// Calculate gap to leader
		if i == 0 {
			// Leader
			resultsWithGaps[i].Gap = ""
			if result.Time.Millis != "" {
				leaderTimeMillis, _ = strconv.ParseInt(result.Time.Millis, 10, 64)
			}
		} else {
			// Other drivers
			if result.Time.Millis != "" && leaderTimeMillis > 0 {
				driverTimeMillis, _ := strconv.ParseInt(result.Time.Millis, 10, 64)
				gapMillis := driverTimeMillis - leaderTimeMillis
				// Convert milliseconds to seconds
				gapSeconds := float64(gapMillis) / 1000.0
				resultsWithGaps[i].Gap = fmt.Sprintf("+%.3f", gapSeconds)
			} else {
				resultsWithGaps[i].Gap = ""
			}
		}
	}

	// Format the results into a table-like string
	resultsStr := ""
	for _, result := range resultsWithGaps {
		gapDisplay := result.Gap
		if gapDisplay == "" && result.Position != "1" {
			gapDisplay = result.TimeOrStatus
		}

		resultsStr += fmt.Sprintf("`%-2s` %-20s %-10s %-4s %-12s %-3s\n",
			result.Position,
			TruncateString(result.DriverName, 20),
			TruncateString(result.Constructor, 10),
			result.Laps,
			TruncateString(gapDisplay, 12),
			result.Points)
	}

	// Find fastest lap info
	fastestLapInfo := "No fastest lap data"
	if len(race.Results) > 0 {
		for _, result := range race.Results {
			if result.FastestLap.Time.Time != "" {
				fastestLapInfo = fmt.Sprintf("%s %s (%s)",
					result.Driver.GivenName,
					result.Driver.FamilyName,
					result.FastestLap.Time.Time)
				break
			}
		}
	}

	// Find pole position (first driver in grid)
	polePosition := "No pole position data"
	if len(race.Results) > 0 {
		for _, result := range race.Results {
			if result.Position == "1" {
				polePosition = fmt.Sprintf("%s %s",
					result.Driver.GivenName,
					result.Driver.FamilyName)
				break
			}
		}
	}

	// Determine the winner's team color
	var color int = 0xFF0000 // Default red color for F1
	if len(race.Results) > 0 {
		winner := race.Results[0]
		team := winner.Constructor.Name
		if teamColor, exists := TeamColors[team]; exists {
			color = teamColor
		}
	}

	// Create an embed with the event information
	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Color:       color,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Race Results",
				Value:  fmt.Sprintf("```\nPos Driver               Team       Laps Gap/Time     Pts\n%s```", resultsStr),
				Inline: false,
			},
			{
				Name:   "Fastest Lap",
				Value:  fastestLapInfo,
				Inline: true,
			},
			{
				Name:   "Winner",
				Value:  polePosition,
				Inline: true,
			},
		},
	}

	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}