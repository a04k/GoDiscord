package f1

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
)

func F1Results(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// If no additional arguments, show session selection buttons for the most recent event
	if len(args) <= 1 {
		ShowSessionSelection(b, s, m)
		return
	}

	// Parse arguments
	sessionType := ""
	raceQuery := ""
	
	if len(args) >= 2 {
		// Check if the second argument is a session type
		possibleSession := strings.ToLower(args[1])
		if possibleSession == "quali" || possibleSession == "qualifying" || 
		   possibleSession == "fp1" || possibleSession == "fp2" || possibleSession == "fp3" || 
		   possibleSession == "practice" || possibleSession == "sprint" || 
		   possibleSession == "sprintquali" || possibleSession == "sprintqualifying" || 
		   possibleSession == "race" {
			sessionType = possibleSession
			// Check if there's a race query
			if len(args) >= 3 {
				raceQuery = strings.ToLower(strings.Join(args[2:], " "))
			}
		} else {
			// Second argument is the race query, default to race session
			sessionType = "race"
			raceQuery = strings.ToLower(strings.Join(args[1:], " "))
		}
	}
	
	// Handle different session types
	switch sessionType {
	case "quali", "qualifying":
		GetQualifyingResults(b, s, m, raceQuery)
	case "fp1", "fp2", "fp3", "practice":
		GetPracticeResults(b, s, m, sessionType, raceQuery)
	case "sprint":
		GetSprintResults(b, s, m, raceQuery)
	case "sprintquali", "sprintqualifying":
		GetSprintQualifyingResults(b, s, m, raceQuery)
	case "race":
		GetRaceResults(b, s, m, raceQuery)
	default:
		// If an invalid session type is provided, show race results by default
		GetRaceResults(b, s, m, raceQuery)
	}
}

func ShowSessionSelection(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate) {
	events, err := FetchF1Events()
	if err != nil {
		log.Printf("Error fetching F1 events: %v", err)
		s.ChannelMessageSend(m.ChannelID, "Error fetching F1 schedule")
		return
	}

	now := time.Now().UTC()
	var previousEvent *Event

	// Find the most recent event that has ended
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

		// If this event has ended, it's our target
		if eventEndTime.Before(now) {
			previousEvent = &event
			break
		}
	}

	if previousEvent == nil {
		s.ChannelMessageSend(m.ChannelID, "No previous F1 events found.")
		return
	}

	// Create buttons for each session
	var components []discordgo.MessageComponent
	var sessionButtons []discordgo.MessageComponent

	for _, session := range previousEvent.Sessions {
		button := discordgo.Button{
			Label: session.Name,
			Style: discordgo.PrimaryButton,
			CustomID: fmt.Sprintf("f1_results_%s_%s", 
				strings.ToLower(strings.ReplaceAll(session.Name, " ", "_")), 
				strings.ToLower(strings.ReplaceAll(previousEvent.Name, " ", "_"))),
		}
		sessionButtons = append(sessionButtons, button)
	}

	// Group buttons in rows of 5
	for i := 0; i < len(sessionButtons); i += 5 {
		end := i + 5
		if end > len(sessionButtons) {
			end = len(sessionButtons)
		}
		components = append(components, discordgo.ActionsRow{
			Components: sessionButtons[i:end],
		})
	}

	// Send message with buttons
	s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
		Embed: &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("F1 Results - %s", previousEvent.Name),
			Description: fmt.Sprintf("Select a session to view results for %s", previousEvent.Name),
			Color:       0xFF0000,
		},
		Components: components,
	})
}

func GetRaceResults(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, raceQuery string) {
	// If raceQuery is empty, get the latest race results
	if raceQuery == "" {
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
		displayRaceResults(s, m, race)
		return
	}

	// If raceQuery is provided, find the specific race
	events, err := FetchF1Events()
	if err != nil {
		log.Printf("Error fetching F1 events: %v", err)
		s.ChannelMessageSend(m.ChannelID, "Error fetching F1 schedule")
		return
	}

	// Find the event that matches the query
	var targetEvent *Event
	for _, event := range events {
		// Check if the query matches the event name, location, or circuit
		if strings.Contains(strings.ToLower(event.Name), raceQuery) ||
			strings.Contains(strings.ToLower(event.Location), raceQuery) ||
			strings.Contains(strings.ToLower(event.Circuit), raceQuery) {
			targetEvent = &event
			break
		}
	}

	if targetEvent == nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("No F1 event found matching '%s'", raceQuery))
		return
	}

	// Find the round number for this event
	round := findEventRound(events, targetEvent)
	if round == 0 {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Could not determine round for event '%s'", targetEvent.Name))
		return
	}

	// Fetch results for the specific race
	apiResponse, err := FetchRaceResultsByRound(round)
	if err != nil {
		log.Printf("Error fetching F1 race results for round %d: %v", round, err)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error fetching F1 race results for %s", targetEvent.Name))
		return
	}

	// Check if we have race data
	if len(apiResponse.MRData.RaceTable.Races) == 0 {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("No race data available for %s", targetEvent.Name))
		return
	}

	race := apiResponse.MRData.RaceTable.Races[0]
	displayRaceResults(s, m, race)
}

func displayRaceResults(s *discordgo.Session, m *discordgo.MessageCreate, race RaceResultsResponse_MRDatum_RaceTable_Race) {
	// Calculate gaps to leader
	var leaderTimeMillis int64
	resultsWithGaps := make([]struct {
		Position     string
		DriverName   string
		DriverCode   string
		Constructor  struct {
			Name string
		}
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
		resultsWithGaps[i].Constructor.Name = result.Constructor.Name
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
			TruncateString(result.Constructor.Name, 10),
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
		Title:       fmt.Sprintf("🏁 %s Results", race.RaceName),
		Description: fmt.Sprintf("%s, %s", race.Circuit.CircuitName, race.Circuit.Location.Country),
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

func GetQualifyingResults(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, raceQuery string) {
	// If raceQuery is empty, get the latest qualifying results
	if raceQuery == "" {
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
		displayQualifyingResults(s, m, race)
		return
	}

	// If raceQuery is provided, find the specific race
	events, err := FetchF1Events()
	if err != nil {
		log.Printf("Error fetching F1 events: %v", err)
		s.ChannelMessageSend(m.ChannelID, "Error fetching F1 schedule")
		return
	}

	// Find the event that matches the query
	var targetEvent *Event
	raceQueryLower := strings.ToLower(raceQuery)
	
	// First try direct matching
	for _, event := range events {
		// Check if the query matches the event name, location, or circuit
		eventNameLower := strings.ToLower(event.Name)
		eventLocationLower := strings.ToLower(event.Location)
		eventCircuitLower := strings.ToLower(event.Circuit)
		
		if strings.Contains(eventNameLower, raceQueryLower) ||
			strings.Contains(eventLocationLower, raceQueryLower) ||
			strings.Contains(eventCircuitLower, raceQueryLower) ||
			// Special case for common names
			strings.Contains(eventNameLower, "grand prix") && strings.Contains(raceQueryLower, "gp") {
			targetEvent = &event
			break
		}
	}

	// Try alternative matching for common names if direct matching failed
	if targetEvent == nil {
		for _, event := range events {
			eventNameLower := strings.ToLower(event.Name)
			eventCircuitLower := strings.ToLower(event.Circuit)
			
			// Handle cases like "spa" matching "Belgian Grand Prix"
			if (strings.Contains(raceQueryLower, "spa") && strings.Contains(eventNameLower, "belgian")) ||
				(strings.Contains(raceQueryLower, "hungary") && strings.Contains(eventNameLower, "hungarian")) ||
				(strings.Contains(raceQueryLower, "hungaroring") && strings.Contains(eventNameLower, "hungarian")) ||
				(strings.Contains(raceQueryLower, "budapest") && strings.Contains(eventNameLower, "hungarian")) ||
				(strings.Contains(raceQueryLower, "monaco") && strings.Contains(eventNameLower, "monaco")) ||
				(strings.Contains(raceQueryLower, "silverstone") && strings.Contains(eventNameLower, "british")) ||
				(strings.Contains(raceQueryLower, "monza") && strings.Contains(eventNameLower, "italian")) ||
				// Additional matching for "spa"
				(strings.Contains(raceQueryLower, "spa") && strings.Contains(eventCircuitLower, "spa")) {
				targetEvent = &event
				break
			}
		}
	}

	if targetEvent == nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("No F1 event found matching '%s'", raceQuery))
		return
	}

	// Find the round number for this event
	round := findEventRound(events, targetEvent)
	if round == 0 {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Could not determine round for event '%s'", targetEvent.Name))
		return
	}

	// Fetch qualifying results for the specific race
	apiResponse, err := FetchQualifyingResultsByRound(round)
	if err != nil {
		log.Printf("Error fetching F1 qualifying results for round %d: %v", round, err)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error fetching F1 qualifying results for %s", targetEvent.Name))
		return
	}

	// Check if we have qualifying data
	if len(apiResponse.MRData.RaceTable.Races) == 0 {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("No qualifying data available for %s", targetEvent.Name))
		return
	}

	race := apiResponse.MRData.RaceTable.Races[0]
	displayQualifyingResults(s, m, race)
}

func displayQualifyingResults(s *discordgo.Session, m *discordgo.MessageCreate, race QualifyingResponse_MRDatum_RaceTable_Race) {
	// Calculate gaps to leader for each qualifying session
	var q1LeaderTime, q2LeaderTime, q3LeaderTime float64
	
	// Parse Q1 times and find leader
	q1Times := make(map[string]float64)
	for _, result := range race.QualifyingResults {
		if result.Q1 != "" {
			if time, err := ParseQualifyingTime(result.Q1); err == nil {
				q1Times[result.Driver.Code] = time
				if q1LeaderTime == 0 || time < q1LeaderTime {
					q1LeaderTime = time
				}
			}
		}
	}
	
	// Parse Q2 times and find leader
	q2Times := make(map[string]float64)
	for _, result := range race.QualifyingResults {
		if result.Q2 != "" {
			if time, err := ParseQualifyingTime(result.Q2); err == nil {
				q2Times[result.Driver.Code] = time
				if q2LeaderTime == 0 || time < q2LeaderTime {
					q2LeaderTime = time
				}
			}
		}
	}
	
	// Parse Q3 times and find leader
	q3Times := make(map[string]float64)
	for _, result := range race.QualifyingResults {
		if result.Q3 != "" {
			if time, err := ParseQualifyingTime(result.Q3); err == nil {
				q3Times[result.Driver.Code] = time
				if q3LeaderTime == 0 || time < q3LeaderTime {
					q3LeaderTime = time
				}
			}
		}
	}
	
	// Format the results into a table-like string
	resultsStr := ""
	for _, result := range race.QualifyingResults {
		driverName := fmt.Sprintf("%s %s", result.Driver.GivenName, result.Driver.FamilyName)
		
		// Calculate gaps to leader for each session
		q1Gap := ""
		q2Gap := ""
		q3Gap := ""
		
		if result.Q1 != "" {
			if time, exists := q1Times[result.Driver.Code]; exists && q1LeaderTime > 0 {
				if time == q1LeaderTime {
					q1Gap = "1st"
				} else {
					gap := time - q1LeaderTime
					q1Gap = fmt.Sprintf("+%.3f", gap)
				}
			}
		} else {
			q1Gap = "—"
		}
		
		if result.Q2 != "" {
			if time, exists := q2Times[result.Driver.Code]; exists && q2LeaderTime > 0 {
				if time == q2LeaderTime {
					q2Gap = "1st"
				} else {
					gap := time - q2LeaderTime
					q2Gap = fmt.Sprintf("+%.3f", gap)
				}
			}
		} else {
			q2Gap = "—"
		}
		
		if result.Q3 != "" {
			if time, exists := q3Times[result.Driver.Code]; exists && q3LeaderTime > 0 {
				if time == q3LeaderTime {
					q3Gap = "1st"
				} else {
					gap := time - q3LeaderTime
					q3Gap = fmt.Sprintf("+%.3f", gap)
				}
			}
		} else {
			q3Gap = "—"
		}
		
		resultsStr += fmt.Sprintf("`%-2s` %-20s %-4s %-10s %-10s %-10s %-10s\n",
			result.Position,
			TruncateString(driverName, 20),
			TruncateString(result.Driver.Code, 4),
			TruncateString(result.Constructor.Name, 10),
			TruncateString(q1Gap, 10),
			TruncateString(q2Gap, 10),
			TruncateString(q3Gap, 10))
	}

	// Determine the pole sitter's team color
	var color int = 0xFF0000 // Default red color for F1
	if len(race.QualifyingResults) > 0 {
		poleSitter := race.QualifyingResults[0]
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
				Value:  fmt.Sprintf("```\nPos Driver               Code Team       Q1 Gap    Q2 Gap    Q3 Gap   \n%s```", resultsStr),
				Inline: false,
			},
		},
	}

	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}

func GetPracticeResults(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, sessionType, raceQuery string) {
	s.ChannelMessageSend(m.ChannelID, "Practice session results functionality to be implemented")
}

func GetSprintResults(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, raceQuery string) {
	s.ChannelMessageSend(m.ChannelID, "Sprint results functionality to be implemented")
}

func GetSprintQualifyingResults(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, raceQuery string) {
	s.ChannelMessageSend(m.ChannelID, "Sprint qualifying results functionality to be implemented")
}

// Helper function to find the round number of an event
func findEventRound(events []Event, targetEvent *Event) int {
	for i, event := range events {
		if event.Name == targetEvent.Name {
			return i + 1 // Rounds are 1-indexed
		}
	}
	return 0
}

// formatTime formats a time string, returning "—" for empty times
func formatTime(timeStr string) string {
	if timeStr == "" {
		return "—"
	}
	return timeStr
}

// ParseQualifyingTime parses a qualifying time string (e.g., "1:32.323") into seconds
func ParseQualifyingTime(timeStr string) (float64, error) {
	if timeStr == "" {
		return 0, fmt.Errorf("empty time string")
	}
	
	// Handle format like "1:32.323" (minutes:seconds.milliseconds)
	if strings.Contains(timeStr, ":") {
		parts := strings.Split(timeStr, ":")
		if len(parts) != 2 {
			return 0, fmt.Errorf("invalid time format")
		}
		
		minutes, err := strconv.ParseFloat(parts[0], 64)
		if err != nil {
			return 0, err
		}
		
		seconds, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return 0, err
		}
		
		return minutes*60 + seconds, nil
	}
	
	// Handle format like "52.323" (seconds.milliseconds)
	seconds, err := strconv.ParseFloat(timeStr, 64)
	if err != nil {
		return 0, err
	}
	
	return seconds, nil
}