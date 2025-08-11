package roles

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
	"DiscordBot/utils"
	"DiscordBot/commands"
)

func init() {
	commands.RegisterCommand("createrole", CreateRole, "cr")
}

func CreateRole(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Usage: .createrole/cr <role name> [color hex] [permissions : (ε: default, mod/owner)] [hoist: (ε:false , hoist/true/1)]")
		return
	}

	// Check if the user has manage roles permission
	hasManageRoles, err := utils.CheckManageRolesPermission(s, m.GuildID, m.Author.ID)
	if err != nil {
		log.Printf("Error checking manage roles permission: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred while verifying permissions.")
		return
	}

	if !hasManageRoles {
		s.ChannelMessageSend(m.ChannelID, "You must have Manage Roles permission to use this command.")
		return
	}

	// Extract role name
	roleName := args[1]
	roleColor := 0 // Default color (no color)
	hoist := false // Default hoisting (not visible)
	mentionable := false

	// Check for optional color hex
	if len(args) > 2 && strings.HasPrefix(args[2], "#") {
		colorHex := args[2]
		parsedColor, err := strconv.ParseInt(colorHex[1:], 16, 32)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Invalid color hex. Example: #FF5733")
			return
		}
		roleColor = int(parsedColor)
	}

	perms := 0
	if len(args) > 3 {
		permVal := args[3]
		perms = utils.ParsePermissions(permVal)
		if perms == 0 {
			s.ChannelMessageSend(m.ChannelID, "Invalid permissions input. Example: 'mod' / 'owner'")
			return
		}
	}

	// Check for hoist flag (optional)
	if len(args) > 4 && (strings.ToLower(args[4]) == "hoist" || strings.ToLower(args[4]) == "true" || strings.ToLower(args[4]) == "1") {
		hoist = true // Enable hoisting if 'hoist' is provided as the 5th argument
	}
	if len(args) > 5 && (strings.ToLower(args[4]) == "hoist" || strings.ToLower(args[4]) == "true" || strings.ToLower(args[4]) == "1") {
		mentionable = true // makes role mentionable if provided as a 6th arg
	}

	// perms to int64
	permsInt64 := int64(perms)

	// executing the role creation call (GuildRoleCreate)
	newRole, err := s.GuildRoleCreate(m.GuildID, &discordgo.RoleParams{ // roleparams requiring perms val in int64 is stupid since perms max value admin = 8
		Name:        roleName,
		Color:       &roleColor,  // pointer to int for color
		Permissions: &permsInt64, // convert perms to int64 and pass a pointer
		Hoist:       &hoist,      // using a pointer to bool for hoist (hoisting: having the role appear on the member list)
		Mentionable: &mentionable,
	})
	if err != nil {
		log.Printf("Error creating role: %v", err)
		s.ChannelMessageSend(m.ChannelID, "An error occurred while creating the role.")
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Role '%s' created successfully!", newRole.Name))
}