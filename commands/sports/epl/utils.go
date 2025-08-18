package epl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
)

// --- Shared helpers ---

// fetchAllFixtures fetches all fixtures from FPL API.
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

// getTeamsData fetches Teams from bootstrap-static.
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

// getEventsData fetches Events (gameweeks) from bootstrap-static.
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

// createResultsPages builds paginated Discord embeds for a set of fixtures.
// teamNameMap and shortNameMap map team ID -> name/short_name.
func createResultsPages(
	results []FPLFixture,
	teamNameMap, shortNameMap map[int]string,
	clubName string,
	gameweek int,
) []*discordgo.MessageEmbed {
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
		if endIndex > totalResults {
			endIndex = totalResults
		}

		embed := &discordgo.MessageEmbed{
			Title: title,
			Color: 0x3b82f6, // tailwind blue-500
		}

		for i := startIndex; i < endIndex; i++ {
			fixture := results[i]

			// KickoffTime is ISO8601; in FPL it’s typically Zulu time.
			matchTime, err := time.Parse("2006-01-02T15:04:05Z", fixture.KickoffTime)
			if err != nil {
				// try RFC3339 as a fallback (some seasons include millis / offsets)
				if t2, err2 := time.Parse(time.RFC3339, fixture.KickoffTime); err2 == nil {
					matchTime = t2
				} else {
					// If it still fails, skip this row gracefully.
					continue
				}
			}

			homeTeam := shortNameMap[fixture.TeamH]
			if homeTeam == "" {
				homeTeam = teamNameMap[fixture.TeamH]
			}
			awayTeam := shortNameMap[fixture.TeamA]
			if awayTeam == "" {
				awayTeam = teamNameMap[fixture.TeamA]
			}

			// Build result text.
			var result string
			if fixture.Finished && fixture.TeamHScore != nil && fixture.TeamAScore != nil {
				result = fmt.Sprintf("%d - %d", *fixture.TeamHScore, *fixture.TeamAScore)
			} else if fixture.Finished {
				result = "Finished"
			} else {
				result = "Not played yet"
			}

			// GW label (0 means unknown/ didn't start)
			gw := fixture.Event
			gwLabel := "GW?"
			if gw > 0 {
				gwLabel = fmt.Sprintf("GW%d", gw)
			}

			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   fmt.Sprintf("%s: %s vs %s", gwLabel, homeTeam, awayTeam),
				Value:  fmt.Sprintf("Date: <t:%d:F>\nResult: %s", matchTime.Unix(), result),
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