package commands

import (
	"fmt"
	"log"
	"strings"
	"github.com/bwmarrin/discordgo"
	"utils"
)

func HandleSetrole(s *discordgo.Session, m *discordgo.MessageCreate, args []string) error {
	if len(args) < 3 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: .setrole/sr <@user> <role name or mention>")
		return err
	}

	userMention := args[1]
	roleName := strings.Join(args[2:], " ")

	// Get user ID from mention
	userID, err := utils.extractUserID(userMention)
	if err != nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Invalid user mention. Example: .addrole @User RoleName")
		return err
	}

	guild, err := s.Guild(m.GuildID)
	if err != nil {
		log.Printf("Error fetching guild: %v", err)
		_, err := s.ChannelMessageSend(m.ChannelID, "An error occurred while fetching server information.")
		return err
	}

	targetRole := utils.findRole(guild.Roles, roleName)
	if targetRole == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Role '%s' not found.", roleName))
		return err
	}

	executor, err := s.GuildMember(m.GuildID, m.Author.ID)
	if err != nil {
		log.Printf("Error fetching executor: %v", err)
		_, err := s.ChannelMessageSend(m.ChannelID, "An error occurred while verifying permissions.")
		return err
	}

	executorHighestPosition := utils.getHighestRolePosition(executor.Roles, guild.Roles)
	if targetRole.Position >= executorHighestPosition {
		_, err := s.ChannelMessageSend(m.ChannelID, "You cannot assign a role equal to or higher than your highest role.")
		return err
	}

	err = s.GuildMemberRoleAdd(m.GuildID, userID, targetRole.ID)
	if err != nil {
		log.Printf("Error adding role: %v", err)
		_, err := s.ChannelMessageSend(m.ChannelID, "An error occurred while adding the role to the user.")
		return err
	}

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Role '%s' added to user <@%s> successfully!", targetRole.Name, userID))
	return err
}

// HandleInRole handles the .inrole command
func HandleInRole(s *discordgo.Session, m *discordgo.MessageCreate, args []string) error {
	if len(args) < 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: .inrole <role name or mention>")
		return err
	}

	input := strings.TrimSpace(strings.Join(args[1:], " "))
	log.Printf("Input received: '%s'", input)

	guild, err := s.Guild(m.GuildID)
	if err != nil {
		log.Printf("Error fetching guild: %v", err)
		_, err := s.ChannelMessageSend(m.ChannelID, "An error occurred while fetching server information.")
		return err
	}

	targetRole := utils.findRole(guild.Roles, input)
	if targetRole == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Role '%s' not found.", input))
		return err
	}

	log.Printf("Role found: '%s' (ID: %s)", targetRole.Name, targetRole.ID)

	allMembers, err := utils.fetchAllMembers(s, m.GuildID)
	if err != nil {
		log.Printf("Error fetching members: %v", err)
		_, err := s.ChannelMessageSend(m.ChannelID, "An error occurred while fetching members.")
		return err
	}

	membersInRole := utils.getMembersInRole(allMembers, targetRole.ID)
	if len(membersInRole) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("No users found in role '%s'.", targetRole.Name))
		return err
	}

	embed := utils.createRoleMembersEmbed(targetRole, membersInRole)
	_, err = s.ChannelMessageSendEmbed(m.ChannelID, embed)
	return err
}
