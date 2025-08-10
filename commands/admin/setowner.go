package admin

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
	"DiscordBot/commands"
)

func init() {
	commands.RegisterCommand("setowner", SetOwner)
}

func SetOwner(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// This command can only be used by the bot owner (not guild owner)
	// Check if the user is the bot owner by checking a special user ID in the env
	botOwnerID := os.Getenv("BOT_OWNER_ID")
	if botOwnerID == "" {
		// Fallback to your ID if not set
		botOwnerID = "174176306865504257"
	}
	
	if m.Author.ID != botOwnerID {
		s.ChannelMessageSend(m.ChannelID, "You are not authorized to use this command.")
		return
	}

	if len(args) < 3 {
		s.ChannelMessageSend(m.ChannelID, "Usage: .setowner <guild_id> <@user>")
		return
	}

	guildID := args[1]
	recipient := strings.TrimPrefix(strings.TrimSuffix(args[2], ">"), "<@")

	// Set the user as owner in the specified guild
	_, err := b.Db.Exec("INSERT INTO users (guild_id, user_id, is_owner) VALUES ($1, $2, TRUE) ON CONFLICT (guild_id, user_id) DO UPDATE SET is_owner = TRUE", guildID, recipient)
	if err != nil {
		log.Printf("Error setting owner: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Set <@%s> as owner in guild %s.", recipient, guildID))
}