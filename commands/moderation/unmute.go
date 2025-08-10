package moderation

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
	"DiscordBot/utils"
	"DiscordBot/commands"
)

func init() {
	commands.RegisterCommand("unmute", Unmute, "um")
}

func Unmute(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// Check if the user is an admin or moderator
	isMod, err := utils.IsModerator(b.Db, m.Author.ID)
	if err != nil {
		log.Printf("Error checking mod status: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
		return
	}

	isOwner, err := utils.IsAdmin(b.Db, m.Author.ID)
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