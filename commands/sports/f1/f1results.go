package f1

import (
	"fmt"
	"log"
	"strings"
	"time"
	"strconv"
	
	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
)

func F1Results(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	log.Printf("F1Results called with args: %v", args)
	
	if len(args) == 1 {
		log.Printf("No additional arguments, showing session selection")
		ShowSessionSelection(b, s, m)
		return
	}

	sessionType := "race" // Default to race results
	raceQuery := ""

	// User might provide session type, race query, or both
	// Examples:
	// .f1results spa -> raceQuery: "spa", sessionType: "race"
	// .f1results quali spa -> sessionType: "quali", raceQuery: "spa"
	// .f1results sprint -> sessionType: "sprint", raceQuery: "" (latest)
	if len(args) > 1 {
		firstArg := strings.ToLower(args[1])
		log.Printf("First arg: %s, isSessionType: %v", firstArg, isSessionType(firstArg))
		if isSessionType(firstArg) {
			sessionType = firstArg
			if len(args) > 2 {
				raceQuery = strings.ToLower(strings.Join(args[2:], " "))
			}
		} else {
			raceQuery = strings.ToLower(strings.Join(args[1:], " "))
		}
	}
	
	log.Printf("Session type: %s, Race query: %s", sessionType, raceQuery)

	// Find the round for the given query
	round := 0
	if raceQuery != "" {
		log.Printf("Fetching F1 events for query: %s", raceQuery)
		events, err := FetchF1Events()
		if err != nil {
			log.Printf("Error fetching F1 events: %v", err)
			s.ChannelMessageSend(m.ChannelID, "Error fetching F1 schedule")
			return
		}
		event := findEventByQuery(events, raceQuery)
		if event == nil {
			log.Printf("No F1 event found matching '%s'", raceQuery)
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("No F1 event found matching '%s'", raceQuery))
			return
		}
		log.Printf("Found event: %s", event.Name)
		round = findEventRound(events, *event)
		log.Printf("Event round: %d", round)
	}

	switch sessionType {
	case "quali", "qualifying":
		log.Printf("Getting qualifying results for round %d", round)
		GetQualifyingResults(b, s, m, round)
	case "sprint":
		log.Printf("Getting sprint results for round %d", round)
		GetSprintResults(b, s, m, round)
	case "race":
		log.Printf("Getting race results for round %d", round)
		GetRaceResults(b, s, m, round)
	default:
		log.Printf("Unsupported session type: %s", sessionType)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Session type '%s' is not supported or results are not available.", sessionType))
	}
}

func ShowSessionSelection(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate) {
	events, err := FetchF1Events()
	if err != nil {
		log.Printf("Error fetching F1 events: %v", err)
		s.ChannelMessageSend(m.ChannelID, "Error fetching F1 schedule")
		return
	}

	previousEvent := getPreviousEvent(events)
	if previousEvent == nil {
		s.ChannelMessageSend(m.ChannelID, "Could not determine the previous F1 event.")
		return
	}
	round := findEventRound(events, *previousEvent)

	var components []discordgo.MessageComponent
	var sessionButtons []discordgo.MessageComponent

	for _, session := range previousEvent.Sessions {
		sessionType := strings.ToLower(session.Type)
		// Ergast API does not support practice results, so we disable the buttons
		disabled := strings.Contains(sessionType, "practice")

		button := discordgo.Button{
			Label:    session.Name,
			Style:    discordgo.PrimaryButton,
			CustomID: fmt.Sprintf("f1_results_%s_%d", sessionType, round),
			Disabled: disabled,
		}
		sessionButtons = append(sessionButtons, button)
	}

	for i := 0; i < len(sessionButtons); i += 5 {
		end := i + 5
		if end > len(sessionButtons) {
			end = len(sessionButtons)
		}
		components = append(components, discordgo.ActionsRow{
			Components: sessionButtons[i:end],
		})
	}

	s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
		Embed: &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("F1 Results - %s", previousEvent.Name),
			Description: fmt.Sprintf("Select a session to view results for the %s.", previousEvent.Name),
			Color:       0xFF0000, // Red for F1
		},
		Components: components,
	})
}

func GetRaceResults(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, round int) {
	log.Printf("GetRaceResults called with round: %d", round)
	var (
		apiResponse *RaceResultsResponse
		err         error
	)

	if round == 0 {
		log.Printf("Fetching latest race results")
		apiResponse, err = FetchLatestRaceResults()
	} else {
		log.Printf("Fetching race results for round: %d", round)
		apiResponse, err = FetchRaceResultsByRound(round)
	}

	if err != nil {
		log.Printf("Error fetching F1 race results: %v", err)
		s.ChannelMessageSend(m.ChannelID, "Error fetching F1 race results.")
		return
	}

	log.Printf("API response: %v", apiResponse)
	if apiResponse == nil {
		log.Printf("API response is nil")
		s.ChannelMessageSend(m.ChannelID, "No recent F1 race data available.")
		return
	}
	
	log.Printf("Number of races in response: %d", len(apiResponse.MRData.RaceTable.Races))
	if len(apiResponse.MRData.RaceTable.Races) == 0 {
		log.Printf("No races in response")
		s.ChannelMessageSend(m.ChannelID, "No recent F1 race data available.")
		return
	}

	race := apiResponse.MRData.RaceTable.Races[0]
	log.Printf("Displaying race results for: %s", race.RaceName)
	displayRaceResults(s, m.ChannelID, race)
}

func GetQualifyingResults(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, round int) {
	var (
		data *QualifyingResponse
		err  error
	)
	if round == 0 {
		data, err = FetchLatestQualifyingResults()
	} else {
		data, err = FetchQualifyingResultsByRound(round)
	}

	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error fetching qualifying results.")
		return
	}

	if data == nil || len(data.MRData.RaceTable.Races) == 0 {
		s.ChannelMessageSend(m.ChannelID, "No qualifying results data available.")
		return
	}

	race := data.MRData.RaceTable.Races[0]
	displayQualifyingResults(s, m.ChannelID, race)
}

func GetSprintResults(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, round int) {
	var (
		apiResponse *SprintResultsResponse
		err         error
	)

	if round == 0 {
		s.ChannelMessageSend(m.ChannelID, "Please specify a race to get sprint results for.")
		return
	}
	apiResponse, err = FetchSprintResultsByRound(round)

	if err != nil {
		log.Printf("Error fetching F1 sprint results: %v", err)
		s.ChannelMessageSend(m.ChannelID, "Error fetching F1 sprint results.")
		return
	}

	if apiResponse == nil || len(apiResponse.MRData.RaceTable.Races) == 0 {
		s.ChannelMessageSend(m.ChannelID, "No F1 sprint data available for this round.")
		return
	}

	race := apiResponse.MRData.RaceTable.Races[0]
	displaySprintResults(s, m.ChannelID, race)
}

func displayRaceResults(s *discordgo.Session, channelID string, race RaceResultsResponse_MRDatum_RaceTable_Race) {
	log.Printf("displayRaceResults called for: %s", race.RaceName)
	var resultsStr strings.Builder
	resultsStr.WriteString("```\nPos Driver               Team       Laps   Time\n")

	// Find the winner's lap count to determine lapped drivers
	winnerLaps := 0
	if len(race.Results) > 0 {
		winnerLaps, _ = strconv.Atoi(race.Results[0].Laps)
	}

	for _, result := range race.Results {
		driverName := fmt.Sprintf("%s %s", result.Driver.GivenName, result.Driver.FamilyName)
		laps, _ := strconv.Atoi(result.Laps)
		
		var timeOrStatus string
		if result.Time.Time != "" {
			timeOrStatus = result.Time.Time
		} else if result.Status != "Finished" && result.Status != "" {
			timeOrStatus = result.Status
		} else if laps < winnerLaps {
			// Driver is lapped
			lapDiff := winnerLaps - laps
			if lapDiff == 1 {
				timeOrStatus = "+1 Lap"
			} else {
				timeOrStatus = fmt.Sprintf("+%d Laps", lapDiff)
			}
		} else {
			timeOrStatus = "Finished"
		}

		line := fmt.Sprintf("%-3s %-20s %-10s %-5s %-12s", result.Position, TruncateString(driverName, 18), TruncateString(result.Constructor.Name, 10), result.Laps, TruncateString(timeOrStatus, 12))
		resultsStr.WriteString(line + "\n")
	}
	resultsStr.WriteString("```")
	
	log.Printf("Results string: %s", resultsStr.String())

	color := 0xFF0000 // Default F1 red
	if len(race.Results) > 0 {
		winnerTeam := race.Results[0].Constructor.Name
		if c, ok := TeamColors[winnerTeam]; ok {
			color = c
		}
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("🏁 %s Results", race.RaceName),
		Description: fmt.Sprintf("**%s**\n%s, %s", race.RaceName, race.Circuit.CircuitName, race.Circuit.Location.Country),
		Color:       color,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Race Results", Value: resultsStr.String()},
		},
		Footer: &discordgo.MessageEmbedFooter{Text: "Formula 1 Results"},
	}
	
	log.Printf("Sending embed to channel: %s", channelID)
	_, err := s.ChannelMessageSendEmbed(channelID, embed)
	if err != nil {
		log.Printf("Error sending race results embed: %v", err)
	}
}

func displayQualifyingResults(s *discordgo.Session, channelID string, race QualifyingResponse_MRDatum_RaceTable_Race) {
	var resultsStr strings.Builder
	resultsStr.WriteString("```\nPos Driver           Team         Q1         Q2         Q3\n")

	for _, result := range race.QualifyingResults {
		driverName := fmt.Sprintf("%s. %s", string(result.Driver.GivenName[0]), result.Driver.FamilyName)
		q1 := result.Q1
		if q1 == "" {
			q1 = "N/A"
		}
		q2 := result.Q2
		if q2 == "" {
			q2 = "N/A"
		}
		q3 := result.Q3
		if q3 == "" {
			q3 = "N/A"
		}

		line := fmt.Sprintf("%-3s %-16s %-12s %-10s %-10s %-10s", result.Position, TruncateString(driverName, 15), TruncateString(result.Constructor.Name, 11), q1, q2, q3)
		resultsStr.WriteString(line + "\n")
	}
	resultsStr.WriteString("```")

	color := 0xFF0000 // Default F1 red
	if len(race.QualifyingResults) > 0 {
		poleSitterTeam := race.QualifyingResults[0].Constructor.Name
		if c, ok := TeamColors[poleSitterTeam]; ok {
			color = c
		}
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Qualifying Results - %s", race.RaceName),
		Description: fmt.Sprintf("**%s**\n%s, %s", race.RaceName, race.Circuit.CircuitName, race.Circuit.Location.Country),
		Color:       color,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Qualifying Times", Value: resultsStr.String()},
		},
		Footer: &discordgo.MessageEmbedFooter{Text: "Formula 1 Qualifying Results"},
	}
	s.ChannelMessageSendEmbed(channelID, embed)
}

func displaySprintResults(s *discordgo.Session, channelID string, race SprintResultsResponse_MRDatum_RaceTable_Race) {
	var resultsStr strings.Builder
	resultsStr.WriteString("```\nPos Driver               Team       Laps   Time\n")

	for _, result := range race.SprintResults {
		driverName := fmt.Sprintf("%s %s", result.Driver.GivenName, result.Driver.FamilyName)
		timeOrStatus := result.Status
		if result.Time.Time != "" {
			timeOrStatus = result.Time.Time
		} else if result.Status == "Finished" || result.Status == "finished" {
			timeOrStatus = "+1 Lap"
		}

		line := fmt.Sprintf("%-3s %-20s %-10s %-5s %-12s", result.Position, TruncateString(driverName, 18), TruncateString(result.Constructor.Name, 10), result.Laps, TruncateString(timeOrStatus, 12))
		resultsStr.WriteString(line + "\n")
	}
	resultsStr.WriteString("```")

	color := 0xFF0000 // Default F1 red
	if len(race.SprintResults) > 0 {
		winnerTeam := race.SprintResults[0].Constructor.Name
		if c, ok := TeamColors[winnerTeam]; ok {
			color = c
		}
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Sprint Race Results - %s", race.RaceName),
		Description: fmt.Sprintf("**%s**\n%s, %s", race.RaceName, race.Circuit.CircuitName, race.Circuit.Location.Country),
		Color:       color,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Sprint Race Results", Value: resultsStr.String()},
		},
		Footer: &discordgo.MessageEmbedFooter{Text: "Formula 1 Sprint Race Results"},
	}
	s.ChannelMessageSendEmbed(channelID, embed)
}

// --- Helper Functions ---

func isSessionType(arg string) bool {
	switch arg {
	case "quali", "qualifying", "sprint", "race", "practice", "fp1", "fp2", "fp3", "sprintquali", "sprintqualifying":
		return true
	default:
		return false
	}
}

func getPreviousEvent(events []Event) *Event {
	now := time.Now().UTC()
	var previousEvent *Event
	for i := len(events) - 1; i >= 0; i-- {
		event := events[i]
		if len(event.Sessions) == 0 {
			continue
		}
		// Check the date of the main race session
		raceSessionDateStr := ""
		for _, s := range event.Sessions {
			if strings.ToLower(s.Type) == "race" {
				raceSessionDateStr = s.Date
				break
			}
		}
		if raceSessionDateStr == "" {
			continue
		}

		raceTime, err := time.Parse(time.RFC3339, raceSessionDateStr)
		if err != nil {
			continue
		}

		// The first race in the past we find is our target
		if raceTime.Before(now) {
			previousEvent = &event
			return previousEvent
		}
	}
	return nil
}

func findEventByQuery(events []Event, query string) *Event {
	query = strings.ToLower(query)
	for i := len(events) - 1; i >= 0; i-- {
		event := events[i]
		eventNameLower := strings.ToLower(event.Name)
		eventLocationLower := strings.ToLower(event.Location)
		eventCircuitLower := strings.ToLower(event.Circuit)
		
		// Direct matching
		if strings.Contains(eventNameLower, query) ||
			strings.Contains(eventLocationLower, query) ||
			strings.Contains(eventCircuitLower, query) ||
			// Special case for common names
			strings.Contains(eventNameLower, "grand prix") && strings.Contains(query, "gp") {
			return &event
		}
		
		// Alternative matching for common names
		if (strings.Contains(query, "spa") && strings.Contains(eventNameLower, "belgian")) ||
			(strings.Contains(query, "hungary") && strings.Contains(eventNameLower, "hungarian")) ||
			(strings.Contains(query, "hungaroring") && strings.Contains(eventNameLower, "hungarian")) ||
			(strings.Contains(query, "budapest") && strings.Contains(eventNameLower, "hungarian")) ||
			(strings.Contains(query, "monaco") && strings.Contains(eventNameLower, "monaco")) ||
			(strings.Contains(query, "silverstone") && strings.Contains(eventNameLower, "british")) ||
			(strings.Contains(query, "monza") && strings.Contains(eventNameLower, "italian")) ||
			// Additional matching for "spa"
			(strings.Contains(query, "spa") && strings.Contains(eventCircuitLower, "spa")) {
			return &event
		}
	}
	return nil
}

func findEventRound(events []Event, targetEvent Event) int {
	for i, event := range events {
		if event.Name == targetEvent.Name {
			return i + 1 // Rounds are 1-indexed
		}
	}
	return 0
}

func TruncateString(s string, maxLen int) string {
	if len(s) > maxLen {
		if maxLen > 3 {
			return s[:maxLen-3] + "..."
		}
		return s[:maxLen]
	}
	return s
}

