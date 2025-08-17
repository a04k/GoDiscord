package epl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"DiscordBot/bot"

	"github.com/bwmarrin/discordgo"
)

func NextMatch(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) <= 1 {
		showUpcomingFixtures(b, s, m)
		return
	}

	clubName := strings.Join(args[1:], " ")
	showClubNextMatch(b, s, m, clubName)
}

func fetchAllFixtures() ([]FPLFixture, error) {
	url := "https://fantasy.premierleague.com/api/fixtures/"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	var data []FPLFixture
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}

func showUpcomingFixtures(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate) {
	fixtures, err := fetchAllFixtures()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error fetching EPL fixtures. Please try again later.")
		return
	}

	teamsData, err := getTeamsData()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error fetching team data. Please try again later.")
		return
	}

	// Create maps for team names
	teamNameMap := make(map[int]string)
	shortNameMap := make(map[int]string)
	for _, t := range teamsData {
		teamNameMap[t.ID] = t.Name
		shortNameMap[t.ID] = t.ShortName
	}

	// Find the current gameweek - first GW with unfinished matches
	currentGW := 0
	for _, fixture := range fixtures {
		if !fixture.Finished {
			if currentGW == 0 || fixture.Event < currentGW {
				currentGW = fixture.Event
			}
		}
	}
	
	// If no unfinished fixtures found, default to first GW
	if currentGW == 0 {
		currentGW = 1
	}

	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("Upcoming Premier League Fixtures (GW%d)", currentGW),
		Color: 0x3b82f6,
	}

	count := 0
	for _, match := range fixtures {
		// Only show fixtures for the current gameweek
		if match.Event != currentGW {
			continue
		}

		if count >= 10 {
			break
		}

		matchTime, err := time.Parse("2006-01-02T15:04:05Z", match.KickoffTime)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error parsing match time. Please try again later.")
			return
		}
		// Convert to Unix timestamp for Discord's timestamp formatting
		unixTimestamp := matchTime.Unix()
		timeStr := fmt.Sprintf("<t:%d:F>", unixTimestamp) // F = "Friday, 1 January 2021 12:00"

		homeTeam := shortNameMap[match.TeamH]
		awayTeam := shortNameMap[match.TeamA]

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  fmt.Sprintf("%s vs %s", homeTeam, awayTeam),
			Value: fmt.Sprintf("Date: %s", timeStr),
		})

		count++
	}

	if len(embed.Fields) == 0 {
		embed.Description = "No upcoming fixtures found for the current gameweek."
	}

	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}

func showClubNextMatch(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, clubName string) {
	fixtures, err := fetchAllFixtures()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error fetching EPL fixtures. Please try again later.")
		return
	}

	teamsData, err := getTeamsData()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error fetching team data. Please try again later.")
		return
	}

	// Create maps for team names
	teamNameMap := make(map[int]string)
	shortNameMap := make(map[int]string)
	for _, t := range teamsData {
		teamNameMap[t.ID] = t.Name
		shortNameMap[t.ID] = t.ShortName
	}

	var nextMatch *FPLFixture
	var isHomeTeam bool
	for _, match := range fixtures {
		// Only consider unfinished matches
		if match.Finished {
			continue
		}

		homeTeam := teamNameMap[match.TeamH]
		awayTeam := teamNameMap[match.TeamA]
		homeShort := shortNameMap[match.TeamH]
		awayShort := shortNameMap[match.TeamA]

		if strings.Contains(strings.ToLower(homeTeam), strings.ToLower(clubName)) ||
			strings.Contains(strings.ToLower(awayTeam), strings.ToLower(clubName)) ||
			strings.Contains(strings.ToLower(homeShort), strings.ToLower(clubName)) ||
			strings.Contains(strings.ToLower(awayShort), strings.ToLower(clubName)) {

			// If we haven't found a match yet, or this match is earlier than the current one
			if nextMatch == nil {
				nextMatch = &match
				isHomeTeam = strings.Contains(strings.ToLower(homeTeam), strings.ToLower(clubName)) ||
					strings.Contains(strings.ToLower(homeShort), strings.ToLower(clubName))
			} else {
				// Compare kickoff times to find the earliest match
				currentTime, _ := time.Parse("2006-01-02T15:04:05Z", nextMatch.KickoffTime)
				newTime, _ := time.Parse("2006-01-02T15:04:05Z", match.KickoffTime)
				if newTime.Before(currentTime) {
					nextMatch = &match
					isHomeTeam = strings.Contains(strings.ToLower(homeTeam), strings.ToLower(clubName)) ||
						strings.Contains(strings.ToLower(homeShort), strings.ToLower(clubName))
				}
			}
		}
	}

	if nextMatch == nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("No upcoming match found for club: %s", clubName))
		return
	}

	matchTime, err := time.Parse("2006-01-02T15:04:05Z", nextMatch.KickoffTime)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error parsing match time. Please try again later.")
		return
	}
	// Convert to Unix timestamp for Discord's timestamp formatting
	unixTimestamp := matchTime.Unix()
	timeStr := fmt.Sprintf("<t:%d:F>", unixTimestamp) // F = "Friday, 1 January 2021 12:00"

	venue := "Away"
	if isHomeTeam {
		venue = "Home"
	}

	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("GW%d: %s vs %s", nextMatch.Event, shortNameMap[nextMatch.TeamH], shortNameMap[nextMatch.TeamA]),
		Color: 0x3b82f6,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Date & Time",
				Value:  timeStr,
				Inline: true,
			},
			{
				Name:   "Venue",
				Value:  venue,
				Inline: true,
			},
		},
	}

	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}

func fetchFixtures() ([]FPLFixture, error) {
	url := "https://fantasy.premierleague.com/api/fixtures/"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	var data []FPLFixture
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}

func getEventsData() ([]FPLEvent, error) {
	url := "https://fantasy.premierleague.com/api/bootstrap-static/"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	var data FPLBootstrapStatic
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data.Events, nil
}

func getTeamsData() ([]FPLTeam, error) {
	url := "https://fantasy.premierleague.com/api/bootstrap-static/"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	var data FPLBootstrapStatic
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data.Teams, nil
}

type FPLFixture struct {
	ID          int    `json:"id"`
	TeamH       int    `json:"team_h"`
	TeamA       int    `json:"team_a"`
	KickoffTime string `json:"kickoff_time"`
	Event       int    `json:"event"`
	Finished    bool   `json:"finished"`
}

type FPLEvent struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Finished     bool   `json:"finished"`
	DeadlineTime string `json:"deadline_time"`
}

type FPLTeam struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"short_name"`
}

type FPLBootstrapStatic struct {
	Teams  []FPLTeam  `json:"teams"`
	Events []FPLEvent `json:"events"`
}
