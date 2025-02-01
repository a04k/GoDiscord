package commands

import (
	"database/sql"
	"github.com/bwmarrin/discordgo"
)
func HandleSlashCommands(s *discordgo.Session, i *discordgo.InteractionCreate, db *sql.DB) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	switch i.ApplicationCommandData().Name {
	case "quota":
		HandleQuotaCommand(s, i, db)
	case "setup":
		HandleSetupCommand(s, i, db)
	}
}