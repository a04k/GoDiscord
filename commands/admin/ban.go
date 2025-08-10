package admin

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"DiscordBot/bot"
	"DiscordBot/commands"
	"DiscordBot/utils"
	"github.com/bwmarrin/discordgo"
)

func init() {
	commands.RegisterCommand("ban", Ban)
}

func Ban(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// Check if the user is an admin
	isAdmin, err := b.IsAdmin(m.GuildID, m.Author.ID)
	if err != nil || !isAdmin {
		s.ChannelMessageSend(m.ChannelID, "You do not have permission to use this command.")
		return
	}

	// Check if the user provided at least a target user
	if len(args) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Usage: .ban @user [reason] [days_to_delete_messages]")
		return
	}

	// Extract target user
	targetUser, err := utils.ExtractUserID(args[1])
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Invalid user / use. Please use a proper mention (e.g., @username).")
		return
	}

	// Set default reason
	reason := "No reason provided - .ban used"
	if len(args) >= 3 {
		reason = strings.Join(args[2:], " ")
	}

	// Default message deletion days = zero
	msgDelDays := 0
	if len(args) >= 4 {
		days, err := strconv.Atoi(args[3])
		if err == nil {
			msgDelDays = days
		}
	}

	// Throw the ban hammer
	err = s.GuildBanCreateWithReason(m.GuildID, targetUser, reason, msgDelDays)
	if err != nil {
		log.Printf("Error banning user: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred while banning the user.")
		return
	}

	// ban confirmation message
	embed := &discordgo.MessageEmbed{
		Title:       "User Banned",
		Description: fmt.Sprintf("Banned user <@%s>", targetUser),
		Color:       0xFF0000, // Red bar on the left
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Reason",
				Value:  reason,
				Inline: false,
			},
			{
				Name:   "Message Deletion",
				Value:  fmt.Sprintf("%d days", msgDelDays),
				Inline: false,
			},
		},
	}

	// send the embed msg
	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}
