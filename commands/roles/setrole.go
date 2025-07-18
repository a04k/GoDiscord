package roles

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
)

func SetRole(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 3 {
		s.ChannelMessageSend(m.ChannelID, "Usage: .setrole/sr <@user> <role name or mention>")
		return
	}

	// Extract user mention and role name
	userMention := args[1]
	roleName := strings.Join(args[2:], " ")

	// Get user ID from mention
	if !strings.HasPrefix(userMention, "<@") || !strings.HasSuffix(userMention, ">") {
		s.ChannelMessageSend(m.ChannelID, "Invalid user mention. Example: .addrole @User RoleName")
		return
	}
	userID := userMention[2 : len(userMention)-1]
	if strings.HasPrefix(userID, "!") {
		userID = userID[1:] // Remove nickname exclamation mark
	}

	// Fetch guild and roles
	guild, err := s.Guild(m.GuildID)
	if err != nil {
		log.Printf("Error fetching guild: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred while fetching server information.")
		return
	}

	// Find the target role
	var targetRole *discordgo.Role
	if strings.HasPrefix(roleName, "<@&") && strings.HasSuffix(roleName, ">") {
		// Handle role mention
		roleID := roleName[3 : len(roleName)-1]
		for _, r := range guild.Roles {
			if r.ID == roleID {
				targetRole = r
				break
			}
		}
	} else {
		// Handle role name
		for _, r := range guild.Roles {
			if strings.EqualFold(r.Name, roleName) {
				targetRole = r
				break
			}
		}
	}

	if targetRole == nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Role '%s' not found.", roleName))
		return
	}

	// Fetch member executing the command
	executor, err := s.GuildMember(m.GuildID, m.Author.ID)
	if err != nil {
		log.Printf("Error fetching executor: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred while verifying permissions.")
		return
	}

	// Find the executor's highest role position
	var executorHighestPosition int
	for _, execRoleID := range executor.Roles {
		for _, role := range guild.Roles {
			if role.ID == execRoleID && role.Position > executorHighestPosition {
				executorHighestPosition = role.Position
			}
		}
	}

	// Ensure executor can assign the target role
	if targetRole.Position >= executorHighestPosition {
		s.ChannelMessageSend(m.ChannelID, "You cannot assign a role equal to or higher than your highest role.")
		return
	}

	// Add the role to the target user
	err = s.GuildMemberRoleAdd(m.GuildID, userID, targetRole.ID)
	if err != nil {
		log.Printf("Error adding role: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred while adding the role to the user.")
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Role '%s' added to user <@%s> successfully!", targetRole.Name, userID))
}
