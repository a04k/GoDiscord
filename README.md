# GOBOT - a multi-utility discord bot built using Go

GOBOT is a discord bot built fully using Go (As a learning project) that has some nice time saving features and integrations with APIs such as WE Egypt
to grab your quota details (more providers to be added soon), Currency exchange rate scraper (Currently only USD to EGP), Bitcoin prices and some basic
economy features.

---

## **Features**

### **1. Economy System**
- **Balance**: `.bal` or `.balance` — Check your balance
- **Daily Reward**: `.daily` — Claim your daily reward (with cooldown)
- **Work**: `.work` — Earn coins (6-hour cooldown)
- **Transfer**: `.transfer <@user> <amount>` — Send coins to another user
- **Flip**: `.flip <amount|all>` — Gamble coins (specified amount or all)
- **Admin/Owner Commands**:
  - **Add coins**: `.add <@user> <amount>` — Add coins to a user 
  - **Create role**: `.cr/createrole <role name> [color] [permissions] [hoist]` — Create a new role with color, permissions, and hoisting options
  - **Assign role**: `.sr/setrole <@user> <role name>` or `.sr <@user> <role name>` — Assign a specific role to a user
  - **View users in role**: `.inrole <role name or mention>` — View all users in a specific role

### **2. Utility Commands**
- **USD to EGP**: `.usd [amount]` — Get USD to EGP exchange rate
- **BTC Price**: `.btc` — Get the current Bitcoin price in USD
- **Internet Quota**: `/quota` — Check internet quota (requires `/setup <landline> <password>`)

### **Note:**
If you're only here for the WE Api, run the `test.go` folder in the `QCheckWE` directory.


# Custom Bot Setup

## Setup (Running release executables)
### **1. Prerequisites**
- Gobot executable for your OS : [Link](https://github.com/a04k/GoDiscord/releases), Discord bot token.
### **2. Setup**
- Create directory for gobot
- Place the gobot executable in the directory
- Create a .env file in the same directory and replace the values with the key data:
  
```env
  DISCORD_TOKEN=your_token
  DATABASE_URL="your_db_URL"
```
- Run the executable


## Setup (From Source)

### **1. Prerequisites**
- Go, PostgreSQL, and a Discord bot token.
- `.env` file:
  
```env
  DISCORD_TOKEN=your_token
  DATABASE_URL="your_db_URL"
```

### **2. Database Schema**
```sql
CREATE TABLE IF NOT EXISTS users (
    user_id TEXT PRIMARY KEY,
    balance INTEGER DEFAULT 0,
    last_daily TIMESTAMP
);
```

### **3. Run the Bot**
```bash
go mod tidy
go run main.go
```

---

## User Commands
| Command                              | Description                        |
|--------------------------------------|------------------------------------|
| `.bal / .balance`                    | Check your balance                 |
| `.work`                              | Earn coins (6h cooldown)           |
| `.flip <amount / all>`               | Gamble coins                       |
| `.transfer <@user> <amount>`         | Send coins to another user         |
| `.usd [amount]`                      | USD to EGP exchange rate           |
| `.btc`                               | Bitcoin price in USD               |
| `/setup <landline> <password>`       | Save WE credentials                |
| `/quota`                             | Check internet quota               |

## Admin/Mod Commands
| Command                              | Description                                |
|--------------------------------------|--------------------------------------------|
| `.add <@user> <amount>`              | Add coins to a user                        |
| `.createrole/cr <role name> [...]`      | Create role with options                   |
| `.setrole/sr <@user> <role name>`    | Assign role to user                        |
| `.roleinfo/ri <role>`                | Show detailed role info and permissions    |
| `.inrole <role name or mention>`     | List users in a role                       |

---

## Permissions System

GOBOT uses a hybrid permissions system:
- **Bot Owner**: The server owner (from Discord) has full access to all commands
- **Admins**: Users with the "Administrator" permission can use admin commands
- **Moderators**: Users with the "Manage Messages" permission can use moderation commands
- **Custom Moderators**: Users can be assigned as custom moderators using the `.sm` command

The system automatically recognizes users with the appropriate Discord permissions, but also allows for custom moderator assignments through the database for more granular control.
