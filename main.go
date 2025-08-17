package main

import (
	"database/sql"
	"log"
	"os"
	"strings"

	"DiscordBot/bot"
	"DiscordBot/commands"

	"DiscordBot/utils"

	// Import modules to register them
	_ "DiscordBot/commands/admin"
	_ "DiscordBot/commands/economy"
	_ "DiscordBot/commands/economy/admin"
	_ "DiscordBot/commands/egypt/currency"
	_ "DiscordBot/commands/egypt/telecom"
	_ "DiscordBot/commands/finance"
	_ "DiscordBot/commands/general"
	"DiscordBot/commands/help"
	_ "DiscordBot/commands/moderation"
	_ "DiscordBot/commands/roles"

	_ "DiscordBot/commands/sports/epl"
	"DiscordBot/commands/sports/f1"
	_ "DiscordBot/commands/sports/fpl"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
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

		// Prevent commands from being used in DMs (except for specific allowed commands)
		if m.GuildID == "" {
			// Allow specific commands in DMs like F1 notifications
			args := strings.Fields(m.Content)
			if len(args) > 0 && strings.HasPrefix(m.Content, ".") {
				cmd := strings.ToLower(args[0][1:])
				// Add commands that are allowed in DMs
				allowedDMCommands := map[string]bool{
					"f1": true, "epl": true,
				}
				if !allowedDMCommands[cmd] {
					return // Ignore commands in DMs that aren't explicitly allowed
				}
			} else {
				return // No command provided or not a command
			}
		}

		// When a message is sent in a guild, we no longer need to ensure the user exists in the database
		// Permissions are now checked directly with Discord

		if !strings.HasPrefix(m.Content, ".") {
			return
		}

		args := strings.Fields(m.Content)
		cmd := strings.ToLower(args[0][1:])

		// Resolve alias first
		cmdName, ok := commands.CommandAliases[cmd]
		if !ok {
			cmdName = cmd
		}

		// Check if the command is disabled in this guild
		if m.GuildID != "" {
			var count int
			// Check for disabled command
			err := bot.Db.QueryRow("SELECT COUNT(*) FROM disabled_commands WHERE guild_id = $1 AND name = $2 AND type = 'command'",
				m.GuildID, cmdName).Scan(&count)
			if err != nil && err != sql.ErrNoRows {
				log.Printf("Error checking if command %s is disabled in guild %s: %v", cmdName, m.GuildID, err)
			} else if count > 0 {
				return // Command is disabled
			}

			// Check if the command's category is disabled using module system
			if module := commands.GetModuleByCommand(cmdName); module != nil {
				err := bot.Db.QueryRow("SELECT COUNT(*) FROM disabled_commands WHERE guild_id = $1 AND name = $2 AND type = 'category'",
					m.GuildID, strings.ToLower(module.Category)).Scan(&count)
				if err != nil && err != sql.ErrNoRows {
					log.Printf("Error checking if category %s is disabled in guild %s: %v", module.Category, m.GuildID, err)
				} else if count > 0 {
					return // Category is disabled
				}
			}
		}

		if handler, ok := commands.CommandMap[cmdName]; ok {
			handler(bot, s, m, args)
		}
	})

	// Add reaction handler for pagination
	bot.Client.AddHandler(func(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
		help.HandlePagination(s, r)
	})

	bot.Client.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type != discordgo.InteractionApplicationCommand {
			return
		}

		// Use modular slash command system
		commandName := i.ApplicationCommandData().Name
		if handler, exists := commands.SlashCommandHandlers[commandName]; exists {
			handler(bot, s, i)
		}
	})

	err = bot.Client.Open()
	if err != nil {
		log.Fatalf("error opening connection to Discord: %v", err)
	}

	// Register slash commands using modular system
	commands.RegisterAllSlashCommands(bot.Client, "")

	// Start F1 Notifier
	f1Notifier := f1.NewF1Notifier(bot.Client, bot.Db)
	go f1Notifier.Start()

	// Start Reminder Service
	reminderService := utils.NewReminderService(bot.Db, bot.Client)
	go reminderService.Start()

	defer bot.Client.Close()

	log.Println("Bot is now running. Press CTRL-C to exit.")
	select {}
}
