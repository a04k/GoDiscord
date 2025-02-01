package commands

import (
	"database/sql"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"WeQuota"
)

func HandleQuotaCommand(s *discordgo.Session, i *discordgo.InteractionCreate, db *sql.DB) error {
	var landline, password string
	err := db.QueryRow("SELECT landline, password FROM credentials WHERE user_id = $1", i.Member.User.ID).Scan(&landline, &password)

	if err == sql.ErrNoRows {
		return respondEphemeral(s, i, "Please set up your credentials first using /setup")
	} else if err != nil {
		return respondEphemeral(s, i, "Error retrieving credentials")
	}

	checker, err := WeQuota.NewWeQuotaChecker(landline, password)
	if err != nil {
		return respondEphemeral(s, i, fmt.Sprintf("Error: %v", err))
	}

	quota, err := checker.CheckQuota()
	if err != nil {
		return respondEphemeral(s, i, fmt.Sprintf("Error: %v", err))
	}

	msg := fmt.Sprintf(`
		Customer: %s
		Plan: %s
		Remaining: %.2f / %.2f (%s%% Used)
		Renewed: %s
		Expires: %s (%s)`,
		quota["name"],
		quota["offerName"],
		quota["remaining"],
		quota["total"],
		quota["usagePercentage"],
		quota["renewalDate"],
		quota["expiryDate"],
		quota["expiryIn"])

	return respondEphemeral(s, i, msg)
}
