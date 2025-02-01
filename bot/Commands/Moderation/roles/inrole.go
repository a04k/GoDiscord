package commands

import (
	"github.com/bwmarrin/discordgo"
	"utils"
)

func HandleInrole(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 1 {
		s.ChannelMessageSend(m.ChannelID, "Please provide a role ID.")
		return
	}

	roleID := args[0]

	// Get the guild (server) info
	guild, err := s.State.Guild(m.GuildID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Failed to fetch server details.")
		return
	}

	// Find members with the role
	members := utils.GetMembersInRole(roleID)
	

	// Find the role details
	var role *discordgo.Role
	for _, r := range guild.Roles {
		if r.ID == roleID {
			role = r
			break
		}
	}

	if role == nil {
		s.ChannelMessageSend(m.ChannelID, "Role not found.")
		return
	}

	// Create and send the embed
	embed := createRoleMembersEmbed(role, members)
	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}
