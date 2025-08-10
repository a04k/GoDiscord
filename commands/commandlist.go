package commands

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
)

func init() {
	RegisterCommand("commandlist", CommandList, "cl")
}

func CommandList(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	rows, err := b.Db.Query("SELECT name, type FROM disabled_commands WHERE guild_id = $1", m.GuildID)
	if err != nil {
		log.Printf("Error getting disabled commands for guild %s: %v", m.GuildID, err)
		s.ChannelMessageSend(m.ChannelID, "Error getting command list.")
		return
	}
	defer rows.Close()

	disabledCommandsMap := make(map[string]bool)
	disabledCategoriesMap := make(map[string]bool)
	var disabledCommandsList []string

	for rows.Next() {
		var name, dType string
		if err := rows.Scan(&name, &dType); err != nil {
			log.Printf("Error scanning disabled command: %v", err)
			continue
		}
		if dType == "command" {
			disabledCommandsMap[name] = true
			disabledCommandsList = append(disabledCommandsList, ".`"+name+"`")
		} else if dType == "category" {
			disabledCategoriesMap[strings.ToLower(name)] = true
		}
	}

	enabledCommandsBuilder := &strings.Builder{}
	for category, cmds := range CommandCategories {
		if !disabledCategoriesMap[strings.ToLower(category)] {
			enabledInCategory := &strings.Builder{}
			for _, cmd := range cmds {
				if !disabledCommandsMap[cmd] {
					fmt.Fprintf(enabledInCategory, ".%s `", cmd)
				}
			}
			if enabledInCategory.Len() > 0 {
				fmt.Fprintf(enabledCommandsBuilder, "**%s**\n%s\n\n", category, enabledInCategory.String())
			}
		}
	}

	embed := &discordgo.MessageEmbed{
		Title: "Command List",
		Color: 0x00ff00,
	}

	if enabledCommandsBuilder.Len() > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "Enabled Commands",
			Value: enabledCommandsBuilder.String(),
			Inline: false,
		})
	}

	if len(disabledCommandsList) > 0 {
			disabledSection := &strings.Builder{}
			fmt.Fprintf(disabledSection, "**Commands**\n%s\n\n", strings.Join(disabledCommandsList, "\n"))

			var disabledCatList []string
			for cat := range disabledCategoriesMap {
				disabledCatList = append(disabledCatList, "`"+strings.Title(cat)+"`")
			}
			if len(disabledCatList) > 0 {
				fmt.Fprintf(disabledSection, "**Categories**\n%s\n", strings.Join(disabledCatList, "\n"))
			}

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "Disabled",
			Value: disabledSection.String(),
			Inline: false,
		})
	}

	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}