package moderation

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
	"DiscordBot/utils"
	"DiscordBot/commands"
)

func init() {
	commands.RegisterCommand("voicemute", VoiceMute, "vm")
}

func VoiceMute(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// Check if the user is an admin or moderator
	isMod, err := utils.IsModerator(b.Db, m.Author.ID)
	if err != nil {
		log.Printf("Error checking mod status: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
		return
	}

	isOwner, err := utils.IsAdmin(b.Db, m.GuildID, m.Author.ID)
	if err != nil {
		log.Printf("Error checking admin status: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
		return
	}

	if !isMod && !isOwner {
		s.ChannelMessageSend(m.ChannelID, "You do not have permission to use this command.")
		return
	}

	// Check if the user provided at least a target user
	if len(args) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Usage: .voicemute/vm <@user> [reason]")
		return
	}

	// Extract target user
	targetUser, err := utils.ExtractUserID(args[1])
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Invalid user / use. Please use a proper mention (e.g., @username).")
		return
	}

	// Set default reason
	reason := "No reason provided - .mute used"
	if len(args) >= 3 {
		reason = strings.Join(args[2:], " ")
	}

	// Mute the user
	err = s.GuildMemberMute(m.GuildID, targetUser, true)
	if err != nil {
		log.Printf("Error muting user: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred while muting the user.")
		return
	}

	// mute confirmation message
	embed := &discordgo.MessageEmbed{
		Title:       "User Muted",
		Description: fmt.Sprintf("Muted user <@%s>", targetUser),
		Color:       0xFF0000, // Red bar on the left
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Reason",
				Value:  reason,
				Inline: false,
			},
		},
	}

	// send the embed msg
	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}