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
	if len(args) == 1 {
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
		if isSessionType(firstArg) {
			sessionType = firstArg
			if len(args) > 2 {
				raceQuery = strings.ToLower(strings.Join(args[2:], " "))
			}
		} else {
			raceQuery = strings.ToLower(strings.Join(args[1:], " "))
		}
	}

	// Find the round for the given query
	round := 0
	if raceQuery != "" {
		events, err := FetchF1Events()
		if err != nil {
			log.Printf("Error fetching F1 events: %v", err)
			s.ChannelMessageSend(m.ChannelID, "Error fetching F1 schedule")
			return
		}
		event := findEventByQuery(events, raceQuery)
		if event == nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("No F1 event found matching '%s'", raceQuery))
			return
		}
		round = findEventRound(events, *event)
	}

	switch sessionType {
	case "quali", "qualifying":
		GetQualifyingResults(b, s, m, round)
	case "sprint":
		GetSprintResults(b, s, m, round)
	case "race":
		GetRaceResults(b, s, m, round)
	default:
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
	var (
		apiResponse *RaceResultsResponse
		err         error
	)

	if round == 0 {
		apiResponse, err = FetchLatestRaceResults()
	} else {
		apiResponse, err = FetchRaceResultsByRound(round)
	}

	if err != nil {
		log.Printf("Error fetching F1 race results: %v", err)
		s.ChannelMessageSend(m.ChannelID, "Error fetching F1 race results.")
		return
	}

	if apiResponse == nil || len(apiResponse.MRData.RaceTable.Races) == 0 {
		s.ChannelMessageSend(m.ChannelID, "No recent F1 race data available.")
		return
	}

	race := apiResponse.MRData.RaceTable.Races[0]
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
	var resultsStr strings.Builder
	resultsStr.WriteString("`Pos Driver               Team       Laps   Time/Status   Pts`
")

	for _, result := range race.Results {
		driverName := fmt.Sprintf("%s %s", result.Driver.GivenName, result.Driver.FamilyName)
		timeOrStatus := result.Status
		if result.Time.Time != "" {
			timeOrStatus = result.Time.Time
		}

		resultsStr.WriteString(fmt.Sprintf("`%-3s %-20s %-10s %-5s %-12s %-3s`
",
			result.Position,
			TruncateString(driverName, 18),
			TruncateString(result.Constructor.Name, 10),
			result.Laps,
			TruncateString(timeOrStatus, 12),
			result.Points))
	}

	color := 0xFF0000 // Default F1 red
	if len(race.Results) > 0 {
		winnerTeam := race.Results[0].Constructor.Name
		if c, ok := TeamColors[winnerTeam]; ok {
			color = c
		}
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("🏁 %s Results", race.RaceName),
		Description: fmt.Sprintf("**%s**
%s, %s", race.RaceName, race.Circuit.CircuitName, race.Circuit.Location.Country),
		Color:       color,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Race Results",
				Value: resultsStr.String(),
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Formula 1 Results",
		},
	}

	s.ChannelMessageSendEmbed(channelID, embed)
}

func displayQualifyingResults(s *discordgo.Session, channelID string, race QualifyingResponse_MRDatum_RaceTable_Race) {
	var resultsStr strings.Builder
	resultsStr.WriteString("`Pos Driver           Team         Q1         Q2         Q3`
")

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

		resultsStr.WriteString(fmt.Sprintf("`%-3s %-16s %-12s %-10s %-10s %-10s`
",
			result.Position,
			TruncateString(driverName, 15),
			TruncateString(result.Constructor.Name, 11),
			q1, q2, q3))
	}

	color := 0xFF0000 // Default F1 red
	if len(race.QualifyingResults) > 0 {
		poleSitterTeam := race.QualifyingResults[0].Constructor.Name
		if c, ok := TeamColors[poleSitterTeam]; ok {
			color = c
		}
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Qualifying Results - %s", race.RaceName),
		Description: fmt.Sprintf("**%s**
%s, %s", race.RaceName, race.Circuit.CircuitName, race.Circuit.Location.Country),
		Color:       color,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Qualifying Times",
				Value: resultsStr.String(),
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Formula 1 Qualifying Results",
		},
	}
	s.ChannelMessageSendEmbed(channelID, embed)
}

func displaySprintResults(s *discordgo.Session, channelID string, race SprintResultsResponse_MRDatum_RaceTable_Race) {
	var resultsStr strings.Builder
	resultsStr.WriteString("`Pos Driver               Team       Laps   Time/Status   Pts`
")

	for _, result := range race.SprintResults {
		driverName := fmt.Sprintf("%s %s", result.Driver.GivenName, result.Driver.FamilyName)
		timeOrStatus := result.Status
		if result.Time.Time != "" {
			timeOrStatus = result.Time.Time
		}

		resultsStr.WriteString(fmt.Sprintf("`%-3s %-20s %-10s %-5s %-12s %-3s`
",
			result.Position,
			TruncateString(driverName, 18),
			TruncateString(result.Constructor.Name, 10),
			result.Laps,
			TruncateString(timeOrStatus, 12),
			result.Points))
	}

	color := 0xFF0000 // Default F1 red
	if len(race.SprintResults) > 0 {
		winnerTeam := race.SprintResults[0].Constructor.Name
		if c, ok := TeamColors[winnerTeam]; ok {
			color = c
		}
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Sprint Race Results - %s", race.RaceName),
		Description: fmt.Sprintf("**%s**
%s, %s", race.RaceName, race.Circuit.CircuitName, race.Circuit.Location.Country),
		Color:       color,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Sprint Race Results",
				Value: resultsStr.String(),
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Formula 1 Sprint Race Results",
		},
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
	return nil // Should not happen if schedule is valid
}

func findEventByQuery(events []Event, query string) *Event {
	query = strings.ToLower(query)
	for i := len(events) - 1; i >= 0; i-- {
		event := events[i]
		if strings.Contains(strings.ToLower(event.Name), query) ||
			strings.Contains(strings.ToLower(event.Location), query) ||
			strings.Contains(strings.ToLower(event.Circuit), query) {
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
