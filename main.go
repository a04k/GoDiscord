package main

import (
	"log"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"DiscordBot/bot"
	"DiscordBot/commands"
	"DiscordBot/commands/admin"
	"DiscordBot/commands/economy"
	"DiscordBot/commands/moderation"
	"DiscordBot/commands/roles"
	"DiscordBot/commands/slash"
)

func main() {
	godotenv.Load()

	bot, err := bot.NewBot(os.Getenv("DISCORD_TOKEN"), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}

	bot.Client.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}

		if !strings.HasPrefix(m.Content, ".") {
			return
		}

		args := strings.Fields(m.Content)
		cmd := strings.ToLower(args[0][1:])

		switch cmd {
		case "help", "h":
			commands.Help(bot, s, m, args)
		case "commandlist", "cl":
			commands.CommandList(bot, s, m, args)
		case "usd":
			commands.USD(bot, s, m, args)
		case "btc":
			commands.BTC(bot, s, m, args)
		case "bal", "balance":
			economy.Balance(bot, s, m, args)
		case "work":
			economy.Work(bot, s, m, args)
		case "transfer":
			economy.Transfer(bot, s, m, args)
		case "flip":
			economy.Flip(bot, s, m, args)
		case "kick":
			admin.Kick(bot, s, m, args)
		case "mute", "m":
			moderation.Mute(bot, s, m, args)
		case "unmute", "um":
			moderation.Unmute(bot, s, m, args)
		case "voicemute", "vm":
			moderation.VoiceMute(bot, s, m, args)
		case "vunmute", "vum":
			moderation.VoiceUnmute(bot, s, m, args)
		case "ban":
			admin.Ban(bot, s, m, args)
		case "unban":
			admin.Unban(bot, s, m, args)
		case "setadmin", "sa":
			admin.SetAdmin(bot, s, m, args)
		case "add":
			admin.Add(bot, s, m, args)
		case "createrole", "cr":
			roles.CreateRole(bot, s, m, args)
		case "setrole", "sr":
			roles.SetRole(bot, s, m, args)
		case "inrole":
			roles.InRole(bot, s, m, args)
		case "roleinfo", "ri":
			roles.RoleInfo(bot, s, m, args)
		}
	})

	bot.Client.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type != discordgo.InteractionApplicationCommand {
			return
		}

		switch i.ApplicationCommandData().Name {
		case "quota":
			slash.Quota(bot, s, i)
		case "setup":
			slash.Setup(bot, s, i)
		}
	})

	err = bot.Client.Open()
	if err != nil {
		log.Fatalf("error opening connection to Discord: %v", err)
	}

	defer bot.Client.Close()

	log.Println("Bot is now running. Press CTRL-C to exit.")
	// Wait here until the program is interrupted
	select {}
}