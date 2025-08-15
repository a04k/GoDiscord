package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type BotConfig struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Modules     []string               `json:"modules"`
	Commands    map[string]bool        `json:"commands"` // command_name: enabled
	Config      map[string]interface{} `json:"config"`
}

var createCmd = &cobra.Command{
	Use:   "create [bot-name]",
	Short: "Create a new Discord bot with selected modules",
	Long: `Create a new Discord bot instance by selecting modules and commands.
Similar to create-next-app, this will generate a customized bot with only the features you need.`,
	Args: cobra.MaximumNArgs(1),
	Run:  runCreate,
}

func runCreate(cmd *cobra.Command, args []string) {
	var botName string

	if len(args) > 0 {
		botName = args[0]
	} else {
		botName = promptInput("Bot name")
	}

	if botName == "" {
		fmt.Println("Bot name is required")
		os.Exit(1)
	}

	// Create bot directory
	botDir := filepath.Join(".", botName)
	if _, err := os.Stat(botDir); !os.IsNotExist(err) {
		fmt.Printf("Directory %s already exists\n", botName)
		os.Exit(1)
	}

	fmt.Printf("Creating Discord bot: %s\n\n", botName)

	// Discover available modules
	modules, err := discoverModulesSimple()
	if err != nil {
		fmt.Printf("Error discovering modules: %v\n", err)
		os.Exit(1)
	}

	// Interactive module selection
	selectedModules := selectModules(modules)

	// Interactive command selection within modules
	selectedCommands := selectCommands(selectedModules, modules)

	// Create bot configuration
	config := BotConfig{
		Name:        botName,
		Description: promptInput("Bot description (optional)"),
		Modules:     selectedModules,
		Commands:    selectedCommands,
		Config:      make(map[string]interface{}),
	}

	// Generate bot files
	err = generateBot(botDir, config)
	if err != nil {
		fmt.Printf("Error generating bot: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✅ Successfully created bot '%s'\n", botName)
	fmt.Printf("📁 Location: %s\n", botDir)
	fmt.Println("\nNext steps:")
	fmt.Printf("  cd %s\n", botName)
	fmt.Println("  cp .env.example .env")
	fmt.Println("  # Edit .env with your Discord token and database URL")
	fmt.Println("  go mod tidy")
	fmt.Println("  go build -o bot")
	fmt.Println("  ./bot")
}

func promptInput(prompt string) string {
	fmt.Printf("%s: ", prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func promptYesNo(prompt string) bool {
	for {
		fmt.Printf("%s (y/n): ", prompt)
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.ToLower(strings.TrimSpace(input))

		if input == "y" || input == "yes" {
			return true
		} else if input == "n" || input == "no" {
			return false
		}
		fmt.Println("Please enter 'y' or 'n'")
	}
}

func selectModules(modules map[string]ModuleInfo) []string {
	fmt.Println("📦 Available modules organized by category:")
	fmt.Println()

	// Group modules by category
	categories := make(map[string][]ModuleInfo)
	for _, module := range modules {
		category := module.Category
		if category == "" {
			category = "Other"
		}
		categories[category] = append(categories[category], module)
	}

	var selected []string

	// Display by category
	for categoryName, categoryModules := range categories {
		fmt.Printf("📂 %s Category:\n", categoryName)

		for _, module := range categoryModules {
			fmt.Printf("  📦 %s - %s\n", module.Name, module.Description)
			fmt.Printf("     Commands: %d", len(module.Commands))
			if len(module.Dependencies) > 0 {
				fmt.Printf(" | Dependencies: %v", module.Dependencies)
			}
			fmt.Println()

			if promptYesNo(fmt.Sprintf("Include %s module?", module.Name)) {
				selected = append(selected, module.Name)
			}
		}
		fmt.Println()
	}

	return selected
}

func selectCommands(selectedModules []string, modules map[string]ModuleInfo) map[string]bool {
	commands := make(map[string]bool)

	fmt.Println("🔧 Command selection:")
	fmt.Println("You can enable/disable individual commands within selected modules")
	fmt.Println()

	for _, moduleName := range selectedModules {
		module := modules[moduleName]
		fmt.Printf("📦 %s module commands:\n", moduleName)

		for _, cmd := range module.Commands {
			fmt.Printf("  .%s", cmd.Name)
			if len(cmd.Aliases) > 0 {
				fmt.Printf(" (aliases: %s)", strings.Join(cmd.Aliases, ", "))
			}
			fmt.Printf(" - %s\n", cmd.Description)

			enabled := promptYesNo(fmt.Sprintf("Enable .%s command?", cmd.Name))
			commands[cmd.Name] = enabled
		}
		fmt.Println()
	}

	return commands
}

func generateBot(botDir string, config BotConfig) error {
	// Create directory structure
	err := os.MkdirAll(botDir, 0755)
	if err != nil {
		return err
	}

	// Save bot configuration
	configFile := filepath.Join(botDir, "bot.config.json")
	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(configFile, configData, 0644)
	if err != nil {
		return err
	}

	// Generate main.go
	err = generateMainFile(botDir, config)
	if err != nil {
		return err
	}

	// Generate go.mod
	err = generateGoMod(botDir, config)
	if err != nil {
		return err
	}

	// Generate .env.example
	err = generateEnvExample(botDir)
	if err != nil {
		return err
	}

	// Copy selected module files
	err = copyModuleFiles(botDir, config)
	if err != nil {
		return err
	}

	return nil
}

func generateMainFile(botDir string, config BotConfig) error {
	mainContent := fmt.Sprintf(`package main

import (
	"database/sql"
	"log"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	
	"%s/bot"
	"%s/commands"
	
	// Import selected modules
`, config.Name, config.Name)

	// Add imports for selected modules
	for _, moduleName := range config.Modules {
		mainContent += fmt.Sprintf("\t_ \"%s/commands/%s\"\n", config.Name, strings.ToLower(moduleName))
	}

	mainContent += `
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

		// Prevent commands from being used in DMs
		if m.GuildID == "" {
			return
		}

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

		// Check if command is enabled
		if !isCommandEnabled(cmdName) {
			return
		}

		if handler, ok := commands.CommandMap[cmdName]; ok {
			handler(bot, s, m, args)
		}
	})

	err = bot.Client.Open()
	if err != nil {
		log.Fatalf("error opening connection to Discord: %v", err)
	}

	defer bot.Client.Close()

	log.Println("Bot is now running. Press CTRL-C to exit.")
	select {}
}

// isCommandEnabled checks if a command is enabled in the bot configuration
func isCommandEnabled(commandName string) bool {
	// This would be loaded from bot.config.json in a real implementation
	// For now, return true for all commands
	return true
}
`

	return os.WriteFile(filepath.Join(botDir, "main.go"), []byte(mainContent), 0644)
}

func generateGoMod(botDir string, config BotConfig) error {
	goModContent := fmt.Sprintf(`module %s

go 1.21

require (
	github.com/bwmarrin/discordgo v0.27.1
	github.com/joho/godotenv v1.4.0
	github.com/lib/pq v1.10.9
)

require (
	github.com/gorilla/websocket v1.4.2 // indirect
	golang.org/x/crypto v0.0.0-20210421170649-83a5a9bb288b // indirect
	golang.org/x/sys v0.0.0-20201119102817-f84b799fce68 // indirect
)
`, config.Name)

	return os.WriteFile(filepath.Join(botDir, "go.mod"), []byte(goModContent), 0644)
}

func generateEnvExample(botDir string) error {
	envContent := `# Discord Bot Configuration
DISCORD_TOKEN=your_discord_bot_token_here
DATABASE_URL=postgres://username:password@localhost/dbname?sslmode=disable

# Optional: Bot Owner ID (for admin commands)
BOT_OWNER_ID=your_discord_user_id
`

	return os.WriteFile(filepath.Join(botDir, ".env.example"), []byte(envContent), 0644)
}

func copyModuleFiles(botDir string, config BotConfig) error {
	// Create commands directory
	commandsDir := filepath.Join(botDir, "commands")
	err := os.MkdirAll(commandsDir, 0755)
	if err != nil {
		return err
	}

	// Copy base command files
	err = copyBaseFiles(commandsDir)
	if err != nil {
		return err
	}

	// Copy selected module files
	for _, moduleName := range config.Modules {
		sourceDir := filepath.Join("commands", strings.ToLower(moduleName))
		destDir := filepath.Join(commandsDir, strings.ToLower(moduleName))

		err = copyDirectory(sourceDir, destDir)
		if err != nil {
			fmt.Printf("Warning: Failed to copy module %s: %v\n", moduleName, err)
		}
	}

	// Copy bot package
	err = copyDirectory("bot", filepath.Join(botDir, "bot"))
	if err != nil {
		return err
	}

	return nil
}

func copyBaseFiles(commandsDir string) error {
	// Copy module.go (the registry)
	moduleContent := `package commands

import (
	"DiscordBot/bot"
	"github.com/bwmarrin/discordgo"
)

type CommandFunc func(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string)

type CommandInfo struct {
	Name        string
	Aliases     []string
	Description string
	Usage       string
	Category    string
	Handler     CommandFunc
}

type ModuleInfo struct {
	Name        string
	Description string
	Version     string
	Author      string
	Commands    []CommandInfo
	Dependencies []string
	Config      map[string]interface{}
}

var (
	CommandMap       = make(map[string]CommandFunc)
	CommandAliases   = make(map[string]string)
	RegisteredModules = make(map[string]*ModuleInfo)
	CommandInfoMap   = make(map[string]*CommandInfo)
)

func RegisterModule(module *ModuleInfo) {
	RegisteredModules[module.Name] = module
	
	for _, cmd := range module.Commands {
		RegisterCommand(cmd.Name, cmd.Handler, cmd.Aliases...)
		CommandInfoMap[cmd.Name] = &cmd
	}
}

func RegisterCommand(name string, handler CommandFunc, aliases ...string) {
	CommandMap[name] = handler
	for _, alias := range aliases {
		CommandAliases[alias] = name
	}
}
`

	return os.WriteFile(filepath.Join(commandsDir, "module.go"), []byte(moduleContent), 0644)
}

func copyDirectory(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate destination path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		// Copy file
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(destPath, data, info.Mode())
	})
}
