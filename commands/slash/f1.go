package slash

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"DiscordBot/bot"

	"github.com/bwmarrin/discordgo"
)

type F1ScheduleResponse struct {
	MRData struct {
		RaceTable struct {
			Races []struct {
				RaceName string `json:"raceName"`
				Circuit  struct {
					Location struct {
						Locality string `json:"locality"`
						Country  string `json:"country"`
					} `json:"Location"`
				} `json:"Circuit"`
				Date          string `json:"date"`
				Time          string `json:"time"`
				FirstPractice struct {
					Date string `json:"date"`
					Time string `json:"time"`
				} `json:"FirstPractice"`
				SecondPractice struct {
					Date string `json:"date"`
					Time string `json:"time"`
				} `json:"SecondPractice"`
				ThirdPractice struct {
					Date string `json:"date"`
					Time string `json:"time"`
				} `json:"ThirdPractice"`
				Qualifying struct {
					Date string `json:"date"`
					Time string `json:"time"`
				} `json:"Qualifying"`
				Sprint struct {
					Date string `json:"date"`
					Time string `json:"time"`
				} `json:"Sprint,omitempty"`
			} `json:"Races"`
		} `json:"RaceTable"`
	} `json:"MRData"`
}

func F1Schedule(b *bot.Bot, s *discordgo.Session, i *discordgo.InteractionCreate) {
	resp, err := http.Get("https://ergast.com/api/f1/current.json")
	if err != nil || resp.StatusCode != 200 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Failed to fetch F1 schedule.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}
	defer resp.Body.Close()

	var data F1ScheduleResponse
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Error decoding schedule.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	now := time.Now().UTC()
	nextRace := data.MRData.RaceTable.Races[0]
	for _, race := range data.MRData.RaceTable.Races {
		parsedDate, _ := time.Parse("2006-01-02T15:04:05Z", race.Date+"T"+race.Time)
		if parsedDate.After(now) {
			nextRace = race
			break
		}
	}

	sessions := map[string]string{
		"First Practice":  nextRace.FirstPractice.Date + "T" + nextRace.FirstPractice.Time,
		"Second Practice": nextRace.SecondPractice.Date + "T" + nextRace.SecondPractice.Time,
		"Third Practice":  nextRace.ThirdPractice.Date + "T" + nextRace.ThirdPractice.Time,
		"Qualifying":      nextRace.Qualifying.Date + "T" + nextRace.Qualifying.Time,
		"Sprint":          nextRace.Sprint.Date + "T" + nextRace.Sprint.Time,
		"Race":            nextRace.Date + "T" + nextRace.Time,
	}

	// Filter out empty sessions (Sprint may not be available)
	for k, v := range sessions {
		if strings.HasSuffix(v, "T") {
			delete(sessions, k)
		}
	}

	type pair struct {
		Name string
		Time int64
	}
	var all []pair
	for name, datetime := range sessions {
		t, err := time.Parse(time.RFC3339, datetime)
		if err != nil {
			continue
		}
		all = append(all, pair{name, t.Unix()})
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].Time < all[j].Time
	})

	lines := []string{
		fmt.Sprintf("**%s** in %s, %s", nextRace.RaceName, nextRace.Circuit.Location.Locality, nextRace.Circuit.Location.Country),
	}
	for _, s := range all {
		lines = append(lines, fmt.Sprintf("• %s: <t:%d:F> (<t:%d:R>)", s.Name, s.Time, s.Time))
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Next Formula 1 Weekend",
					Description: strings.Join(lines, "\n"),
					Color:       0xE10600, // Ferrari Red :)
				},
			},
		},
	})
}
