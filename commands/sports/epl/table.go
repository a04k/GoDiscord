package epl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
)

func EPLTable(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// Fetch Premier League table from API
	// Using the FPL API which is free and doesn't require authentication
	url := "https://fantasy.premierleague.com/api/bootstrap-static/"
	
	resp, err := http.Get(url)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error fetching EPL table. Please try again later.")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.ChannelMessageSend(m.ChannelID, "Error fetching EPL table. Please try again later.")
		return
	}

	var data FPLBootstrap
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
	sort.Slice(data.Teams, func(i, j int) bool {
		if data.Teams[i].Points != data.Teams[j].Points {
			return data.Teams[i].Points > data.Teams[j].Points // Higher points first
		}
		// If points are equal, sort by goal difference
		return data.Teams[i].GoalDifference > data.Teams[j].GoalDifference
	})

	// Update positions based on sorted order
	for i := range data.Teams {
		data.Teams[i].Position = i + 1
	}

	// Create table as a string
	var table strings.Builder
	table.WriteString("Pos | Team | P | W | D | L | GF | GA | GD | Pts\n")
	table.WriteString("--- | ---- | - | - | - | - | -- | -- | -- | ---\n")
	
	for _, team := range data.Teams {
		// Skip teams after position 20 (if any)
		if team.Position > 20 {
			continue
		}
		
		table.WriteString(fmt.Sprintf(
			"%2d | %-4s | %2d | %2d | %2d | %2d | %2d | %2d | %3d | %3d\n",
			team.Position,
			team.ShortName,
			team.Played,
			team.Win,
			team.Draw,
			team.Loss,
			team.GoalsFor,
			team.GoalsAgainst,
			team.GoalDifference,
			team.Points,
		))
	}

	embed.Description = fmt.Sprintf("```\n%s```", table.String())
	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}

// FPL Bootstrap Data Structures
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