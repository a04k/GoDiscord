package help

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"DiscordBot/bot"
	"DiscordBot/commands"

	"github.com/bwmarrin/discordgo"
)

func init() {
	commands.RegisterCommand("help", Help, "h")
}

func Help(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// Ensure command is used in a guild
	if m.GuildID == "" {
		return // Don't respond to DMs
	}

	if len(args) > 1 {
		commandName := strings.ToLower(args[1])

		// Resolve alias to actual command name
		if actualName, isAlias := commands.CommandAliases[commandName]; isAlias {
			commandName = actualName
		}

		// Check if the command exists
		commandInfo, exists := commands.CommandDetails[commandName]
		if !exists {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Command `%s` not found.", commandName))
			return
		}

		// Check if the command is disabled
		var count int
		err := b.Db.QueryRow("SELECT COUNT(*) FROM disabled_commands WHERE guild_id = $1 AND name = $2 AND type = 'command'",
			m.GuildID, commandName).Scan(&count)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("Error checking if command %s is disabled in guild %s: %v", commandName, m.GuildID, err)
		} else if count > 0 {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Command `%s` is disabled.", commandName))
			return
		}

		// Check if the command's category is disabled
		category := commandInfo.Category
		err = b.Db.QueryRow("SELECT COUNT(*) FROM disabled_commands WHERE guild_id = $1 AND name = $2 AND type = 'category'",
			m.GuildID, strings.ToLower(category)).Scan(&count)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("Error checking if category %s is disabled in guild %s: %v", category, m.GuildID, err)
		} else if count > 0 {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Command `%s` is in category `%s` which is disabled.", commandName, category))
			return
		}

		// Build help embed for the specific command
		embed := &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("Help: %s", commandInfo.Name),
			Description: commandInfo.Description,
			Color:       0x00ff00,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:  "Usage",
					Value: fmt.Sprintf("`%s`", commandInfo.Usage),
				},
			},
		}

		// Add aliases if they exist
		if len(commandInfo.Aliases) > 0 {
			aliases := strings.Join(commandInfo.Aliases, ", ")
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:  "Aliases",
				Value: aliases,
			})
		}

		// Add category
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "Category",
			Value: commandInfo.Category,
		})

		s.ChannelMessageSendEmbed(m.ChannelID, embed)
		return
	}

	// General help - show categories
	embed := &discordgo.MessageEmbed{
		Title:       "Help",
		Description: "Here is a list of command categories. For more information on a specific command, type `.help <command>`.",
		Color:       0x00ff00,
	}

	// Add a field for each category using the module system
	categories := make(map[string]bool)
	for _, module := range commands.RegisteredModules {
		if module.Category != "" {
			categories[module.Category] = true
		}
	}

	for category := range categories {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   category,
			Value:  fmt.Sprintf("`.help %s` for commands in this category", strings.ToLower(category)),
			Inline: false,
		})
	}

	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}
