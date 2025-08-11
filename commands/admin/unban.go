package admin

import (
	"fmt"
	"log"
	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
	"DiscordBot/utils"
	"DiscordBot/commands"
)

func init() {
	commands.RegisterCommand("unban", Unban)
}

func Unban(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Please specify a user ID to unban.")
		return
	}

	// Check if the user has ban members permission
	hasBanPerm, err := utils.CheckBanMembersPermission(s, m.GuildID, m.Author.ID)
	if err != nil {
		log.Printf("Error checking ban members permission: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
		return
	}

	if !hasBanPerm {
		s.ChannelMessageSend(m.ChannelID, "You do not have permission to use this command.")
		return
	}

	targetUser := args[1]

	// exec remove ban => guildbandelete
	err = s.GuildBanDelete(m.GuildID, targetUser)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Failed to unban user: %s", err))
		return
	}
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Unbanned user <@%s>", targetUser))
}