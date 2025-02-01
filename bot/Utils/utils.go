package utils

import (
	"fmt"
	"strings"
	"github.com/bwmarrin/discordgo"
)

func ExtractUserID(mention string) (string, error) {
	if !strings.HasPrefix(mention, "<@") || !strings.HasSuffix(mention, ">") {
		return "", fmt.Errorf("invalid mention format")
	}
	userID := mention[2 : len(mention)-1]
	if strings.HasPrefix(userID, "!") {
		userID = userID[1:] // to remove nickname exclamation mark
	}
	return userID, nil
}

func FindRole(roles []*discordgo.Role, input string) *discordgo.Role {
	if strings.HasPrefix(input, "<@&") && strings.HasSuffix(input, ">") {
		// Handle role mention
		roleID := input[3 : len(input)-1]
		for _, r := range roles {
			if r.ID == roleID {
				return r
			}
		}
	} else {
		// Handle role name
		for _, r := range roles {
			if strings.EqualFold(r.Name, input) {
				return r
			}
		}
	}
	return nil
}

func GetHighestRolePosition(memberRoles []string, guildRoles []*discordgo.Role) int { // this is for mods to not set roles higher than their own
	var highest int
	for _, roleID := range memberRoles {
		for _, role := range guildRoles {
			if role.ID == roleID && role.Position > highest {
				highest = role.Position
			}
		}
	}
	return highest
}

func FetchAllMembers(s *discordgo.Session, guildID string) ([]*discordgo.Member, error) {
	var allMembers []*discordgo.Member
	after := ""
	for {
		members, err := s.GuildMembers(guildID, after, 1000)
		if err != nil {
			return nil, err
		}
		if len(members) == 0 {
			break
		}
		allMembers = append(allMembers, members...)
		after = members[len(members)-1].User.ID
	}
	return allMembers, nil
}

func GetMembersInRole(members []*discordgo.Member, roleID string) []string {
	var membersInRole []string
	for _, member := range members {
		for _, memberRoleID := range member.Roles {
			if memberRoleID == roleID {
				name := member.Nick
				if name == "" {
					name = member.User.Username
				}
				membersInRole = append(membersInRole, name)
				break
			}
		}
	}
	return membersInRole
}

func CreateRoleMembersEmbed(role *discordgo.Role, members []string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Users in role: %s", role.Name),
		Description: strings.Join(members, "\n"),
		Color:       role.Color,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Total members: %d", len(members)),
		},
	}
}
