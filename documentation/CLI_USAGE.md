# 🛠️ CLI Tool Usage Guide

## Overview

The Discord Bot CLI tool allows you to create custom bot instances by selecting modules and commands, similar to `create-next-app`. It provides a modular approach to bot development and deployment.

## 🚀 Installation

### Prerequisites
- Go 1.21 or higher
- Git (for cloning modules)

### Build the CLI

```bash
cd cli
go mod tidy
go build -o botcli
```

On Windows:
```bash
go build -o botcli.exe
```

## 📋 Available Commands

### `create` - Create a New Bot

Create a new Discord bot instance with selected modules:

```bash
# Interactive creation
./botcli create

# Create with specific name
./botcli create my-discord-bot
```

**What it does:**
1. Discovers all available modules
2. Presents them organized by category
3. Lets you select which modules to include
4. Lets you choose specific commands within modules
5. Generates a complete bot project

**Example Output:**
```
Creating Discord bot: my-discord-bot

📦 Available modules organized by category:

📂 Sports Category:
  📦 F1 - Formula 1 race results, standings, and live notifications
     Commands: 8 | Dependencies: []
Include F1 module? (y/n): y

  📦 EPL - English Premier League table and fixture information
     Commands: 2 | Dependencies: []
Include EPL module? (y/n): n

📂 Economy Category:
  📦 Economy - Virtual economy system with balance, work, and gambling commands
     Commands: 7 | Dependencies: []
Include Economy module? (y/n): y

🔧 Command selection:
You can enable/disable individual commands within selected modules

📦 F1 module commands:
  .f1 - Shows current F1 information
Enable .f1 command? (y/n): y

  .f1results - Shows the latest F1 race results
Enable .f1results command? (y/n): y

✅ Successfully created bot 'my-discord-bot'
📁 Location: ./my-discord-bot

Next steps:
  cd my-discord-bot
  cp .env.example .env
  # Edit .env with your Discord token and database URL
  go mod tidy
  go build -o bot
  ./bot
```

### `list` - List Available Modules

Display all available modules and their commands:

```bash
# List everything
./botcli list

# List only modules
./botcli list --modules

# List only commands
./botcli list --commands

# Filter by specific module
./botcli list --filter Economy
```

**Example Output:**
```
📦 Available Modules and Commands:

📦 Economy v1.0.0 - Virtual economy system with balance, work, and gambling commands
   Author: Bot Team
   Commands:
     .balance (bal, $) - Check your current balance
     .work (w) - Work to earn money
     .transfer (pay, give) - Transfer money to another user
     .flip - Flip a coin and gamble your money

📦 F1 v1.2.0 - Formula 1 race results, standings, and live notifications
   Author: Bot Team
   Dependencies: []
   Commands:
     .f1 - Shows current F1 information
     .f1results - Shows the latest F1 race results
     .f1standings - Shows the current F1 championship standings

📊 Summary: 8 modules, 45 commands
```

### `add` - Add Modules to Existing Bot

Add modules to an existing bot project:

```bash
# Add from Git repository
./botcli add https://github.com/user/bot-module

# Add from local directory
./botcli add ./my-local-module
```

**What it does:**
1. Checks if you're in a bot project directory
2. Clones or accesses the module source
3. Discovers modules in the source
4. Lets you select which modules to add
5. Integrates them into your existing bot

**Example Output:**
```
Adding module: https://github.com/user/weather-module

📦 Available modules to add:

  Weather - Weather information commands
    Commands: 3

Add Weather module? (y/n): y

✅ Added module: Weather
🎉 Modules added successfully!
Run 'go mod tidy' to update dependencies
Run 'go build' to rebuild your bot
```

### `build` - Build Bot Binary

Build your bot into an executable:

```bash
# Build for current platform
./botcli build

# Build for specific platform
./botcli build --os linux --arch amd64

# Custom output name
./botcli build --output my-bot

# Build for Windows
./botcli build --os windows --arch amd64 --output my-bot.exe
```

**Example Output:**
```
Building bot: my-discord-bot
✅ Successfully built: my-discord-bot

To run your bot:
  ./my-discord-bot
```

## 📁 Generated Project Structure

After creating a bot, you'll have:

```
my-discord-bot/
├── bot.config.json     # Bot configuration
├── main.go            # Bot entry point
├── go.mod             # Go module file
├── .env.example       # Environment variables template
├── bot/               # Bot core package
└── commands/          # Selected modules
    ├── module.go      # Command registry
    ├── economy/       # Economy module (if selected)
    │   ├── module.go
    │   ├── balance.go
    │   └── work.go
    └── sports/        # Sports modules (if selected)
        └── f1/
            ├── module.go
            ├── f1.go
            └── f1results.go
```

## ⚙️ Configuration Files

### `bot.config.json`

Contains your bot's configuration:

```json
{
  "name": "my-discord-bot",
  "description": "My custom Discord bot",
  "modules": ["Economy", "F1"],
  "commands": {
    "balance": true,
    "work": true,
    "f1": true,
    "f1results": false
  },
  "config": {
    "economy": {
      "starting_balance": 100,
      "work_rewards": [50, 200]
    }
  }
}
```

### `.env` File

Environment variables for your bot:

```env
# Discord Bot Configuration
DISCORD_TOKEN=your_discord_bot_token_here
DATABASE_URL=postgres://username:password@localhost/dbname?sslmode=disable

# Optional: Bot Owner ID (for admin commands)
BOT_OWNER_ID=your_discord_user_id
```

## 🔧 Advanced Usage

### Creating Custom Module Sources

You can create your own module repositories:

1. **Create Repository Structure:**
   ```
   my-bot-modules/
   ├── weather/
   │   ├── module.go
   │   └── weather.go
   └── music/
       ├── module.go
       └── music.go
   ```

2. **Add Modules:**
   ```bash
   ./botcli add https://github.com/myuser/my-bot-modules
   ```

### Module Dependencies

Modules can depend on other modules:

```go
module := &commands.ModuleInfo{
    Name: "AdvancedEconomy",
    Dependencies: []string{"Economy", "Roles"},
    // ...
}
```

The CLI will automatically include dependencies when you select a module.

### Cross-Platform Building

Build for different platforms:

```bash
# Linux
./botcli build --os linux --arch amd64

# Windows
./botcli build --os windows --arch amd64

# macOS
./botcli build --os darwin --arch amd64

# ARM64 (Apple Silicon, Raspberry Pi)
./botcli build --os linux --arch arm64
```

## 🐛 Troubleshooting

### Common Issues

**"Not in a bot project directory"**
- Make sure you're in a directory with `bot.config.json`
- Run `./botcli create` first to create a new bot

**"Error discovering modules"**
- Ensure you're running the CLI from the correct directory
- Check that module files have correct Go syntax

**"Module not found"**
- Verify the Git URL is correct and accessible
- Check that the repository contains valid module files

**Build failures**
- Run `go mod tidy` in your bot directory
- Ensure all dependencies are available
- Check for Go syntax errors in generated files

### Getting Help

```bash
# General help
./botcli --help

# Command-specific help
./botcli create --help
./botcli add --help
./botcli build --help
```

## 🎯 Best Practices

### Bot Development
1. **Start Small**: Begin with a few essential modules
2. **Test Locally**: Always test your bot locally before deployment
3. **Version Control**: Use Git to track your bot's configuration
4. **Environment Variables**: Never commit tokens or sensitive data

### Module Selection
1. **Choose Relevant Modules**: Only include modules you'll actually use
2. **Consider Dependencies**: Be aware of module dependencies
3. **Review Commands**: Disable commands you don't need
4. **Check Permissions**: Ensure your bot has necessary Discord permissions

### Deployment
1. **Environment Setup**: Properly configure your `.env` file
2. **Database Setup**: Ensure your database is accessible
3. **Permissions**: Set up proper Discord bot permissions
4. **Monitoring**: Monitor your bot's performance and logs

## 🚀 Quick Start Example

Here's a complete example of creating and running a bot:

```bash
# 1. Build the CLI
cd cli
go build -o botcli

# 2. Create a new bot
./botcli create my-awesome-bot

# 3. Configure the bot
cd my-awesome-bot
cp .env.example .env
# Edit .env with your Discord token

# 4. Build and run
go mod tidy
go build -o bot
./bot
```

## 📚 Additional Resources

- [Module Development Guide](MODULE_DEVELOPMENT.md)
- [Discord.js Documentation](https://discord.js.org/)
- [Go Documentation](https://golang.org/doc/)

## 🎉 Happy Bot Building!

The CLI tool makes it easy to create custom Discord bots with exactly the features you need. Start building your perfect bot today! 🤖✨