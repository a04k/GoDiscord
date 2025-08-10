package economy

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
	"DiscordBot/utils"
	"DiscordBot/commands"
)

func init() {
	commands.RegisterCommand("balance", Balance, "bal")
}

func Balance(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	var targetUserID string
	// Check if a mention is provided & validate
	if len(args) >= 2 {
		// Extract and validate the target user mention
		recipientMention := args[1]
		var err error
		targetUserID, err = utils.ExtractUserID(recipientMention)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Invalid mention / use. Please use a proper mention (e.g., @username).")
			return
		}

		// Check if mentioned user is in the guild
		member, err := s.GuildMember(m.GuildID, targetUserID)
		if err != nil || member == nil {
			s.ChannelMessageSend(m.ChannelID, "mentioned user is not in this server.")
			return
		}
	} else {
		// If no mention is provided, use the author's ID
		targetUserID = m.Author.ID
	}

	// register the user if they don't exist
	_, err := b.Db.Exec(`
        INSERT INTO users (user_id, balance)
        VALUES ($1, $2)
        ON CONFLICT (user_id) DO NOTHING`,
		targetUserID, 0)
	if err != nil {
		log.Printf("Error in user registration: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
		return
	}

	// Now query their balance (will exist due to prior insert)
	var balance int
	err = b.Db.QueryRow("SELECT balance FROM users WHERE user_id = $1", targetUserID).Scan(&balance)
	if err != nil {
		log.Printf("Error querying balance: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("<@%s>'s balance: %d coins", targetUserID, balance))
}