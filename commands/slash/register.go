package slash

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

// registers and updates slash commands.
func RegisterCommands(s *discordgo.Session, guildID string) {
	log.Println("Registering slash commands...")

	// Fetch existing commands
	existingCommands, err := s.ApplicationCommands(s.State.User.ID, guildID)
	if err != nil {
		log.Fatalf("Could not fetch existing commands: %v", err)
	}

	// Define the commands we want
	desiredCommands := []*discordgo.ApplicationCommand{
		{
			Name:        "f1",
			Description: "Toggle F1 weekend and session notifications",
		},
		{
			Name:        "wequota",
			Description: "Check your WE internet quota",
		},
		{
			Name:        "wesetup",
			Description: "Set up your WE account credentials",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "landline",
					Description: "Your WE landline number",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "password",
					Description: "Your WE account password",
					Required:    true,
				},
			},
		},
	}

	// Create a map of existing commands for easy lookup
	existingCommandMap := make(map[string]*discordgo.ApplicationCommand)
	for _, cmd := range existingCommands {
		existingCommandMap[cmd.Name] = cmd
	}

	// Process each desired command
	for _, desiredCmd := range desiredCommands {
		existingCmd, exists := existingCommandMap[desiredCmd.Name]
		
		if !exists {
			// Command doesn't exist, create it
			_, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, desiredCmd)
			if err != nil {
				log.Fatalf("Cannot create slash command %q: %v", desiredCmd.Name, err)
			}
			log.Printf("Created command: %s", desiredCmd.Name)
		} else {
			// Command exists, check if it needs to be updated
			if commandNeedsUpdate(existingCmd, desiredCmd) {
				// Delete the existing command
				err := s.ApplicationCommandDelete(s.State.User.ID, guildID, existingCmd.ID)
				if err != nil {
					log.Fatalf("Could not delete command %q: %v", existingCmd.Name, err)
				}
				log.Printf("Deleted outdated command: %s", existingCmd.Name)
				
				// Create the updated command
				_, err = s.ApplicationCommandCreate(s.State.User.ID, guildID, desiredCmd)
				if err != nil {
					log.Fatalf("Cannot create updated slash command %q: %v", desiredCmd.Name, err)
				}
				log.Printf("Created updated command: %s", desiredCmd.Name)
			} else {
				// Command is up to date
				log.Printf("Command %s is already up to date", desiredCmd.Name)
			}
		}
	}

	// Check for commands that should be removed (exist but are not desired)
	for _, existingCmd := range existingCommands {
		found := false
		for _, desiredCmd := range desiredCommands {
			if existingCmd.Name == desiredCmd.Name {
				found = true
				break
			}
		}
		
		if !found {
			// This command should be removed
			err := s.ApplicationCommandDelete(s.State.User.ID, guildID, existingCmd.ID)
			if err != nil {
				log.Fatalf("Could not delete obsolete command %q: %v", existingCmd.Name, err)
			}
			log.Printf("Deleted obsolete command: %s", existingCmd.Name)
		}
	}

	log.Println("Slash commands registered successfully.")
}

// commandNeedsUpdate checks if an existing command needs to be updated
func commandNeedsUpdate(existing, desired *discordgo.ApplicationCommand) bool {
	// Check basic properties
	if existing.Description != desired.Description {
		return true
	}
	
	// Check options count
	if len(existing.Options) != len(desired.Options) {
		return true
	}
	
	// Check each option
	for i, existingOption := range existing.Options {
		desiredOption := desired.Options[i]
		
		if existingOption.Type != desiredOption.Type ||
			existingOption.Name != desiredOption.Name ||
			existingOption.Description != desiredOption.Description ||
			existingOption.Required != desiredOption.Required {
			return true
		}
	}
	
	return false
}
