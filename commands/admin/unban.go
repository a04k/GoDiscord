package admin

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"DiscordBot/bot"
)

func Unban(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) >= 2 {
		targetUser := args[1]

		// exec remove ban => guildbandelete
		err := s.GuildBanDelete(m.GuildID, targetUser)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Failed to unban user: %s", err))
			return
		}
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Unbanned user <@%s>", targetUser))
	} else {
		s.ChannelMessageSend(m.ChannelID, "Please specify a user ID to unban.")
	}
}
