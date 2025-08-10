package roles

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
	"DiscordBot/commands"
)

func init() {
	commands.RegisterCommand("inrole", InRole)
}

func InRole(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Usage: .inrole <role name or mention>")
		return
	}

	// Combine arguments into a role name or mention
	input := strings.TrimSpace(strings.Join(args[1:], " "))
	log.Printf("Input received: '%s'", input)

	// Fetch guild information
	guild, err := s.Guild(m.GuildID)
	if err != nil {
		log.Printf("Error fetching guild: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred while fetching server information.")
		return
	}

	// Find the target role (by mention or name)
	var targetRole *discordgo.Role
	if strings.HasPrefix(input, "<@&") && strings.HasSuffix(input, ">") {
		// Handle role mention
		roleID := input[3 : len(input)-1] // Extract role ID
		for _, r := range guild.Roles {
			if r.ID == roleID {
				targetRole = r
				break
			}
		}
	} else {
		// Handle role name
		for _, r := range guild.Roles {
			if strings.EqualFold(r.Name, input) {
				targetRole = r
				break
			}
		}
	}

	if targetRole == nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Role '%s' not found.", input))
		return
	}

	log.Printf("Role found: '%s' (ID: %s)", targetRole.Name, targetRole.ID)

	// Fetch members in the guild
	var allMembers []*discordgo.Member
	after := ""
	for {
		members, err := s.GuildMembers(m.GuildID, after, 1000)
		if err != nil {
			log.Printf("Error fetching members: %v", err)
			s.ChannelMessageSend(m.ChannelID, "An error occurred while fetching members.")
			return
		}
		if len(members) == 0 {
			break
		}
		allMembers = append(allMembers, members...)
		after = members[len(members)-1].User.ID
	}

	// Filter members with the target role
	var membersInRole []string
	for _, member := range allMembers {
		for _, roleID := range member.Roles {
			if roleID == targetRole.ID {
				// Add nickname or username
				if member.Nick != "" {
					membersInRole = append(membersInRole, member.Nick)
				} else {
					membersInRole = append(membersInRole, member.User.Username)
				}
				break
			}
		}
	}

	if len(membersInRole) == 0 {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("No users found in role '%s'.", targetRole.Name))
		return
	}

	// Create and send the embed message
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Users in role: %s", targetRole.Name),
		Description: strings.Join(membersInRole, "\n"),
		Color:       targetRole.Color,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Total members: %d", len(membersInRole)),
		},
	}

	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}
