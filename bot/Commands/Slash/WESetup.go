package commands

import (
	"database/sql"
	"log"
	"github.com/bwmarrin/discordgo"
	"WeQuota" 
)

func HandleSetupCommand(s *discordgo.Session, i *discordgo.InteractionCreate, db *sql.DB) error {
	options := i.ApplicationCommandData().Options
	landline := options[0].StringValue()
	password := options[1].StringValue()

	// Test the credentials
	checker, err := WeQuota.NewWeQuotaChecker(landline, password)
	if err != nil {
		log.Printf("Error creating WE Quota Checker: %v", err)
		return respondEphemeral(s, i, "Invalid credentials")
	}

	// Test if we can get the quota
	_, err = checker.CheckQuota()
	if err != nil {
		log.Printf("Error checking quota: %v", err)
		return respondEphemeral(s, i, "Invalid credentials")
	}

	// Save the credentials
	_, err = db.Exec(`
		INSERT INTO credentials (user_id, landline, password)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id)
		DO UPDATE SET landline = $2, password = $3
	`, i.Member.User.ID, landline, password)

	if err != nil {
		log.Printf("Error saving credentials: %v", err)
		return respondEphemeral(s, i, "Error saving credentials")
	}

	return respondEphemeral(s, i, "Credentials saved successfully! You can now use /quota")
}
