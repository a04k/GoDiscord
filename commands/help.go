package commands

import (
	"database/sql"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
)

func init() {
	RegisterCommand("help", Help, "h")
}

func Help(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) > 1 {
		commandName := strings.ToLower(args[1])
		// Check if the command is disabled
		var count int
		err := b.Db.QueryRow("SELECT COUNT(*) FROM disabled_commands WHERE guild_id = $1 AND name = $2 AND type = 'command'",
			m.GuildID, commandName).Scan(&count)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("Error checking if command %s is disabled in guild %s: %v", commandName, m.GuildID, err)
		} else if count > 0 {
			s.ChannelMessageSend(m.ChannelID, "This command is disabled.")
			return
		}

		// Check if the command's category is disabled
		for category, cmds := range CommandCategories {
			for _, cmd := range cmds {
				if cmd == commandName {
					err := b.Db.QueryRow("SELECT COUNT(*) FROM disabled_commands WHERE guild_id = $1 AND name = $2 AND type = 'category'",
						m.GuildID, strings.ToLower(category)).Scan(&count)
					if err != nil && err != sql.ErrNoRows {
						log.Printf("Error checking if category %s is disabled in guild %s: %v", category, m.GuildID, err)
					} else if count > 0 {
						s.ChannelMessageSend(m.ChannelID, "This command's category is disabled.")
						return
					}
					break
				}
			}
		}
	}

	embed := &discordgo.MessageEmbed{
		Title: "Help",
		Description: "Here is a list of commands. For more information on a specific command, type `.help <command>`.",
		Color: 0x00ff00,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name: "General",
				Value: "`.help`, `.commandlist`, `.usd`, `.btc`",
				Inline: false,
			},
			{
				Name: "Economy",
				Value: "`.balance`, `.work`, `.transfer`, `.flip`",
				Inline: false,
			},
			{
				Name: "F1",
				Value: "`.f1`, `.f1results`, `.f1standings`, `.f1wdc`, `.f1wcc`, `.qualiresults`, `.nextf1session`, `.f1sub`",
				Inline: false,
			},
			{
				Name: "Moderation",
				Value: "`.kick`, `.mute`, `.unmute`, `.voicemute`, `.vunmute`, `.ban`, `.unban`",
				Inline: false,
			},
			{
				Name: "Admin",
				Value: "`.setadmin`, `.add`, `.take`, `.disable`, `.enable`",
				Inline: false,
			},
			{
				Name: "Roles",
				Value: "`.createrole`, `.setrole`, `.inrole`, `.roleinfo`",
				Inline: false,
			},
		},
	}

	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}