package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func UserExists(s *discordgo.Session, userID string) (bool, error) {
	_, err := s.User(userID)
	if err != nil {
		if strings.Contains(err.Error(), "Unknown User") {
			return false, nil // User does not exist
		}
		return false, err // Other error
	}
	return true, nil // User exists
}

// ExtractUserID extracts the user ID from a mention
func ExtractUserID(mention string) (string, error) {
	// Check if the mention is properly formatted
	if !strings.HasPrefix(mention, "<@") || !strings.HasSuffix(mention, ">") {
		return "", fmt.Errorf("invalid mention format")
	}

	// Extract the user ID
	userID := strings.TrimPrefix(strings.TrimSuffix(mention, ">"), "<@")

	// Remove the nickname exclamation mark if present
	userID = strings.TrimPrefix(userID, "!")

	// Validate that the user ID is a valid Snowflake (Discord ID)
	if _, err := strconv.ParseUint(userID, 10, 64); err != nil {
		return "", fmt.Errorf("invalid user ID")
	}

	return userID, nil
}


// permission parsing function
func ParsePermissions(permVal string) int {
	perms := 0
	switch strings.ToUpper(permVal) {
	case "MOD":
		// Mod bundle
		perms |= discordgo.PermissionManageRoles
		perms |= discordgo.PermissionChangeNickname
		perms |= discordgo.PermissionVoiceMuteMembers
		perms |= discordgo.PermissionManageMessages
	case "OWNER":
		// Owner
		perms |= discordgo.PermissionBanMembers
		perms |= discordgo.PermissionManageChannels
		perms |= discordgo.PermissionManageRoles
		perms |= discordgo.PermissionChangeNickname
		perms |= discordgo.PermissionVoiceMuteMembers
		perms |= discordgo.PermissionManageMessages
	case "DEFAULT":
		// Default user permissions (no extra permissions)
		// just return 0 for no permissions (this is the default behavior)
		return 0
	default:
		return 0 // If the permission is not recognized, return 0
	}
	return perms
}

func ExtractRoleID(input string) string {
	if strings.HasPrefix(input, "<@&") && strings.HasSuffix(input, ">") {
		return input[3 : len(input)-1]
	}
	return input // Return as is for ID/name validation
}

// function to validate and find role
func FindRole(s *discordgo.Session, guildID string, roleInput string) (*discordgo.Role, error) {
	roles, err := s.GuildRoles(guildID)
	if err != nil {
		return nil, err
	}

	cleanedInput := ExtractRoleID(roleInput)

	// Check by ID first
	for _, role := range roles {
		if role.ID == cleanedInput {
			return role, nil
		}
	}

	// Check by name (case-insensitive)
	cleanedInput = strings.ToLower(strings.TrimSpace(cleanedInput))
	for _, role := range roles {
		if strings.ToLower(role.Name) == cleanedInput {
			return role, nil
		}
	}

	return nil, fmt.Errorf("role not found")
}

// this fetches the highest role a mod can assign, cannot be above the users/mod's current role.
func GetHighestRole(member *discordgo.Member, roles []*discordgo.Role) *discordgo.Role {
	var highestRole *discordgo.Role
	for _, roleID := range member.Roles {
		for _, role := range roles {
			if role.ID == roleID {
				if highestRole == nil || role.Position > highestRole.Position {
					highestRole = role
				}
			}
		}
	}
	return highestRole
}

func GetMutedRole(s *discordgo.Session, guildID string) (*discordgo.Role, error) {
	// fetch all roles in server
	roles, err := s.GuildRoles(guildID)
	if err != nil {
		return nil, fmt.Errorf("error fetching guild roles: %v", err)
	}

	// look for muted role if it exists (which it should, if not then it is handled below)
	for _, role := range roles {
		if role.Name == "Muted" {
			return role, nil
		}
	}

	// creating a muted role if it doesn't exist : this will really only be used for servers that havent been set up yet / new servers
	mutedRole, err := s.GuildRoleCreate(guildID, &discordgo.RoleParams{
		Name: "Muted",
	})
	if err != nil {
		return nil, fmt.Errorf("error creating muted role: %v", err)
	}

	// Set the role name to "Muted" and permissions to 0 (no permissions)
	_, err = s.GuildRoleEdit(guildID, mutedRole.ID, &discordgo.RoleParams{
		Name:        "Muted",
		Permissions: new(int64),
	})
	if err != nil {
		return nil, fmt.Errorf("error editing muted role: %v", err)
	}

	// Update channel permissions to restrict the "Muted" role from sending messages
	channels, err := s.GuildChannels(guildID)
	if err != nil {
		return nil, fmt.Errorf("error fetching guild channels: %v", err)
	}
	for _, channel := range channels {
		// Deny SendMessages permission for the "Muted" role
		err = s.ChannelPermissionSet(channel.ID, mutedRole.ID, discordgo.PermissionOverwriteTypeRole, 0, discordgo.PermissionSendMessages)
		if err != nil {
			return nil, fmt.Errorf("error setting channel permissions for muted role: %v", err)
		}

		//Deny Speak permission on voice channels too
		if channel.Type == discordgo.ChannelTypeGuildVoice {
			err = s.ChannelPermissionSet(channel.ID, mutedRole.ID, discordgo.PermissionOverwriteTypeRole, 0, discordgo.PermissionVoiceSpeak)
			if err != nil {
				return nil, fmt.Errorf("error setting voice channel permissions for muted role: %v", err)
			}
		}
	}

	return mutedRole, nil
}

// CheckPermission checks if a user has a specific permission in a guild
func CheckPermission(s *discordgo.Session, guildID, userID string, permission int64) (bool, error) {
	// Fetch guild member details
	member, err := s.GuildMember(guildID, userID)
	if err != nil {
		return false, fmt.Errorf("error fetching member: %v", err)
	}

	// Fetch guild roles
	guild, err := s.Guild(guildID)
	if err != nil {
		return false, fmt.Errorf("error fetching guild: %v", err)
	}

	// Check if the user has the required permission
	hasPermission := false
	for _, roleID := range member.Roles {
		for _, role := range guild.Roles {
			if role.ID == roleID && role.Permissions&permission != 0 {
				hasPermission = true
				break
			}
		}
	}

	return hasPermission, nil
}

// CheckAdminPermission checks if a user has administrator permissions in a guild
func CheckAdminPermission(s *discordgo.Session, guildID, userID string) (bool, error) {
	return CheckPermission(s, guildID, userID, discordgo.PermissionAdministrator)
}

// CheckManageMessagesPermission checks if a user has manage messages permissions in a guild
func CheckManageMessagesPermission(s *discordgo.Session, guildID, userID string) (bool, error) {
	return CheckPermission(s, guildID, userID, discordgo.PermissionManageMessages)
}

// CheckManageRolesPermission checks if a user has manage roles permissions in a guild
func CheckManageRolesPermission(s *discordgo.Session, guildID, userID string) (bool, error) {
	return CheckPermission(s, guildID, userID, discordgo.PermissionManageRoles)
}

// CheckKickMembersPermission checks if a user has kick members permissions in a guild
func CheckKickMembersPermission(s *discordgo.Session, guildID, userID string) (bool, error) {
	return CheckPermission(s, guildID, userID, discordgo.PermissionKickMembers)
}

// CheckBanMembersPermission checks if a user has ban members permissions in a guild
func CheckBanMembersPermission(s *discordgo.Session, guildID, userID string) (bool, error) {
	return CheckPermission(s, guildID, userID, discordgo.PermissionBanMembers)
}

// CheckMuteMembersPermission checks if a user has mute members permissions in a guild
func CheckMuteMembersPermission(s *discordgo.Session, guildID, userID string) (bool, error) {
	return CheckPermission(s, guildID, userID, discordgo.PermissionManageMessages)
}