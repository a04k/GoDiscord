# GOBOT - a multi-utility discord bot built using Go

GOBOT is a discord bot built fully using Go (As a learning project) that has some nice time saving features and integrations with APIs such as WE Egypt
to grab your quota details (more providers to be added soon), Currency exchange rate scraper (Currently only USD to EGP), Bitcoin prices and some basic
economy features.

---

## Features

### **1. Economy System**
- **Balance**: `.bal`
- **Daily Reward**: `.daily`
- **Work**: `.work` (6-hour cooldown)
- **Transfer**: `.transfer <@user> <amount>`
- **Flip**: `.flip <amount|all>`
- **Admin Commands**:
  - Add coins: `.add <@user> <amount>`
  - Promote: `.promote <@user>`

### **2. Utility Commands**
- **USD to EGP**: `.usd [amount]`
- **BTC Price**: `.btc`
- **Internet Quota**: `/quota` (requires `/setup <landline> <password>`)

### **Note: if you're only here for the WE Api, run the test.go folder in the QCheckWE directory.**

---
## Setup (Running release executables)
### **1. Prerequisites**
- Gobot executable for your OS : [Link](https://github.com/a04k/GoDiscord/releases)
### **2. Setup**
- Create directory for gobot
- Place the gobot executable
- Create a .env file and replace the values with the key data:
  
```env
  DISCORD_TOKEN=your_token
  DATABASE_URL="your_db_URL"
```


## Setup (For Building)

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
    last_daily TIMESTAMP,
    is_admin BOOLEAN DEFAULT FALSE
);
```

### **3. Run the Bot**
```bash
go mod tidy
go run main.go
```

---

## User Commands
| Command                              | Description                                                        |
|--------------------------------------|--------------------------------------------------------------------|
| `.bal / .balance`                    | Check your balance                                                 |
| `.work`                              | Earn coins (6h cooldown)                                           |
| `.flip <amount / all>`               | Gamble coins: (specified amount / all)                             |
| `.transfer <@user> <amount>`         | Send coins to another user                                         |
| `.usd [amount]`                      | USD to EGP exchange rate                                           |
| `.btc`                               | Bitcoin price in USD                                               |
| `/setup <landline> <password>`| Save WE credentials                                                       |
| `/quota`                             | Check internet quota (after setup)                                 |

## Mod/Admin Commands
| Command                              | Description                                                        |
|--------------------------------------|--------------------------------------------------------------------|
| `.add <@user> <amount>`              | Admin: Add coins to a user                                         |
| `.sa <@user>`                        | Admin: Promote a user to admin                                     |
| `.createrole <role name> [color] [permissions] [hoist]` | Create a new role with options for color, permissions, and hoisting |
| `.setrole/sr <@user> <role name>`     | Gives the user the role                                       |
| `.inrole <role name or mention>`     | View all users in a specific role                                       |

---
