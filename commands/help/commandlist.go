package help

import (
	"fmt"
	"log"
	"strings"

	"DiscordBot/bot"
	"DiscordBot/commands"

	"github.com/bwmarrin/discordgo"
)

func init() {
	commands.RegisterCommand("commandlist", CommandList, "cl")
}

func CommandList(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// Ensure command is used in a guild
	if m.GuildID == "" {
		return // Don't respond to DMs
	}

	// If a category is specified, show only commands from that category
	if len(args) > 1 {
		category := strings.Title(strings.ToLower(args[1]))

		// Build categories from modules
		categories := make(map[string][]string)
		for _, module := range commands.RegisteredModules {
			if module.Category != "" {
				for _, cmd := range module.Commands {
					categories[module.Category] = append(categories[module.Category], cmd.Name)
				}
			}
		}

		cmds, exists := categories[category]
		if !exists {
			s.ChannelMessageSend(m.ChannelID, "Invalid category. Use `.commandlist` to see all categories.")
			return
		}

		// Build embed for the specific category
		embed := &discordgo.MessageEmbed{
			Title: fmt.Sprintf("Commands - %s", category),
			Color: 0x00ff00,
		}

		// Add commands in this category
		for _, cmdName := range cmds {
			// Get command info
			cmdInfo, exists := commands.CommandDetails[cmdName]
			if !exists {
				continue
			}

			description := cmdInfo.Description
			if description == "" {
				description = "No description available"
			}

			// Add aliases if they exist
			aliasText := ""
			if len(cmdInfo.Aliases) > 0 {
				aliasText = fmt.Sprintf(" (Aliases: %s)", strings.Join(cmdInfo.Aliases, ", "))
			}

			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:  fmt.Sprintf(".%s%s", cmdName, aliasText),
				Value: description,
			})
		}

		s.ChannelMessageSendEmbed(m.ChannelID, embed)
		return
	}

	// Create paginated command list
	pages, err := CreateCommandListPages(b, s, m)
	if err != nil {
		log.Printf("Error creating command list pages: %v", err)
		s.ChannelMessageSend(m.ChannelID, "Error generating command list.")
		return
	}

	if len(pages) == 0 {
		s.ChannelMessageSend(m.ChannelID, "No commands available.")
		return
	}

	// Send the first page
	msg, err := s.ChannelMessageSendEmbed(m.ChannelID, pages[0])
	if err != nil {
		log.Printf("Error sending command list: %v", err)
		return
	}

	// If there's only one page, no need for pagination
	if len(pages) <= 1 {
		return
	}

	// Add reactions for pagination
	err = s.MessageReactionAdd(msg.ChannelID, msg.ID, "⬅️")
	if err != nil {
		log.Printf("Error adding left reaction: %v", err)
	}

	err = s.MessageReactionAdd(msg.ChannelID, msg.ID, "➡️")
	if err != nil {
		log.Printf("Error adding right reaction: %v", err)
	}

	// Store pagination state
	state := &PaginationState{
		UserID:      m.Author.ID,
		MessageID:   msg.ID,
		CurrentPage: 0,
		TotalPages:  len(pages),
		Pages:       pages,
		GuildID:     m.GuildID,
	}

	paginationManager.AddState(state)
}
