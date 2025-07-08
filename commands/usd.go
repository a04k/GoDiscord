package commands

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
)

func USD(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// default amount = 1
	amount := 1.0

	// Check if the user provided an amount
	if len(args) > 1 {
		var err error
		amount, err = strconv.ParseFloat(args[1], 64)
		if err != nil || amount <= 0 {
			s.ChannelMessageSend(m.ChannelID, "Invalid amount. Usage: .usd [amount]")
			return
		}
	}

	rate, err := getUSDEGP()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error fetching exchange rate")
		return
	}

	equivalent := amount * rate

	if amount == 1 {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("1 USD = %.2f EGP", rate))
	} else {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%.2f USD = %.2f EGP", amount, equivalent))
	}
}

func getUSDEGP() (float64, error) {
	// scrape town
	// URL for USD to EGP on Google Finance
	url := "https://www.google.com/finance/quote/USD-EGP"

	// Fetch the HTML content
	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to fetch URL: status code %d", resp.StatusCode)
	}

	// Parse the HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to parse HTML: %v", err)
	}

	// Find the exchange rate element
	rateText := doc.Find(".YMlKec.fxKbKc").First().Text()
	if rateText == "" {
		return 0, fmt.Errorf("failed to find exchange rate element")
	}

	// Clean and convert the rate text to a float
	rateText = strings.TrimSpace(rateText)
	rate, err := strconv.ParseFloat(rateText, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse exchange rate: %v", err)
	}

	return rate, nil
}
