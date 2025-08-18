package epl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
)

// FPL Data Structures
type FPLFixture struct {
	ID          int    `json:"id"`
	TeamH       int    `json:"team_h"`
	TeamA       int    `json:"team_a"`
	TeamHScore  *int   `json:"team_h_score"`
	TeamAScore  *int   `json:"team_a_score"`
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

type FPLBootstrap struct {
	Teams []struct {
		ID              int    `json:"id"`
		Name            string `json:"name"`
		ShortName       string `json:"short_name"`
		Strength        int    `json:"strength"`
		Position        int    `json:"-"`
		Played          int    `json:"played"`
		Win             int    `json:"win"`
		Draw            int    `json:"draw"`
		Loss            int    `json:"loss"`
		Points          int    `json:"points"`
		GoalsFor        int    `json:"goals_for"`
		GoalsAgainst    int    `json:"goals_against"`
		GoalDifference  int    `json:"goal_difference"`
	} `json:"teams"`
}

// Shared functions that are used across multiple command files
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

func createResultsPages(results []FPLFixture, teamNameMap, shortNameMap map[int]string, clubName string, gameweek int) []*discordgo.MessageEmbed {
	const itemsPerPage = 5
	totalResults := len(results)
	numPages := (totalResults + itemsPerPage - 1) / itemsPerPage
	
	if numPages == 0 {
		numPages = 1
	}
	
	pages := make([]*discordgo.MessageEmbed, numPages)
	
	title := "EPL Results"
	if clubName != "" {
		title = fmt.Sprintf("Results for %s", clubName)
	} else if gameweek > 0 {
		title = fmt.Sprintf("Results for Gameweek %d", gameweek)
	}
	
	for page := 0; page < numPages; page++ {
		startIndex := page * itemsPerPage
		endIndex := startIndex + itemsPerPage
		
		// Don't go beyond the actual number of results
		if endIndex > totalResults {
			endIndex = totalResults
		}
		
		embed := &discordgo.MessageEmbed{
			Title: title,
			Color: 0x3b82f6,
		}
		
		// Add results for this page
		for i := startIndex; i < endIndex; i++ {
			fixture := results[i]
			
			matchTime, err := time.Parse("2006-01-02T15:04:05Z", fixture.KickoffTime)
			if err != nil {
				continue
			}
			
			homeTeam := shortNameMap[fixture.TeamH]
			awayTeam := shortNameMap[fixture.TeamA]
			
			// Format the result with actual scores if available
			var result string
			if fixture.Finished && fixture.TeamHScore != nil && fixture.TeamAScore != nil {
				result = fmt.Sprintf("%d - %d", *fixture.TeamHScore, *fixture.TeamAScore)
			} else if fixture.Finished {
				result = "Finished"
			} else {
				result = "Not played yet"
			}
			
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name: fmt.Sprintf("GW%d: %s vs %s", fixture.Event, homeTeam, awayTeam),
				Value: fmt.Sprintf("Date: <t:%d:F>\nResult: %s", matchTime.Unix(), result),
				Inline: false,
			})
		}
		
		embed.Footer = &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Page %d of %d", page+1, numPages),
		}
		
		pages[page] = embed
	}
	
	return pages
}