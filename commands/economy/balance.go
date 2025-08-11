package economy

import (
	"fmt"
	"log"

	"DiscordBot/bot"
	"DiscordBot/commands"
	"DiscordBot/utils"
	"github.com/bwmarrin/discordgo"
)

func init() {
	commands.RegisterCommand("balance", Balance, "bal", "$")
}

func Balance(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// Ensure command is used in a guild
	if m.GuildID == "" {
		return // Don't respond to DMs
	}

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

	// Ensure user exists in global users table
	_, err := b.Db.Exec(`
		INSERT INTO users (user_id, username, avatar, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			username = EXCLUDED.username,
			avatar = EXCLUDED.avatar,
			updated_at = NOW()
	`, targetUserID, m.Author.Username, m.Author.AvatarURL(""))
	if err != nil {
		log.Printf("Error upserting user: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
		return
	}

	// Ensure user exists in guild_members table
	_, err = b.Db.Exec(`
		INSERT INTO guild_members (guild_id, user_id, joined_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (guild_id, user_id) DO NOTHING
	`, m.GuildID, targetUserID)
	if err != nil {
		log.Printf("Error ensuring guild member: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
		return
	}

	// Now query their balance (will exist due to prior insert)
	var balance int
	err = b.Db.QueryRow("SELECT balance FROM guild_members WHERE guild_id = $1 AND user_id = $2", m.GuildID, targetUserID).Scan(&balance)
	if err != nil {
		log.Printf("Error querying balance: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("<@%s>'s balance: %d coins", targetUserID, balance))
}