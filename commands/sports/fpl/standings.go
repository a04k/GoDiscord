package fpl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
	"DiscordBot/commands/help"
)

func FPLStandings(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// Get the FPL league ID for this guild
	var leagueID int64
	err := b.Db.QueryRow("SELECT fpl_league_id FROM guilds WHERE guild_id = $1", m.GuildID).Scan(&leagueID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "No FPL league configured for this server. Please contact an admin to set one up.")
		return
	}

	// If no league ID is set
	if leagueID == 0 {
		s.ChannelMessageSend(m.ChannelID, "No FPL league configured for this server. Please contact an admin to set one up.")
		return
	}

	// Fetch standings from FPL API
	url := fmt.Sprintf("https://fantasy.premierleague.com/api/leagues-classic/%d/standings/", leagueID)
	resp, err := http.Get(url)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error fetching FPL standings. Please try again later.")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.ChannelMessageSend(m.ChannelID, "Error fetching FPL standings. Please try again later.")
		return
	}

	var data FPLLeague
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error parsing FPL standings data. Please try again later.")
		return
	}

	// Check if there are any results
	if len(data.Standings.Results) == 0 {
		embed := &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("FPL League Standings - %s", data.League.Name),
			Description: fmt.Sprintf("Total Players: %d\n\nNo standings available yet. The league may not have started or there might be no players.", data.Standings.TotalPlayers),
			Color:       0x3b82f6,
		}
		s.ChannelMessageSendEmbed(m.ChannelID, embed)
		return
	}

	// Sort standings by rank
	sort.Slice(data.Standings.Results, func(i, j int) bool {
		return data.Standings.Results[i].Rank < data.Standings.Results[j].Rank
	})

	// Create paginated pages - increase limit to top 100
	pages := createFPLStandingsPages(data)

	// Send the first page
	msg, err := s.ChannelMessageSendEmbed(m.ChannelID, pages[0])
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error sending FPL standings.")
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

func createFPLStandingsPages(data FPLLeague) []*discordgo.MessageEmbed {
	const itemsPerPage = 15
	const maxItems = 100
	totalItems := len(data.Standings.Results)
	
	// Limit to maxItems
	if totalItems > maxItems {
		totalItems = maxItems
	}
	
	// Calculate number of pages needed
	numPages := (totalItems + itemsPerPage - 1) / itemsPerPage
	if numPages == 0 {
		numPages = 1
	}
	
	pages := make([]*discordgo.MessageEmbed, numPages)
	
	for page := 0; page < numPages; page++ {
		startIndex := page * itemsPerPage
		endIndex := startIndex + itemsPerPage
		
		// Don't go beyond the actual number of items
		if endIndex > totalItems {
			endIndex = totalItems
		}
		
		// Create a nice formatted table
		var table strings.Builder
		table.WriteString("```\n")
		table.WriteString("Rank  Team Manager          Entry Name         Points\n")
		table.WriteString("----------------------------------------------------\n")
		
		for i := startIndex; i < endIndex; i++ {
			player := data.Standings.Results[i]
			
			// Add medal emojis for top 3
			rankStr := fmt.Sprintf("%d", player.Rank)
			switch player.Rank {
			case 1:
				rankStr = "🥇1"
			case 2:
				rankStr = "🥈2"
			case 3:
				rankStr = "🥉3"
			}
			
			// Format player name and entry name
			playerName := player.PlayerName
			if len(playerName) > 18 {
				playerName = playerName[:15] + "..."
			}
			
			entryName := player.EntryName
			if len(entryName) > 18 {
				entryName = entryName[:15] + "..."
			}
			
			table.WriteString(fmt.Sprintf(
				"%-4s  %-18s %-18s %6d\n",
				rankStr,
				playerName,
				entryName,
				player.Total,
			))
		}
		
		table.WriteString("```")
		
		embed := &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("FPL League Standings - %s", data.League.Name),
			Description: fmt.Sprintf("Total Players: %d\n%s", data.Standings.TotalPlayers, table.String()),
			Color:       0x3b82f6,
			Footer: &discordgo.MessageEmbedFooter{
				Text: fmt.Sprintf("Page %d of %d", page+1, numPages),
			},
		}
		
		pages[page] = embed
	}
	
	return pages
}

// FPL Data Structures
type FPLLeague struct {
	League struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"league"`
	Standings struct {
		HasNext     bool `json:"has_next"`
		Page        int  `json:"page"`
		TotalPages  int  `json:"total_pages"`
		TotalPlayers int `json:"total_players"`
		Results     []struct {
			ID            int    `json:"id"`
			EventTotal    int    `json:"event_total"`
			PlayerName    string `json:"player_name"`
			Rank          int    `json:"rank"`
			LastRank      int    `json:"last_rank"`
			RankSort      int    `json:"rank_sort"`
			Total         int    `json:"total"`
			EntryID       int    `json:"entry"`
			EntryName     string `json:"entry_name"`
			HasCup        bool   `json:"has_cup"`
			CupRank       int    `json:"cup_rank"`
			CupLeague     int    `json:"cup_league"`
			CupLeagueName string `json:"cup_league_name"`
			CupMatched    bool   `json:"cup_matched"`
			CupPoints     int    `json:"cup_points"`
			CupRankSort   int    `json:"cup_rank_sort"`
			CupTotal      int    `json:"cup_total"`
			Draws         int    `json:"draws"`
			Losses        int    `json:"losses"`
			PointsFor     int    `json:"points_for"`
			PointsAgainst int    `json:"points_against"`
			PointsTotal   int    `json:"points_total"`
			Wins          int    `json:"wins"`
		} `json:"results"`
	} `json:"standings"`
}