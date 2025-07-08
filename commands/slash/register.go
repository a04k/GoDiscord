package slash

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

// RegisterCommands registers and updates slash commands.
func RegisterCommands(s *discordgo.Session, guildID string) {
	log.Println("Registering slash commands...")

	// Fetch existing commands
	existingCommands, err := s.ApplicationCommands(s.State.User.ID, guildID)
	if err != nil {
		log.Fatalf("Could not fetch existing commands: %v", err)
	}

	// Delete all existing commands
	for _, cmd := range existingCommands {
		err := s.ApplicationCommandDelete(s.State.User.ID, guildID, cmd.ID)
		if err != nil {
			log.Fatalf("Could not delete command %q: %v", cmd.Name, err)
		}
		log.Printf("Deleted command: %s", cmd.Name)
	}

	// Define new commands
	commands := []*discordgo.ApplicationCommand{
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

	// Register new commands
	for _, cmd := range commands {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, cmd)
		if err != nil {
			log.Fatalf("Cannot create slash command %q: %v", cmd.Name, err)
		}
		log.Printf("Created command: %s", cmd.Name)
	}

	log.Println("Slash commands registered successfully.")
}
