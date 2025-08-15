# Discord Bot CLI

A modular Discord bot CLI tool that allows you to create custom bot instances by selecting modules and commands, similar to create-next-app.

## Installation

```bash
go build -o botcli
```

## Usage

### Create a new bot

```bash
# Interactive creation
./botcli create

# Create with name
./botcli create my-discord-bot
```

This will:
1. Discover available modules
2. Let you select which modules to include
3. Let you choose specific commands within modules
4. Generate a complete bot project

### List available modules and commands

```bash
# List everything
./botcli list

# List only modules
./botcli list --modules

# List only commands
./botcli list --commands

# Filter by module
./botcli list --filter Economy
```

### Add modules to existing bot

```bash
# Add from Git repository
./botcli add https://github.com/user/bot-module

# Add from local directory
./botcli add ./my-local-module
```

### Build the bot

```bash
# Build for current platform
./botcli build

# Build for specific platform
./botcli build --os linux --arch amd64

# Custom output name
./botcli build --output my-bot
```

## Module Development

### Creating a Module

1. Create a directory structure:
```
commands/
  mymodule/
    module.go
    commands.go
```

2. Define your module in `module.go`:
```go
package mymodule

import "DiscordBot/commands"

func init() {
    module := &commands.ModuleInfo{
        Name:        "MyModule",
        Description: "Description of what this module does",
        Version:     "1.0.0",
        Author:      "Your Name",
        Commands: []commands.CommandInfo{
            {
                Name:        "mycommand",
                Aliases:     []string{"mc"},
                Description: "What this command does",
                Usage:       ".mycommand <args>",
                Category:    "MyCategory",
                Handler:     MyCommandHandler,
            },
        },
    }
    
    commands.RegisterModule(module)
}
```

3. Implement your command handlers in `commands.go`:
```go
package mymodule

import (
    "DiscordBot/bot"
    "github.com/bwmarrin/discordgo"
)

func MyCommandHandler(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
    // Your command logic here
}
```

### Module Guidelines

- Each module should be self-contained
- Use clear, descriptive names for commands
- Include proper usage examples
- Handle errors gracefully
- Follow Go naming conventions
- Add appropriate aliases for common commands

## Project Structure

After creating a bot, you'll have:

```
my-bot/
├── bot.config.json     # Bot configuration
├── main.go            # Bot entry point
├── go.mod             # Go module file
├── .env.example       # Environment variables template
├── bot/               # Bot core package
└── commands/          # Selected modules
    ├── module.go      # Command registry
    ├── economy/       # Economy module (if selected)
    └── f1/           # F1 module (if selected)
```

## Configuration

The `bot.config.json` file contains:
- Selected modules
- Enabled/disabled commands
- Module-specific configuration

## Environment Variables

Copy `.env.example` to `.env` and configure:
- `DISCORD_TOKEN`: Your Discord bot token
- `DATABASE_URL`: PostgreSQL connection string

## Contributing

1. Create your module following the guidelines above
2. Test it with the CLI tool
3. Submit a pull request with your module

The modular architecture means you only need to understand the public API to contribute new functionality!