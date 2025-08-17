package epl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"

	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
	"DiscordBot/commands/help"
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

	// Check if there's any data to show
	hasData := false
	for _, team := range data.Teams {
		if team.Played > 0 || team.Points > 0 {
			hasData = true
			break
		}
	}

	// If no data, show a message
	if !hasData {
		embed := &discordgo.MessageEmbed{
			Title:       "Premier League Table",
			Description: "The Premier League season has not started yet. Check back after the first matches!",
			Color:       0x3b82f6,
		}
		s.ChannelMessageSendEmbed(m.ChannelID, embed)
		return
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

	// Create paginated pages
	pages := createEPLTablePages(data.Teams)

	// Send the first page
	msg, err := s.ChannelMessageSendEmbed(m.ChannelID, pages[0])
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error sending EPL table.")
		return
	}

	// If there's only one page, no need for pagination
	if len(pages) <= 1 {
		return
	}

	// Add reactions for pagination
	err = s.MessageReactionAdd(msg.ChannelID, msg.ID, "⬅️")
	if err != nil {
		// Ignore error, not critical
	}

	err = s.MessageReactionAdd(msg.ChannelID, msg.ID, "➡️")
	if err != nil {
		// Ignore error, not critical
	}

	// Store pagination state
	state := &help.PaginationState{
		UserID:      m.Author.ID,
		MessageID:   msg.ID,
		CurrentPage: 0,
		TotalPages:  len(pages),
		Pages:       pages,
		GuildID:     m.GuildID,
	}

	help.GetPaginationManager().AddState(state)
}

func createEPLTablePages(teams []struct {
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
}) []*discordgo.MessageEmbed {
	const teamsPerPage = 10
	totalTeams := len(teams)
	numPages := (totalTeams + teamsPerPage - 1) / teamsPerPage
	
	if numPages == 0 {
		numPages = 1
	}
	
	pages := make([]*discordgo.MessageEmbed, numPages)
	
	for page := 0; page < numPages; page++ {
		startIndex := page * teamsPerPage
		endIndex := startIndex + teamsPerPage
		
		// Don't go beyond the actual number of teams
		if endIndex > totalTeams {
			endIndex = totalTeams
		}
		
		embed := &discordgo.MessageEmbed{
			Title: "Premier League Table",
			Color: 0x3b82f6,
		}
		
		// Add teams for this page
		for i := startIndex; i < endIndex; i++ {
			team := teams[i]
			
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name: fmt.Sprintf("#%d %s", team.Position, team.ShortName),
				Value: fmt.Sprintf("Played: %d\nWon: %d | Drawn: %d | Lost: %d\nGoals For: %d | Against: %d | Diff: %d\nPoints: %d", 
					team.Played, team.Win, team.Draw, team.Loss,
					team.GoalsFor, team.GoalsAgainst, team.GoalDifference, team.Points),
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