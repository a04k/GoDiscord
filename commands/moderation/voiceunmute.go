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
	commands.RegisterCommand("voiceunmute", VoiceUnmute, "vum")
}

func VoiceUnmute(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// Check if the user has mute members permission
	hasMuteMembers, err := utils.CheckMuteMembersPermission(s, m.GuildID, m.Author.ID)
	if err != nil {
		log.Printf("Error checking mute members permission: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
		return
	}

	if !hasMuteMembers {
		s.ChannelMessageSend(m.ChannelID, "You do not have permission to use this command.")
		return
	}

	// Check if the user provided at least a target user
	if len(args) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Usage: .voiceunmute/vum <@user>")
		return
	}

	// Extract target user
	targetUser, err := utils.ExtractUserID(args[1])
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Invalid user / use. Please use a proper mention (e.g., @username).")
		return
	}

	// Unmute the user
	err = s.GuildMemberMute(m.GuildID, targetUser, false)
	if err != nil {
		log.Printf("Error unmuting user: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred while unmuting the user.")
		return
	}
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Unmuted user <@%s>", targetUser))
}
