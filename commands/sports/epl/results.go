package epl

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"DiscordBot/bot"
	"DiscordBot/commands/help"

	"github.com/bwmarrin/discordgo"
)

func ShowResults(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Usage: .epl results <club_name> or .epl results <gameweek>")
		return
	}

	// Check if the argument is a gameweek number
	gameweek := 0
	clubName := ""
	
	_, err := fmt.Sscanf(args[1], "%d", &gameweek)
	if err != nil || gameweek <= 0 {
		// It's not a valid gameweek number, treat as club name
		clubName = args[1]
		gameweek = 0
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

	// If club name is provided, find the team ID
	var clubID int
	clubFound := false
	if clubName != "" {
		for _, team := range teamsData {
			if strings.Contains(strings.ToLower(team.Name), strings.ToLower(clubName)) || 
				strings.Contains(strings.ToLower(team.ShortName), strings.ToLower(clubName)) {
				clubID = team.ID
				clubFound = true
				break
			}
		}
		
		if !clubFound {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Could not find club: %s", clubName))
			return
		}
	}

	// Filter fixtures based on club or gameweek
	var results []FPLFixture
	now := time.Now().UTC()
	
	for _, fixture := range fixtures {
		// Only consider finished fixtures
		if !fixture.Finished {
			continue
		}
		
		// Parse match time
		matchTime, err := time.Parse("2006-01-02T15:04:05Z", fixture.KickoffTime)
		if err != nil {
			continue
		}
		
		// Don't show future matches
		if matchTime.After(now) {
			continue
		}
		
		if clubName != "" {
			// Filter by club
			if fixture.TeamH == clubID || fixture.TeamA == clubID {
				results = append(results, fixture)
			}
		} else if gameweek > 0 {
			// Filter by gameweek
			if fixture.Event == gameweek {
				results = append(results, fixture)
			}
		}
	}

	if len(results) == 0 {
		if clubName != "" {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("No results found for club: %s", clubName))
		} else {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("No results found for gameweek: %d", gameweek))
		}
		return
	}

	// Sort results by kickoff time (newest first)
	sort.Slice(results, func(i, j int) bool {
		timeI, _ := time.Parse("2006-01-02T15:04:05Z", results[i].KickoffTime)
		timeJ, _ := time.Parse("2006-01-02T15:04:05Z", results[j].KickoffTime)
		return timeI.After(timeJ)
	})

	// Create paginated pages
	pages := createResultsPages(results, teamNameMap, shortNameMap, clubName, gameweek)

	// Send the first page
	msg, err := s.ChannelMessageSendEmbed(m.ChannelID, pages[0])
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error sending results.")
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