package utils

import (
	"database/sql"
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

func ExtractUserID(mention string) (string, error) {
	// Check if the mention is properly formatted
	if !strings.HasPrefix(mention, "<@") || !strings.HasSuffix(mention, ">") {
		return "", fmt.Errorf("invalid mention format")
	}

	// Extract the user ID
	userID := strings.TrimPrefix(strings.TrimSuffix(mention, ">"), "<@")

	// Validate that the user ID is a valid Snowflake (Discord ID)
	if _, err := strconv.ParseUint(userID, 10, 64); err != nil {
		return "", fmt.Errorf("invalid user ID")
	}

	return userID, nil
}

func IsAdmin(db *sql.DB, guildID, userID string) (bool, error) {
	var isAdmin bool
	var isOwner bool
	err := db.QueryRow("SELECT is_admin, is_owner FROM users WHERE guild_id = $1 AND user_id = $2", guildID, userID).Scan(&isAdmin, &isOwner)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil // user not found, cant be admin
		}
		return false, err // db err
	}
	return isAdmin || isOwner, nil
}

func IsOwner(db *sql.DB, guildID, userID string) (bool, error) {
	var isOwner bool
	err := db.QueryRow("SELECT is_owner FROM users WHERE guild_id = $1 AND user_id = $2", guildID, userID).Scan(&isOwner)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil // user not found, cant be owner
		}
		return false, err // db err
	}
	return isOwner, nil
}

func IsModerator(db *sql.DB, userID string) (bool, error) {
	var isMod bool
	err := db.QueryRow("SELECT is_mod FROM users WHERE user_id = $1", userID).Scan(&isMod)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil // user not found, cant be mod
		}
		return false, err // db err
	}
	return isMod, nil
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
