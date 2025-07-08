package commands

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
)

func BTC(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	resp, err := http.Get("https://api.coingecko.com/api/v3/simple/price?ids=bitcoin&vs_currencies=usd")
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error fetching BTC price")
		return
	}
	defer resp.Body.Close()

	var result bot.BTCResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error fetching BTC price")
		return
	}
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("BTC Price: $%.2f", result.Bitcoin.USD))
}
