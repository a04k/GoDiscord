# 🤖 ModuBot-Go

A **fully modular Discord bot** built in Go with a powerful CLI tool for creating custom bot instances. Choose only the modules you need, just like create-next-app!

---

## 📋 Navigation

| 🚀 [Quick Start](#-quick-start) | 📦 [Modules](#-available-modulescategories) | 🛠️ [CLI Tool](#️-cli-tool-usage) | 🏗️ [Development](#️-development) | 🤝 [Contributing](#-how-to-contribute) | 📚 [Docs](#-documentation) |
|:---:|:---:|:---:|:---:|:---:|:---:|
| Get started fast | Browse modules | Use the CLI | Build modules | Join the project | Read the guides |

---

## ✨ Features

- 🧩 **Fully Modular Architecture** - Pick and choose only the features you need
- 🛠️ **CLI Tool** - Create custom bot instances with interactive module selection
- 🚀 **Easy Development** - Add new modules without touching core files
- 📦 **Self-Contained Modules** - Each module is independent and self-registering
- ⚡ **High Performance** - Efficient command handling and memory management
- 🔧 **Configurable** - Extensive configuration options per module
- 🌐 **Multi-Language Support** - Easy to extend with region-specific modules

## 🚀 Quick Start

### Prerequisites

- Go 1.21 or higher
- PostgreSQL database
- Discord Bot Token

### Using the CLI Tool (Recommended)

1. **Build the CLI tool:**
   ```bash
   cd cli
   go mod tidy
   go build -o botcli
   ```

2. **Create your custom bot:**
   ```bash
   ./botcli create my-awesome-bot
   ```

3. **Configure and run:**
   ```bash
   cd my-awesome-bot
   cp .env.example .env
   # Edit .env with your Discord token and database URL
   go mod tidy
   go build -o bot
   ./bot
   ```

### Manual Setup

1. **Clone and build:**
   ```bash
   git clone https://github.com/yourusername/modubot-go
   cd modubot-go
   go mod tidy
   go build -o bot
   ```

2. **Configure environment:**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Run the bot:**
   ```bash
   ./bot
   ```

## 📋 Available Base Commands

### 🔧 General Commands
- `.help [command]` - Get help information
- `.commandlist` - List all available commands
- `.remindme <duration> <message>` - Set reminders

### 💰 Economy System
- `.balance [@user]` - Check coin balance
- `.work` - Work to earn coins
- `.transfer @user <amount>` - Transfer coins
- `.flip <amount>` - Coin flip gambling

### 👑 Administration
- `.disable <command|category> <name>` - Disable commands/categories
- `.enable <command|category> <name>` - Enable commands/categories
- `.add @user <amount>` - Add coins (admin only)
- `.take @user <amount>` - Remove coins (admin only)

### 🛡️ Moderation
- `.mute @user` - Mute a user
- `.unmute @user` - Unmute a user
- `.voicemute @user` - Voice mute a user
- `.vunmute @user` - Voice unmute a user

### 🎭 Role Management
- `.createrole <name>` - Create a new role
- `.setrole @user <role>` - Assign role to user
- `.removerole @user <role>` - Remove role from user
- `.inrole <role>` - List users in role
- `.roleinfo <role>` - Show role information

## 📦 Available Modules/Categories

### 🏆 Sports
- **F1 Module** - Formula 1 race results, standings, and live notifications
  - `.f1`, `.f1results`, `.f1standings`, `.f1wdc`, `.f1wcc`, `.qualiresults`, `.nextf1session`, `.f1sub`
  - Slash command: `/f1` for subscription management
- **EPL Module** - English Premier League information
  - `.epltable`, `.nextmatch [club]`
- **FPL Module** - Fantasy Premier League integration
  - `.fplstandings`, `.setfplleague <id>`

### 💰 Economy
- **Economy Module** - Virtual currency system
- **Economy Admin** - Administrative economy tools

### 🌍 Regional (Egypt)
- **Currency Module** - USD to EGP conversion (`.usd <amount>`)
- **Telecom Module** - WE internet quota checking
  - Slash commands: `/wequota`, `/wesetup`

### 💹 Finance
- **Finance Module** - Financial market data (`.btc`)

### 🔧 System
- **General Module** - Basic utilities
- **Help Module** - Help system and command listing
- **Admin Module** - Bot administration
- **Moderation Module** - Server moderation tools
- **Roles Module** - Role management system

## 🛠️ CLI Tool Usage

### Create New Bot
```bash
./botcli create [bot-name]
```
Interactive module selection with category organization.

### List Available Modules
```bash
./botcli list [--modules|--commands] [--filter ModuleName]
```

### Add Modules to Existing Bot
```bash
./botcli add <git-url|local-path>
```

### Build Bot Binary
```bash
./botcli build [--os linux] [--arch amd64] [--output bot-name]
```

## 🔧 Configuration

### Environment Variables (.env)
```env
DISCORD_TOKEN=your_discord_bot_token
DATABASE_URL=postgres://user:pass@localhost/dbname?sslmode=disable
BOT_OWNER_ID=your_discord_user_id
```

### Bot Configuration (bot.config.json)
```json
{
  "name": "my-bot",
  "description": "My custom Discord bot",
  "modules": ["Economy", "F1", "General"],
  "commands": {
    "balance": true,
    "f1": true,
    "work": false
  }
}
```

## 🏗️ Development

### Creating a New Module

1. **Create module directory:**
   ```bash
   mkdir commands/mymodule
   ```

2. **Create module.go:**
   ```go
   package mymodule
   
   import "DiscordBot/commands"
   
   func init() {
       module := &commands.ModuleInfo{
           Name:        "MyModule",
           Description: "What this module does",
           Category:    "Custom",
           Commands: []commands.CommandInfo{
               {
                   Name:        "mycommand",
                   Description: "Command description",
                   Usage:       ".mycommand <args>",
                   Category:    "Custom",
               },
           },
       }
       
       commands.RegisterModule(module)
       commands.RegisterCommand("mycommand", MyCommandHandler)
   }
   ```

3. **Implement command handler:**
   ```go
   func MyCommandHandler(b *bot.Bot, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
       // Your command logic here
   }
   ```

### Module Guidelines
- Keep modules focused and self-contained
- Use clear, descriptive command names
- Include proper error handling
- Follow Go naming conventions
- Add comprehensive documentation

## 🤝 How to Contribute

We welcome contributions! Here's how to get started:

### 1. Fork the Repository
```bash
git fork https://github.com/yourusername/modubot-go
cd modubot-go
```

### 2. Create Your Module
- Create your module in the `commands/` directory
- Follow the module development guidelines
- Test your module thoroughly

### 3. Submit a Pull Request
Include in your PR:
- **Module Description**: What does your module do?
- **Commands List**: All commands with descriptions
- **Dependencies**: Any external dependencies
- **Configuration**: Required configuration options
- **Testing**: How you tested the module

### PR Template
```markdown
## Module: [Module Name]

**Category**: [Category Name]
**Description**: [Brief description of what the module does]

### Commands Added:
- `.command1` - Description
- `.command2` - Description

### Dependencies:
- [List any dependencies]

### Configuration:
- [Any required configuration]

### Testing:
- [How you tested the module]
```

### 4. Future: Optional Modules Folder
**TODO**: We're planning to move optional/community modules to a separate `optional-modules/` folder to keep the core lean while allowing extensive customization.

## 📚 Documentation

### 🚀 Getting Started
- [CLI Usage Guide](documentation/CLI_USAGE.md) - Comprehensive CLI tool documentation
- [Database Schema](documentation/DBSchema.md) - Database structure and setup

### 🏗️ Development
- [Module Development Guide](documentation/MODULE_DEVELOPMENT.md) - Complete guide for creating modules
- [CLI Tool Architecture](documentation/CLI_TOOL.md) - CLI tool design and implementation

### 📋 Project Management
- [TODO & Roadmap](documentation/TODO.md) - Current tasks and future plans
- [Permission System](documentation/PERMISSION_REFACTOR.md) - Permission system documentation
- [Database Schema Updates](documentation/newSchema.md) - Schema migration notes
- [EPL/FPL Integration](documentation/EPL_FPL_INTEGRATION.md) - Sports module integration guide

## 🏗️ Architecture

```
modubot-go/
├── cli/                    # CLI tool for bot creation
├── bot/                    # Core bot functionality
├── commands/               # All modules
│   ├── module.go          # Module registry
│   ├── economy/           # Economy module
│   ├── sports/            # Sports modules (F1, EPL, FPL)
│   ├── egypt/             # Egypt-specific modules
│   ├── finance/           # Finance module
│   ├── general/           # General utilities
│   ├── help/              # Help system
│   ├── admin/             # Administration
│   ├── moderation/        # Moderation tools
│   └── roles/             # Role management
├── utils/                 # Shared utilities
└── webhooks/              # Webhook handlers
```

## 🔒 Security

- All commands require appropriate Discord permissions
- Database queries use parameterized statements
- Input validation on all user inputs
- Rate limiting on expensive operations
- Secure token handling

## 📊 Performance

- Efficient command lookup using hash maps
- Minimal memory footprint per module
- Lazy loading of module resources
- Connection pooling for database operations
- Optimized embed generation

## 🐛 Troubleshooting

### Common Issues

**Bot not responding to commands:**
- Check Discord permissions
- Verify bot token is correct
- Ensure database connection is working

**Module not loading:**
- Check module.go syntax
- Verify import paths
- Check for naming conflicts

**CLI tool errors:**
- Run `go mod tidy` in CLI directory
- Check Go version (1.21+ required)
- Verify file permissions

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [DiscordGo](https://github.com/bwmarrin/discordgo) - Discord API wrapper
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- All contributors who help make this project better

## 📞 Support

- Create an [Issue](https://github.com/yourusername/modubot-go/issues) for bug reports
- Join our [Discord Server](https://discord.gg/your-server) for community support
- Check the [Wiki](https://github.com/yourusername/modubot-go/wiki) for detailed guides

---

**Made with ❤️ by the ModuBot-Go community**

*Build your perfect Discord bot, one module at a time!* 🚀