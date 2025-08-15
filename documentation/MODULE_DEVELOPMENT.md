# 🚀 Module Development Guide

## Overview

This Discord bot uses a **fully modular architecture** that allows developers to create independent modules without touching core files. Each module is self-contained and automatically discovered by the system.

## 📁 Module Structure

```
commands/
├── mymodule/
│   ├── module.go          # Module definition and registration
│   ├── command1.go        # Command implementations
│   ├── command2.go        # More command implementations
│   └── slash_commands.go  # Slash command implementations (optional)
```

## 🔧 Creating a Module

### 1. Create Module Directory

Create a new directory under `commands/` for your module:

```bash
mkdir commands/mymodule
```

### 2. Create module.go

This is the core file that defines your module:

```go
package mymodule

import (
	"DiscordBot/commands"
	"github.com/bwmarrin/discordgo"
)

func init() {
	module := &commands.ModuleInfo{
		Name:        "MyModule",
		Description: "Description of what this module does",
		Version:     "1.0.0",
		Author:      "Your Name",
		Category:    "Custom", // Category for organization
		Dependencies: []string{}, // Other modules this depends on
		Config: map[string]interface{}{
			"setting1": "default_value",
			"setting2": 42,
		},
		Commands: []commands.CommandInfo{
			{
				Name:        "mycommand",
				Aliases:     []string{"mc", "mycmd"},
				Description: "What this command does",
				Usage:       ".mycommand <args>",
				Category:    "Custom",
			},
			{
				Name:        "anothercommand",
				Aliases:     []string{},
				Description: "Another command description",
				Usage:       ".anothercommand",
				Category:    "Custom",
			},
		},
		SlashCommands: []commands.SlashCommandInfo{
			{
				Name:        "myslash",
				Description: "My slash command",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "input",
						Description: "Input parameter",
						Required:    true,
					},
				},
				Handler: MySlashHandler,
			},
		},
	}
	
	commands.RegisterModule(module)

	// Register command handlers
	commands.RegisterCommand("mycommand", MyCommand, "mc", "mycmd")
	commands.RegisterCommand("anothercommand", AnotherCommand)
}
```

### 3. Implement Command Functions

Create your command handler functions:

```go
// mycommand.go
package mymodule

import (
	"DiscordBot/bot"
	"github.com/bwmarrin/discordgo"
)

func MyCommand(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// Ensure command is used in a guild
	if m.GuildID == "" {
		return
	}

	// Your command logic here
	s.ChannelMessageSend(m.ChannelID, "Hello from my custom command!")
}

func AnotherCommand(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// Another command implementation
	s.ChannelMessageSend(m.ChannelID, "This is another command!")
}
```

### 4. Implement Slash Commands (Optional)

```go
// slash_commands.go
package mymodule

import (
	"DiscordBot/bot"
	"github.com/bwmarrin/discordgo"
)

func MySlashHandler(b *bot.Bot, s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	input := options[0].StringValue()

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "You said: " + input,
		},
	})
}
```

## 📋 Module Categories

Organize your modules into logical categories:

- **General** - Basic utility commands
- **Economy** - Virtual economy features
- **Moderation** - Server moderation tools
- **Sports** - Sports-related commands
- **Finance** - Financial data and tools
- **Egypt** - Region-specific commands
- **Admin** - Administrative functions
- **Roles** - Role management
- **Custom** - Your custom category

## 🔄 Module Registration Process

1. **Module Registration**: `commands.RegisterModule(module)` registers metadata
2. **Command Registration**: `commands.RegisterCommand()` registers handlers
3. **Auto-Discovery**: CLI and help system automatically discover your module
4. **Auto-Compilation**: Command info is compiled into global registries

## 🎯 Best Practices

### Command Design
- Use clear, descriptive command names
- Provide helpful usage examples
- Include appropriate aliases for common commands
- Handle errors gracefully
- Validate user input

### Code Quality
- Follow Go naming conventions
- Add comments for complex logic
- Handle edge cases
- Use consistent error messages
- Implement proper permission checks

### Module Organization
- Keep related commands in the same module
- Use logical categories
- Minimize dependencies between modules
- Include configuration options for customization

## 🔍 Example: Complete Module

```go
// commands/weather/module.go
package weather

import (
	"DiscordBot/commands"
	"DiscordBot/bot"
	"github.com/bwmarrin/discordgo"
	"fmt"
	"net/http"
	"encoding/json"
)

func init() {
	module := &commands.ModuleInfo{
		Name:        "Weather",
		Description: "Weather information commands",
		Version:     "1.0.0",
		Author:      "Weather Team",
		Category:    "Utility",
		Dependencies: []string{},
		Config: map[string]interface{}{
			"api_key": "",
			"default_units": "metric",
		},
		Commands: []commands.CommandInfo{
			{
				Name:        "weather",
				Aliases:     []string{"w"},
				Description: "Get weather information for a city",
				Usage:       ".weather <city>",
				Category:    "Utility",
			},
		},
		SlashCommands: []commands.SlashCommandInfo{},
	}
	
	commands.RegisterModule(module)
	commands.RegisterCommand("weather", WeatherCommand, "w")
}

func WeatherCommand(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if m.GuildID == "" {
		return
	}

	if len(args) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Usage: .weather <city>")
		return
	}

	city := strings.Join(args[1:], " ")
	
	// Fetch weather data (implement your API call here)
	weather := fmt.Sprintf("Weather in %s: Sunny, 25°C", city)
	
	embed := &discordgo.MessageEmbed{
		Title:       "Weather Information",
		Description: weather,
		Color:       0x00ff00,
	}
	
	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}
```

## 🚀 Testing Your Module

1. **Add Import**: Add your module import to `main.go`:
   ```go
   _ "DiscordBot/commands/mymodule"
   ```

2. **Build and Test**:
   ```bash
   go build
   ./DiscordBot
   ```

3. **Test Commands**: Try your commands in Discord:
   ```
   .mycommand
   .help mycommand
   .commandlist
   ```

## 🔧 Advanced Features

### Database Access
```go
func MyCommand(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// Access database through bot instance
	var count int
	err := b.Db.QueryRow("SELECT COUNT(*) FROM users WHERE guild_id = $1", m.GuildID).Scan(&count)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Database error")
		return
	}
	
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("User count: %d", count))
}
```

### Permission Checks
```go
func AdminCommand(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// Check if user is admin
	isAdmin, err := b.IsAdmin(m.GuildID, m.Author.ID)
	if err != nil || !isAdmin {
		s.ChannelMessageSend(m.ChannelID, "You need admin permissions for this command.")
		return
	}
	
	// Admin-only logic here
}
```

### Configuration Access
```go
func ConfigurableCommand(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// Access module configuration
	module := commands.RegisteredModules["MyModule"]
	setting := module.Config["setting1"].(string)
	
	s.ChannelMessageSend(m.ChannelID, "Setting value: " + setting)
}
```

## 📚 API Reference

### Core Types

```go
type ModuleInfo struct {
	Name          string
	Description   string
	Version       string
	Author        string
	Category      string
	Commands      []CommandInfo
	SlashCommands []SlashCommandInfo
	Dependencies  []string
	Config        map[string]interface{}
}

type CommandInfo struct {
	Name        string
	Aliases     []string
	Description string
	Usage       string
	Category    string
}

type SlashCommandInfo struct {
	Name        string
	Description string
	Options     []*discordgo.ApplicationCommandOption
	Handler     func(*bot.Bot, *discordgo.Session, *discordgo.InteractionCreate)
}
```

### Core Functions

```go
// Register your module
commands.RegisterModule(module *ModuleInfo)

// Register command handlers
commands.RegisterCommand(name string, handler CommandFunc, aliases ...string)

// Access registered modules
commands.RegisteredModules[moduleName]

// Access command details
commands.CommandDetails[commandName]
```

## 🎉 That's It!

Your module is now fully integrated and will be:
- ✅ Automatically discovered by the CLI tool
- ✅ Included in help system
- ✅ Available for bot creation
- ✅ Listed in command lists
- ✅ Slash commands auto-registered

Happy coding! 🚀