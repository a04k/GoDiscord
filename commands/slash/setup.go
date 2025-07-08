package slash

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
	"QCheckWE"
)

func WeAccountSetup(b *bot.Bot, s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	landline := options[0].StringValue()
	password := options[1].StringValue()

	// test the credentials first
	checker, err := QCheckWE.NewWeQuotaChecker(landline, password)
	if err != nil {
		log.Printf("Error creating WE Quota Checker: %v", err)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Invalid credentials",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// test if we can actually get the quota
	_, err = checker.CheckQuota()
	if err != nil {
		log.Printf("Error checking quota: %v", err)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Invalid credentials",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// Save the credentials
	_, err = b.Db.Exec(`
                INSERT INTO credentials (user_id, landline, password)
                VALUES ($1, $2, $3)
                ON CONFLICT (user_id)
                DO UPDATE SET landline = $2, password = $3
            `, i.Member.User.ID, landline, password)

	if err != nil {
		log.Printf("Error saving credentials: %v", err)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Error saving credentials",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Credentials saved successfully! You can now use /quota",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
