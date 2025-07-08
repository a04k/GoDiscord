package slash

import (
	"database/sql"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
	"QCheckWE"
)

func Quota(b *bot.Bot, s *discordgo.Session, i *discordgo.InteractionCreate) {
	// First try to get saved credentials
	var landline, password string
	err := b.Db.QueryRow("SELECT landline, password FROM credentials WHERE user_id = $1", i.Member.User.ID).Scan(&landline, &password)

	if err == sql.ErrNoRows {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Please set up your credentials first using /setup",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	} else if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Error retrieving credentials",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	checker, err := QCheckWE.NewWeQuotaChecker(landline, password)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error: %v", err),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	quota, err := checker.CheckQuota()
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error: %v", err),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
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

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
