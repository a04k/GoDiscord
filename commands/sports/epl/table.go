package epl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
	"DiscordBot/commands"
)

func init() {
	commands.RegisterCommand("epltable", EPLTable)
}

func EPLTable(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// Fetch Premier League table from API
	// Using a common football API endpoint
	url := "https://api.football-data.org/v4/competitions/PL/standings"
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error creating request for EPL table. Please try again later.")
		return
	}
	
	// Note: This would require an API key in a real implementation
	// req.Header.Set("X-Response-Control", "minified")
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error fetching EPL table. Please try again later.")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.ChannelMessageSend(m.ChannelID, "Error fetching EPL table. Please try again later.")
		return
	}

	var data EPLStandings
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error parsing EPL table data. Please try again later.")
		return
	}

	// Create embed
	embed := &discordgo.MessageEmbed{
		Title: "Premier League Table",
		Color: 0x3b82f6,
	}

	// Sort teams by position
	sort.Slice(data.Standings[0].Table, func(i, j int) bool {
		return data.Standings[0].Table[i].Position < data.Standings[0].Table[j].Position
	})

	// Create table as a string
	var table strings.Builder
	table.WriteString("Pos | Team | P | W | D | L | GF | GA | GD | Pts\n")
	table.WriteString("--- | ---- | - | - | - | - | -- | -- | -- | ---\n")
	
	for _, team := range data.Standings[0].Table {
		// Skip teams after position 20 (if any)
		if team.Position > 20 {
			continue
		}
		
		table.WriteString(fmt.Sprintf(
			"%2d | %s | %d | %d | %d | %d | %d | %d | %d | %d\n",
			team.Position,
			team.Team.ShortName,
			team.PlayedGames,
			team.Won,
			team.Draw,
			team.Lost,
			team.GoalsFor,
			team.GoalsAgainst,
			team.GoalDifference,
			team.Points,
		))
	}

	embed.Description = fmt.Sprintf("```\n%s```", table.String())
	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}

// EPL Data Structures
type EPLStandings struct {
	Competition struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Code string `json:"code"`
	} `json:"competition"`
	Standings []struct {
		Stage string `json:"stage"`
		Type  string `json:"type"`
		Table []struct {
			Position     int `json:"position"`
			Team         struct {
				ID       int    `json:"id"`
				Name     string `json:"name"`
				ShortName string `json:"shortName"`
				Tla      string `json:"tla"`
			} `json:"team"`
			PlayedGames   int `json:"playedGames"`
			Won           int `json:"won"`
			Draw          int `json:"draw"`
			Lost          int `json:"lost"`
			Points        int `json:"points"`
			GoalsFor      int `json:"goalsFor"`
			GoalsAgainst  int `json:"goalsAgainst"`
			GoalDifference int `json:"goalDifference"`
		} `json:"table"`
	} `json:"standings"`
}