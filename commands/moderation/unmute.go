package moderation

import (
	"fmt"
	"log"

	"DiscordBot/bot"
	"DiscordBot/commands"
	"DiscordBot/utils"
	"github.com/bwmarrin/discordgo"
)

func init() {
	commands.RegisterCommand("unmute", Unmute, "um")
}

func Unmute(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// Check if the user has manage messages permission
	hasManageMessages, err := utils.CheckManageMessagesPermission(s, m.GuildID, m.Author.ID)
	if err != nil {
		log.Printf("Error checking manage messages permission: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
		return
	}

	if !hasManageMessages {
		s.ChannelMessageSend(m.ChannelID, "You do not have permission to use this command.")
		return
	}

	// Check if the user provided at least a target user
	if len(args) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Usage: .unmute <@user>")
		return
	}

	// Extract target user
	targetUser, err := utils.ExtractUserID(args[1])
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Invalid user / use. Please use a proper mention (e.g., @username).")
		return
	}

	// Get the Muted role
	mutedRole, err := utils.GetMutedRole(s, m.GuildID)
	if err != nil {
		log.Printf("Error getting muted role: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred while getting the muted role.")
		return
	}

	// Remove the Muted role from the user
	err = s.GuildMemberRoleRemove(m.GuildID, targetUser, mutedRole.ID)
	if err != nil {
		log.Printf("Error unmuting user: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred while unmuting the user.")
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Unmuted user <@%s>", targetUser))
}
