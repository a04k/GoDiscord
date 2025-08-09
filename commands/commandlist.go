package commands

import (
	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
)

func CommandList(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	embed := &discordgo.MessageEmbed{
		Title:       "Available Commands",
		Description: "Here is a list of all available commands:",
		Color:       0x0000FF, //blue left bar for commands list

		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "User Commands",
				Value:  "`bal / balance` - Check your balance.\n`work` - Earn coins (6h cooldown).\n`flip <amount / all>` - Gamble coins.\n`transfer <@user> <amount>` - Send coins to another user.\n`usd [amount]` - USD to EGP exchange rate.\n`btc` - Bitcoin price in USD.\n`f1` - View next F1 event and details.\n`nextf1session` - View next F1 session with time.\n`f1sub` - Subscribe/unsubscribe to F1 notifications (DM only).\n`/f1` - Toggle F1 notifications (slash command).\n`/setup <landline> <password>` - Save WE credentials.\n`/quota` - Check internet quota (after setup).",
				Inline: false,
			},
			{
				Name:   "Moderator Commands",
				Value:  "`mute <@user> [reason]` - Mute a user.\n`unmute <@user>` - Unmute a user.\n`voicemute/vm <@user> [reason]` - Mute a user's voice.\n`voiceunmute/vum <@user>` - Unmute a user's voice.",
				Inline: false,
			},
			{
				Name:   "Admin Commands",
				Value:  "`add <@user> <amount>` - add: Add coins to a user.\n`take <@user> <amount>` - take: Remove coins from a user.\n`sa <@user>` - sa/setadmin: Promote a user to admin.\n`cr/createrole <role name> [color] [permissions] [hoist]` - Create a new role.\n`sr/setrole <@user> <role name>` - Assign role to user.\n`inrole <role name or mention>` - View users in a role.\n `ri/roleinfo <role name or mention>` - View role information.\n`ban <@user> [reason(OPTIONAL)] [days(OPTIONAL)]` - Ban a user with optional reason and days of message deletion.",
				Inline: false,
			},
		},
	}

	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}
