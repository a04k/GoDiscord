package epl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"DiscordBot/bot"
	"DiscordBot/commands/help"

	"github.com/bwmarrin/discordgo"
)

func ShowMatchups(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 3 {
		s.ChannelMessageSend(m.ChannelID, "Usage: .epl matchups <club1> <club2> [season]\nExample: .epl matchups arsenal chelsea")
		return
	}

	club1 := args[1]
	club2 := args[2]
	
	// Optional season parameter
	season := ""
	if len(args) > 3 {
		season = args[3]
	}

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

	// Find team IDs for the specified clubs
	var club1ID, club2ID int
	club1Found := false
	club2Found := false

	for _, team := range teamsData {
		if !club1Found && (strings.Contains(strings.ToLower(team.Name), strings.ToLower(club1)) || 
			strings.Contains(strings.ToLower(team.ShortName), strings.ToLower(club1))) {
			club1ID = team.ID
			club1Found = true
		}
		
		if !club2Found && (strings.Contains(strings.ToLower(team.Name), strings.ToLower(club2)) || 
			strings.Contains(strings.ToLower(team.ShortName), strings.ToLower(club2))) {
			club2ID = team.ID
			club2Found = true
		}
		
		if club1Found && club2Found {
			break
		}
	}

	if !club1Found || !club2Found {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Could not find one or both clubs: %s, %s", club1, club2))
		return
	}

	// Filter fixtures between these two clubs
	var matchups []FPLFixture
	for _, fixture := range fixtures {
		// Check if this fixture is between the two clubs
		if (fixture.TeamH == club1ID && fixture.TeamA == club2ID) || 
		   (fixture.TeamH == club2ID && fixture.TeamA == club1ID) {
			// If season is specified, we could filter by it, but for now include all
			matchups = append(matchups, fixture)
		}
	}

	if len(matchups) == 0 {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("No matchups found between %s and %s", club1, club2))
		return
	}

	// Sort matchups by kickoff time (newest first)
	sort.Slice(matchups, func(i, j int) bool {
		timeI, _ := time.Parse("2006-01-02T15:04:05Z", matchups[i].KickoffTime)
		timeJ, _ := time.Parse("2006-01-02T15:04:05Z", matchups[j].KickoffTime)
		return timeI.After(timeJ)
	})

	// Create paginated pages
	pages := createMatchupsPages(matchups, teamNameMap, shortNameMap, club1, club2, season)

	// Send the first page
	msg, err := s.ChannelMessageSendEmbed(m.ChannelID, pages[0])
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error sending matchups.")
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

func createMatchupsPages(matchups []FPLFixture, teamNameMap, shortNameMap map[int]string, club1, club2, season string) []*discordgo.MessageEmbed {
	const itemsPerPage = 5
	totalMatchups := len(matchups)
	numPages := (totalMatchups + itemsPerPage - 1) / itemsPerPage
	
	if numPages == 0 {
		numPages = 1
	}
	
	pages := make([]*discordgo.MessageEmbed, numPages)
	
	title := fmt.Sprintf("Matchups: %s vs %s", club1, club2)
	if season != "" {
		title = fmt.Sprintf("Matchups: %s vs %s (%s)", club1, club2, season)
	}
	
	for page := 0; page < numPages; page++ {
		startIndex := page * itemsPerPage
		endIndex := startIndex + itemsPerPage
		
		// Don't go beyond the actual number of matchups
		if endIndex > totalMatchups {
			endIndex = totalMatchups
		}
		
		embed := &discordgo.MessageEmbed{
			Title: title,
			Color: 0x3b82f6,
		}
		
		// Add matchups for this page
		for i := startIndex; i < endIndex; i++ {
			fixture := matchups[i]
			
			matchTime, err := time.Parse("2006-01-02T15:04:05Z", fixture.KickoffTime)
			if err != nil {
				continue
			}
			
			homeTeam := shortNameMap[fixture.TeamH]
			awayTeam := shortNameMap[fixture.TeamA]
			
			// Format the match result if finished
			result := "Not played yet"
			if fixture.Finished {
				// Note: The FPL API doesn't provide scores in the fixtures endpoint
				// We would need to use a different API or data source for actual scores
				result = "Finished"
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