package fpl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"

	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
	"DiscordBot/commands"
)

func init() {
	commands.RegisterCommand("fplstandings", FPLStandings, "fpl")
}

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

	// Create embed
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("FPL League Standings - %s", data.League.Name),
		Description: fmt.Sprintf("Total Players: %d", data.Standings.TotalPlayers),
		Color:       0x3b82f6,
	}

	// Check if there are any results
	if len(data.Standings.Results) == 0 {
		embed.Description = fmt.Sprintf("Total Players: %d\n\nNo standings available yet. The league may not have started or there might be no players.", data.Standings.TotalPlayers)
		s.ChannelMessageSendEmbed(m.ChannelID, embed)
		return
	}

	// Sort standings by rank
	sort.Slice(data.Standings.Results, func(i, j int) bool {
		return data.Standings.Results[i].Rank < data.Standings.Results[j].Rank
	})

	// Add top 15 players
	count := 0
	for _, player := range data.Standings.Results {
		if count >= 15 {
			break
		}
		
		// Format the entry name to fit better
		entryName := player.EntryName
		if len(entryName) > 20 {
			entryName = entryName[:17] + "..."
		}
		
		// Format the player name to fit better
		playerName := player.PlayerName
		if len(playerName) > 15 {
			playerName = playerName[:12] + "..."
		}
		
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  fmt.Sprintf("#%d %s", player.Rank, entryName),
			Value: fmt.Sprintf("Player: %s\nPoints: %d", playerName, player.Total),
			Inline: false,
		})
		
		count++
	}

	s.ChannelMessageSendEmbed(m.ChannelID, embed)
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