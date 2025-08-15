package commands

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

// commandNeedsUpdate checks if an existing command needs to be updated
func commandNeedsUpdate(existing, desired *discordgo.ApplicationCommand) bool {
	if existing.Name != desired.Name {
		return true
	}
	if existing.Description != desired.Description {
		return true
	}
	if len(existing.Options) != len(desired.Options) {
		return true
	}
	for i, option := range existing.Options {
		if i >= len(desired.Options) {
			return true
		}
		desiredOption := desired.Options[i]
		if option.Name != desiredOption.Name ||
			option.Description != desiredOption.Description ||
			option.Type != desiredOption.Type ||
			option.Required != desiredOption.Required {
			return true
		}
	}
	return false
}

// RegisterAllSlashCommands registers and updates slash commands from all modules
func RegisterAllSlashCommands(s *discordgo.Session, guildID string) {
	// Get existing commands
	existingCommands, err := s.ApplicationCommands(s.State.User.ID, guildID)
	if err != nil {
		log.Printf("Error fetching existing commands: %v", err)
		return
	}

	// Create a map of existing commands for quick lookup
	existingMap := make(map[string]*discordgo.ApplicationCommand)
	for _, cmd := range existingCommands {
		existingMap[cmd.Name] = cmd
	}

	// Get all slash commands from modules
	desiredCommands := GetAllSlashCommands()

	// Register or update commands
	for _, desired := range desiredCommands {
		if existing, exists := existingMap[desired.Name]; exists {
			// Command exists, check if it needs updating
			if commandNeedsUpdate(existing, desired) {
				log.Printf("Updating slash command: %s", desired.Name)
				_, err := s.ApplicationCommandEdit(s.State.User.ID, guildID, existing.ID, desired)
				if err != nil {
					log.Printf("Error updating command %s: %v", desired.Name, err)
				}
			}
			// Remove from existing map so we know it's still wanted
			delete(existingMap, desired.Name)
		} else {
			// Command doesn't exist, create it
			log.Printf("Creating slash command: %s", desired.Name)
			_, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, desired)
			if err != nil {
				log.Printf("Error creating command %s: %v", desired.Name, err)
			}
		}
	}

	// Delete any remaining commands that are no longer wanted
	for _, cmd := range existingMap {
		log.Printf("Deleting unused slash command: %s", cmd.Name)
		err := s.ApplicationCommandDelete(s.State.User.ID, guildID, cmd.ID)
		if err != nil {
			log.Printf("Error deleting command %s: %v", cmd.Name, err)
		}
	}
}
