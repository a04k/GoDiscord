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
	commands.RegisterCommand("roleinfo", RoleInfo, "ri")
}

func RoleInfo(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Please specify a role name, ID, or mention.")
		return
	}

	roleInput := strings.Join(args[1:], " ")
	targetRole, err := utils.FindRole(s, m.GuildID, roleInput)

	if err != nil {
		log.Printf("Role lookup error: %v | Input: %s", err, roleInput)
		s.ChannelMessageSend(m.ChannelID, "Role not found or cannot be fetched.")
		return
	}

	createdTime, _ := discordgo.SnowflakeTimestamp(targetRole.ID)

	var keyPerms []string
	perms := targetRole.Permissions
	if perms&discordgo.PermissionAdministrator != 0 {
		keyPerms = append(keyPerms, "Administrator")
	}
	if perms&discordgo.PermissionManageRoles != 0 {
		keyPerms = append(keyPerms, "Manage Roles")
	}
	if perms&discordgo.PermissionManageMessages != 0 {
		keyPerms = append(keyPerms, "Manage Messages")
	}
	if perms&discordgo.PermissionBanMembers != 0 {
		keyPerms = append(keyPerms, "Ban Members")
	}
	if perms&discordgo.PermissionKickMembers != 0 {
		keyPerms = append(keyPerms, "Kick Members")
	}
	if perms&discordgo.PermissionModerateMembers != 0 {
		keyPerms = append(keyPerms, "Timeout Members")
	}

	permString := "No key permissions"
	if len(keyPerms) > 0 {
		permString = strings.Join(keyPerms, ", ")
	}

	embed := &discordgo.MessageEmbed{
		Title: "Role Information",
		Color: targetRole.Color,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "ID", Value: targetRole.ID, Inline: true},
			{Name: "Name", Value: targetRole.Name, Inline: true},
			{Name: "Color", Value: fmt.Sprintf("#%06x", targetRole.Color), Inline: true},

			{Name: "Mention", Value: fmt.Sprintf("<@&%s>", targetRole.ID), Inline: true},
			{Name: "Hoisted", Value: map[bool]string{true: "Yes", false: "No"}[targetRole.Hoist], Inline: true},
			{Name: "Position", Value: strconv.Itoa(targetRole.Position), Inline: true},

			{Name: "Mentionable", Value: map[bool]string{true: "Yes", false: "No"}[targetRole.Mentionable], Inline: true},
			{Name: "Managed", Value: map[bool]string{true: "Yes", false: "No"}[targetRole.Managed], Inline: true},
			{Name: "\u200B", Value: "\u200B", Inline: true},

			{Name: "Key Permissions", Value: "```\n" + permString + "\n```", Inline: false},
			{Name: "Created At", Value: createdTime.Format("January 2, 2006"), Inline: false},
		},
	}

	if _, err := s.ChannelMessageSendEmbed(m.ChannelID, embed); err != nil {
		log.Printf("Failed to send roleinfo embed: %v", err)
	}
}