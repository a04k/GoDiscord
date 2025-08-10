package admin

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
	commands.RegisterCommand("kick", Kick)
}

func Kick(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// Check if the user is an admin
	isOwner, err := utils.IsAdmin(b.Db, m.Author.ID)
	if err != nil || !isOwner {
		s.ChannelMessageSend(m.ChannelID, "You do not have permission to use this command.")
		return
	}

	// Check if the user provided at least a target user
	if len(args) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Usage: .kick <@user> [reason]")
		return
	}

	// Extract target user
	targetUser, err := utils.ExtractUserID(args[1])
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Invalid user / use. Please use a proper mention (e.g., @username).")
		return
	}

	// Set default reason
	reason := "No reason provided - .kick used"
	if len(args) >= 3 {
		reason = strings.Join(args[2:], " ")
	}

	// Kick the user
	err = s.GuildMemberDeleteWithReason(m.GuildID, targetUser, reason)
	if err != nil {
		log.Printf("Error kicking user: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred while kicking the user.")
		return
	}

	// kick confirmation message
	embed := &discordgo.MessageEmbed{
		Title:       "User Kicked",
		Description: fmt.Sprintf("Kicked user <@%s>", targetUser),
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