package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
	"DiscordBot/utils"
)

func Help(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 2 {
		embed := &discordgo.MessageEmbed{
			Title:       "Usage:",
			Description: ".h / .help [commandname] \n Use .cl / .commandlist to view all commands",
			Color:       0x0000FF, //blue left bar for commands list
		}

		s.ChannelMessageSendEmbed(m.ChannelID, embed)
		return
	}

	// If a command is specified, display individual command help
	command := strings.ToLower(args[1])
	var response string
	noPermissionMessage := "You do not have permission to access this command."

	// Check if the user is an admin or mod
	isMod, _ := utils.IsModerator(b.Db, m.Author.ID)
	isOwner, _ := utils.IsAdmin(b.Db, m.Author.ID)

	switch command {
	case "bal", "balance":
		response = ".bal or .balance\nCheck your balance."

	case "work":
		response = ".work\nEarn coins (6h cooldown)."

	case "flip":
		response = ".flip <amount / all>\nGamble coins: (specified amount / all)."

	case "transfer":
		response = ".transfer <@user> <amount>\nSend coins to another user."

	case "usd":
		response = ".usd [amount]\nUSD to EGP exchange rate."

	case "btc":
		response = ".btc\nBitcoin price in USD."

	case "setup":
		response = "/setup <landline> <password>\nSave WE credentials."

	case "quota":
		response = "/quota\nCheck internet quota (after setup)."

	// Admin Commands
	case "add":
		if isOwner {
			response = ".add <@user> <amount>\nAdmin: Add coins to a user."
		} else {
			response = noPermissionMessage
		}

	case "sa":
		if isOwner {
			response = ".sa <@user>\nAdmin: Promote a user to admin."
		} else {
			response = noPermissionMessage
		}

	case "createrole":
		if isOwner {
			response = ".createrole <role name> [color] [permissions] [hoist]\nCreate a new role with options for color, permissions, and hoisting."
		} else {
			response = noPermissionMessage
		}

	case "setrole", "sr":
		if isOwner || isMod {
			response = ".setrole <@user> <role name>\nGives the user the role."
		} else {
			response = noPermissionMessage
		}

	case "inrole":
		if isOwner || isMod {
			response = ".inrole <role name or mention>\nView all users in a specific role."
		} else {
			response = noPermissionMessage
		}

	case "roleinfo", "ri":
		if isOwner || isMod {
			response = ".ri / roleinfo <role name or mention>\nView role information."
		} else {
			response = noPermissionMessage
		}

	case "ban":
		if isOwner {
			response = ".ban <@user> [reason] [days]\nBan a user with optional reason and days of message deletion."
		} else {
			response = noPermissionMessage
		}

	default:
		response = "Command not found. Use .help to get a list of available commands."
	}

	// Send command help in an embed
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Help: %s", command),
		Description: response,
		Color:       0xFF0000, // red left bar for command help
	}
	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}