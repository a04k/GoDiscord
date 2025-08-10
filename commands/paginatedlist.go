package commands

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
)

// PaginationState holds the state of paginated messages
type PaginationState struct {
	UserID       string
	MessageID    string
	CurrentPage  int
	TotalPages   int
	Pages        []*discordgo.MessageEmbed
	GuildID      string
}

// PaginationManager manages pagination states
type PaginationManager struct {
	states map[string]*PaginationState // messageID -> state
	mu     sync.RWMutex
}

// NewPaginationManager creates a new pagination manager
func NewPaginationManager() *PaginationManager {
	return &PaginationManager{
		states: make(map[string]*PaginationState),
	}
}

// AddState adds a pagination state
func (pm *PaginationManager) AddState(state *PaginationState) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.states[state.MessageID] = state
}

// GetState retrieves a pagination state
func (pm *PaginationManager) GetState(messageID string) (*PaginationState, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	state, exists := pm.states[messageID]
	return state, exists
}

// RemoveState removes a pagination state
func (pm *PaginationManager) RemoveState(messageID string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	delete(pm.states, messageID)
}

// Global pagination manager
var paginationManager = NewPaginationManager()

// CreateCommandListPages creates paginated embeds for the command list
func CreateCommandListPages(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate) ([]*discordgo.MessageEmbed, error) {
	// Get disabled commands and categories for this guild
	rows, err := b.Db.Query("SELECT name, type FROM disabled_commands WHERE guild_id = $1", m.GuildID)
	if err != nil {
		log.Printf("Error getting disabled commands for guild %s: %v", m.GuildID, err)
		return nil, err
	}
	defer rows.Close()

	disabledCommandsMap := make(map[string]bool)
	disabledCategoriesMap := make(map[string]bool)

	for rows.Next() {
		var name, dType string
		if err := rows.Scan(&name, &dType); err != nil {
			log.Printf("Error scanning disabled command: %v", err)
			continue
		}
		if dType == "command" {
			disabledCommandsMap[name] = true
		} else if dType == "category" {
			disabledCategoriesMap[strings.ToLower(name)] = true
		}
	}

	// Create pages for each category
	var pages []*discordgo.MessageEmbed
	
	// Page 1: Categories overview
	categoryPage := &discordgo.MessageEmbed{
		Title: "Command List - Categories",
		Description: "Use the arrow reactions to navigate between pages.\nEach page shows commands in a specific category.",
		Color: 0x00ff00,
	}
	
	for category, cmds := range CommandCategories {
		// Count enabled commands in this category
		enabledCount := 0
		for _, cmd := range cmds {
			if !disabledCommandsMap[cmd] {
				enabledCount++
			}
		}
		
		// Check if category itself is disabled
		isDisabled := disabledCategoriesMap[strings.ToLower(category)]
		
		status := "✅"
		if isDisabled {
			status = "❌"
		}
		
		categoryPage.Fields = append(categoryPage.Fields, &discordgo.MessageEmbedField{
			Name:  fmt.Sprintf("%s %s", status, category),
			Value: fmt.Sprintf("%d commands (`.cl %s`)", enabledCount, strings.ToLower(category)),
		})
	}
	
	pages = append(pages, categoryPage)
	
	// Create a page for each category
	for category, cmds := range CommandCategories {
		// Skip disabled categories
		if disabledCategoriesMap[strings.ToLower(category)] {
			continue
		}
		
		categoryPage := &discordgo.MessageEmbed{
			Title: fmt.Sprintf("Command List - %s", category),
			Color: 0x00ff00,
		}
		
		// Add commands in this category
		for _, cmdName := range cmds {
			// Skip disabled commands
			if disabledCommandsMap[cmdName] {
				continue
			}
			
			// Get command info
			cmdInfo, exists := CommandDetails[cmdName]
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
			
			categoryPage.Fields = append(categoryPage.Fields, &discordgo.MessageEmbedField{
				Name:  fmt.Sprintf(".%s%s", cmdName, aliasText),
				Value: description,
			})
		}
		
		// Only add the page if there are commands to show
		if len(categoryPage.Fields) > 0 {
			pages = append(pages, categoryPage)
		}
	}
	
	// Create a page for disabled commands if there are any
	if len(disabledCommandsMap) > 0 || len(disabledCategoriesMap) > 0 {
		disabledPage := &discordgo.MessageEmbed{
			Title: "Disabled Commands & Categories",
			Color: 0xff0000,
		}
		
		// Add disabled commands
		if len(disabledCommandsMap) > 0 {
			disabledCmds := make([]string, 0, len(disabledCommandsMap))
			for cmd := range disabledCommandsMap {
				disabledCmds = append(disabledCmds, fmt.Sprintf("`.%s`", cmd))
			}
			
			disabledPage.Fields = append(disabledPage.Fields, &discordgo.MessageEmbedField{
				Name:  "Disabled Commands",
				Value: strings.Join(disabledCmds, "\n"),
			})
		}
		
		// Add disabled categories
		if len(disabledCategoriesMap) > 0 {
			disabledCats := make([]string, 0, len(disabledCategoriesMap))
			for cat := range disabledCategoriesMap {
				disabledCats = append(disabledCats, fmt.Sprintf("`%s`", strings.Title(cat)))
			}
			
			disabledPage.Fields = append(disabledPage.Fields, &discordgo.MessageEmbedField{
				Name:  "Disabled Categories",
				Value: strings.Join(disabledCats, "\n"),
			})
		}
		
		pages = append(pages, disabledPage)
	}
	
	// Add footer to each page with page number
	for i, page := range pages {
		page.Footer = &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Page %d of %d", i+1, len(pages)),
		}
	}
	
	return pages, nil
}

// HandlePagination handles pagination reactions
func HandlePagination(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	// Ignore bot reactions
	if r.UserID == s.State.User.ID {
		return
	}
	
	// Only handle arrow reactions
	if r.Emoji.Name != "⬅️" && r.Emoji.Name != "➡️" {
		return
	}
	
	// Get pagination state
	state, exists := paginationManager.GetState(r.MessageID)
	if !exists {
		return
	}
	
	// Only allow the user who requested the command to paginate
	if r.UserID != state.UserID {
		return
	}
	
	// Update page based on reaction
	if r.Emoji.Name == "➡️" {
		state.CurrentPage++
		if state.CurrentPage >= len(state.Pages) {
			state.CurrentPage = 0
		}
	} else if r.Emoji.Name == "⬅️" {
		state.CurrentPage--
		if state.CurrentPage < 0 {
			state.CurrentPage = len(state.Pages) - 1
		}
	}
	
	// Edit the message with the new page
	_, err := s.ChannelMessageEditEmbed(r.ChannelID, r.MessageID, state.Pages[state.CurrentPage])
	if err != nil {
		log.Printf("Error editing paginated message: %v", err)
		return
	}
	
	// Remove the user's reaction
	err = s.MessageReactionRemove(r.ChannelID, r.MessageID, r.Emoji.Name, r.UserID)
	if err != nil {
		log.Printf("Error removing reaction: %v", err)
	}
}