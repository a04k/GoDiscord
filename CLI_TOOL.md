# CLI Tool for Custom Bot Creation

## Overview
This document outlines the features and considerations for creating a CLI tool that allows users to create their own customized instance of the bot.

## Key Features to Implement

### 1. Bot Configuration
- **Bot Token**: Discord bot token for the new instance
- **Database Configuration**: Connection string for PostgreSQL database
- **Bot Owner ID**: User ID of the person who owns this bot instance
- **Default Prefix**: Command prefix (default: ".")
- **Enabled Categories**: Which command categories to enable by default

### 2. Guild-Specific Configuration
- **Auto-Owner Detection**: Automatically detect and set guild owners based on Discord's guild owner ID
- **Owner Override**: Allow bot owner to manually set guild owners in existing guilds
- **Permission Hierarchy**:
  - Bot Owner (from env) > Guild Owner (from Discord) > Admins (set by guild owner) > Mods (set by admins/guild owner)

### 3. Command Management
- **Enable/Disable Commands**: Allow enabling/disabling individual commands
- **Enable/Disable Categories**: Allow enabling/disabling entire command categories
- **Custom Command Aliases**: Allow setting custom aliases for commands
- **Command Permissions**: Set which roles/users can use specific commands

### 4. Economy System
- **Currency Name**: Custom name for the economy currency
- **Starting Balance**: Default balance for new users
- **Work Rewards**: Configure min/max rewards for work command
- **Transfer Limits**: Set limits on how much can be transferred

### 5. F1 Notifications
- **Enable/Disable**: Toggle F1 notifications
- **Notification Times**: Configure when to send notifications
- **Timezone**: Set timezone for notifications

## Handling Bot Owner ID Across Multiple Guilds

The bot owner ID in the environment variable refers to the person who **owns the bot instance**, not the owner of any specific guild. This is important to understand:

### How It Works:
1. **Bot Instance Owner**: The person who runs the bot (you, in this case) is the "bot owner"
2. **Guild Owners**: Each Discord server has its own owner, automatically detected by the bot
3. **Permission Hierarchy**: 
   - Bot Owner (from env) has the highest privileges across ALL guilds
   - Guild Owners (from Discord) have high privileges in their specific guild
   - Admins and Mods have guild-specific privileges set by the guild owner

### Example Scenario:
- You create a bot instance with BOT_OWNER_ID=174176306865504257
- The bot is added to Guild A (owned by User 123) and Guild B (owned by User 456)
- You (as bot owner) can use `.setupowner` to manually set owners if needed
- User 123 is automatically recognized as owner of Guild A
- User 456 is automatically recognized as owner of Guild B
- In Guild A, User 123 can set admins and mods
- In Guild B, User 456 can set admins and mods
- You (bot owner) can do anything in both guilds

### Security Considerations:
1. The BOT_OWNER_ID should be kept secret and only known to the person running the bot instance
2. Bot owner commands should be clearly marked and limited in the help system
3. Consider adding a confirmation step for destructive bot owner actions
4. Log all bot owner actions for security auditing

## Command Category and Individual Command Management

### Category Management
Users should be able to enable or disable entire command categories:
- **General**: Basic commands like help, command list, currency conversion
- **Economy**: All economy-related commands (balance, work, transfer, flip)
- **F1**: All F1-related commands and notifications
- **Moderation**: Server moderation commands (kick, ban, mute, etc.)
- **Admin**: Administrative commands (set admin, disable/enable commands)
- **Roles**: Role management commands

### Individual Command Management
Within each category, users should be able to:
- Disable specific commands they don't want
- Enable specific commands they do want (if the category is disabled)
- Set custom aliases for commands
- Set permissions for individual commands

### Example CLI Interface:
```bash
# List all command categories
bot-cli commands categories

# Disable an entire category
bot-cli commands category-disable Economy

# Enable a specific command within a disabled category
bot-cli commands enable flip

# Disable a specific command within an enabled category
bot-cli commands disable f1wdc

# List all commands with their status
bot-cli commands list --all

# Set a custom alias for a command
bot-cli commands alias flip coinflip
```

### Database Schema for Command Management

```sql
CREATE TABLE command_config (
    guild_id TEXT NOT NULL,
    command_name TEXT NOT NULL,
    category TEXT NOT NULL,
    enabled BOOLEAN DEFAULT TRUE,
    custom_alias TEXT,
    PRIMARY KEY (guild_id, command_name)
);

CREATE TABLE category_config (
    guild_id TEXT NOT NULL,
    category TEXT NOT NULL,
    enabled BOOLEAN DEFAULT TRUE,
    PRIMARY KEY (guild_id, category)
);
```

### Priority System
1. **Category Settings**: If a category is disabled, all commands in that category are disabled by default
2. **Individual Command Override**: Commands can be individually enabled even if their category is disabled
3. **Individual Command Disable**: Commands can be individually disabled even if their category is enabled

### Example Configuration:
- Category "Economy" is disabled
- Command "flip" is individually enabled
- Result: Only the "flip" command works, all other economy commands are disabled

## Database Schema Considerations

### Users Table
```sql
CREATE TABLE users (
    guild_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    balance INTEGER DEFAULT 0,
    last_daily TIMESTAMP,
    is_admin BOOLEAN DEFAULT FALSE,
    is_mod BOOLEAN DEFAULT FALSE,
    is_owner BOOLEAN DEFAULT FALSE,  -- Guild owner (from Discord)
    PRIMARY KEY (guild_id, user_id)
);
```

### Bot Configuration Table
```sql
CREATE TABLE bot_config (
    bot_owner_id TEXT NOT NULL,  -- Bot instance owner (from env)
    guild_id TEXT NOT NULL,
    prefix TEXT DEFAULT '.',
    economy_enabled BOOLEAN DEFAULT TRUE,
    f1_notifications_enabled BOOLEAN DEFAULT TRUE,
    PRIMARY KEY (guild_id)
);
```

### Command Configuration Tables
```sql
CREATE TABLE command_config (
    guild_id TEXT NOT NULL,
    command_name TEXT NOT NULL,
    category TEXT NOT NULL,
    enabled BOOLEAN DEFAULT TRUE,
    custom_alias TEXT,
    PRIMARY KEY (guild_id, command_name)
);

CREATE TABLE category_config (
    guild_id TEXT NOT NULL,
    category TEXT NOT NULL,
    enabled BOOLEAN DEFAULT TRUE,
    PRIMARY KEY (guild_id, category)
);
```

## CLI Tool Commands

### Initialization
```bash
bot-cli init --token=DISCORD_TOKEN --db-url=DATABASE_URL --owner-id=USER_ID
```

### Guild Management
```bash
bot-cli guild list                    # List all guilds the bot is in
bot-cli guild info <guild_id>         # Show guild configuration
bot-cli guild setup-owner <guild_id> <user_id>  # Manually set guild owner
```

### Command Management
```bash
bot-cli commands categories           # List all command categories
bot-cli commands list                 # List all available commands
bot-cli commands enable <command>     # Enable a command
bot-cli commands disable <command>    # Disable a command
bot-cli commands category-enable <category>    # Enable a command category
bot-cli commands category-disable <category>   # Disable a command category
bot-cli commands alias <command> <alias>       # Set a custom alias for a command
```

### Economy Configuration
```bash
bot-cli economy config --currency-name=Coins --starting-balance=100
bot-cli economy work-config --min-reward=50 --max-reward=200
```

### F1 Configuration
```bash
bot-cli f1 enable                     # Enable F1 notifications
bot-cli f1 disable                    # Disable F1 notifications
bot-cli f1 timezone --set=UTC+3       # Set timezone for notifications
```

## Implementation Notes

1. **Environment Variables**: Store sensitive configuration in environment variables
2. **Configuration Files**: Use YAML/JSON files for complex configuration
3. **Database Migration**: Implement a migration system for schema changes
4. **Backup System**: Include backup/restore functionality for configurations
5. **Validation**: Validate all inputs and configurations before applying
6. **Logging**: Implement comprehensive logging for debugging and auditing
7. **Interactive Mode**: Provide an interactive mode for easier configuration
8. **Batch Operations**: Allow batch enabling/disabling of commands/categories