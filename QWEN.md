# ModuBot-Go Project Context

## Project Overview

ModuBot-Go is a fully modular Discord bot built in Go with a powerful CLI tool for creating custom bot instances. The bot follows a modular architecture where users can choose only the modules they need, similar to create-next-app.

### Key Features
- Fully Modular Architecture - Pick and choose only the features you need
- CLI Tool - Create custom bot instances with interactive module selection
- Easy Development - Add new modules without touching core files
- Self-Contained Modules - Each module is independent and self-registering
- High Performance - Efficient command handling and memory management

### Technologies Used
- Go 1.21+
- DiscordGo library for Discord API integration
- PostgreSQL for database storage
- Cobra for CLI tool implementation

## Project Structure

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

## Core Components

### Module System
The bot uses a modular system where each module is self-contained and registers itself. Modules can include both regular commands and slash commands. The system supports:
- Command registration with aliases
- Slash command registration
- Module metadata (name, description, category, version, etc.)
- Dependencies between modules
- Configuration options per module

### Bot Core
The core bot functionality is in the `bot/` package and includes:
- Bot initialization with Discord token and database connection
- Permission checking (owner, admin, moderator)
- Guild creation handling
- Rate limiting for commands

### CLI Tool
The CLI tool allows users to create custom bot instances with only the modules they need. Features include:
- Interactive module selection organized by category
- Command-level enable/disable options
- Automatic generation of bot configuration files
- Copying of selected module files to new bot instance

## Module Categories

1. **Sports** - F1, EPL, FPL modules with race results, standings, and notifications
2. **Economy** - Virtual currency system with work, transfer, and gambling commands
3. **Egypt** - Region-specific modules like USD to EGP conversion and WE telecom quota checking
4. **Finance** - Financial market data like Bitcoin prices
5. **System** - General utilities, help system, administration, moderation, and role management

## Building and Running

### Prerequisites
- Go 1.21 or higher
- PostgreSQL database
- Discord Bot Token

### Using the CLI Tool (Recommended)
1. Build the CLI tool:
   ```bash
   cd cli
   go mod tidy
   go build -o botcli
   ```

2. Create your custom bot:
   ```bash
   ./botcli create my-awesome-bot
   ```

3. Configure and run:
   ```bash
   cd my-awesome-bot
   cp .env.example .env
   # Edit .env with your Discord token and database URL
   go mod tidy
   go build -o bot
   ./bot
   ```

### Manual Setup
1. Clone and build:
   ```bash
   go mod tidy
   go build -o bot
   ```

2. Configure environment:
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. Run the bot:
   ```bash
   ./bot
   ```

## Development Process

### Creating a New Module
1. Create module directory in `commands/`
2. Create `module.go` with module registration and command definitions
3. Implement command handlers
4. Register commands with the module system

### Module Guidelines
- Keep modules focused and self-contained
- Use clear, descriptive command names
- Include proper error handling
- Follow Go naming conventions
- Add comprehensive documentation

## Configuration

### Environment Variables (.env)
- DISCORD_TOKEN - Your Discord bot token
- DATABASE_URL - PostgreSQL database connection string
- BOT_OWNER_ID - Your Discord user ID for admin commands

### Bot Configuration (bot.config.json)
- name - Bot instance name
- description - Bot description
- modules - List of enabled modules
- commands - Command-level enable/disable settings
- config - Module-specific configuration options

## Contributing

We welcome contributions! To contribute:
1. Fork the repository
2. Create your module in the `commands/` directory
3. Follow the module development guidelines
4. Test your module thoroughly
5. Submit a pull request with module description, commands list, dependencies, and configuration requirements