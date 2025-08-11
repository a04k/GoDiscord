package epl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
	"DiscordBot/commands"
)

func init() {
	commands.RegisterCommand("nextmatch", NextMatch)
}

func NextMatch(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// If no arguments, show all upcoming fixtures for current gameweek
	if len(args) <= 1 {
		showUpcomingFixtures(b, s, m)
		return
	}

	// If club name is provided, show next match for that club
	clubName := strings.Join(args[1:], " ")
	showClubNextMatch(b, s, m, clubName)
}

func showUpcomingFixtures(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate) {
	// Fetch upcoming fixtures from API
	url := "https://api.football-data.org/v4/competitions/PL/matches?status=SCHEDULED"
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error creating request for EPL fixtures. Please try again later.")
		return
	}
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error fetching EPL fixtures. Please try again later.")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.ChannelMessageSend(m.ChannelID, "Error fetching EPL fixtures. Please try again later.")
		return
	}

	var data EPLMatches
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error parsing EPL fixtures data. Please try again later.")
		return
	}

	// Create embed
	embed := &discordgo.MessageEmbed{
		Title: "Upcoming Premier League Fixtures",
		Color: 0x3b82f6,
	}

	// Show only the next 10 fixtures
	count := 0
	for _, match := range data.Matches {
		if count >= 10 {
			break
		}
		
		// Convert UTC time to user's local time
		matchTime, err := time.Parse(time.RFC3339, match.UTCDate)
		if err != nil {
			// If parsing fails, use the original string
			matchTime = time.Now()
		}
		
		// Format time in a user-friendly way
		timeStr := matchTime.Format("Mon, Jan 2, 15:04 MST")
		
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name: fmt.Sprintf("%s vs %s", match.HomeTeam.Name, match.AwayTeam.Name),
			Value: fmt.Sprintf("Date: %s\nVenue: %s", timeStr, match.Venue),
			Inline: false,
		})
		
		count++
	}

	if len(embed.Fields) == 0 {
		embed.Description = "No upcoming fixtures found."
	}

	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}

func showClubNextMatch(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, clubName string) {
	// Fetch upcoming fixtures from API
	url := "https://api.football-data.org/v4/competitions/PL/matches?status=SCHEDULED"
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error creating request for EPL fixtures. Please try again later.")
		return
	}
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error fetching EPL fixtures. Please try again later.")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.ChannelMessageSend(m.ChannelID, "Error fetching EPL fixtures. Please try again later.")
		return
	}

	var data EPLMatches
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error parsing EPL fixtures data. Please try again later.")
		return
	}

	// Find the next match for the specified club
	var nextMatch *EPLMatch
	for _, match := range data.Matches {
		// Check if the club is playing in this match
		if strings.Contains(strings.ToLower(match.HomeTeam.Name), strings.ToLower(clubName)) ||
		   strings.Contains(strings.ToLower(match.AwayTeam.Name), strings.ToLower(clubName)) {
			nextMatch = &match
			break
		}
	}

	// If no match found for the club
	if nextMatch == nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("No upcoming match found for club: %s", clubName))
		return
	}

	// Convert UTC time to user's local time
	matchTime, err := time.Parse(time.RFC3339, nextMatch.UTCDate)
	if err != nil {
		// If parsing fails, use the original string
		matchTime = time.Now()
	}
	
	// Format time in a user-friendly way
	timeStr := matchTime.Format("Mon, Jan 2, 15:04 MST")

	// Create embed
	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("%s vs %s", nextMatch.HomeTeam.Name, nextMatch.AwayTeam.Name),
		Color: 0x3b82f6,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Date & Time",
				Value:  timeStr,
				Inline: true,
			},
			{
				Name:   "Venue",
				Value:  nextMatch.Venue,
				Inline: true,
			},
		},
	}

	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}

// EPL Matches Data Structure
type EPLMatches struct {
	Matches []EPLMatch `json:"matches"`
}

type EPLMatch struct {
	ID        int    `json:"id"`
	UTCDate   string `json:"utcDate"`
	Status    string `json:"status"`
	Venue     string `json:"venue"`
	HomeTeam  struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"homeTeam"`
	AwayTeam struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"awayTeam"`
}