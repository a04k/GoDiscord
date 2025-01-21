# GOBOT - a simple multi-utility discord bot built using Go

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

## Setup

### **1. Prerequisites**
- Go, PostgreSQL, and a Discord bot token.
- `.env` file:
  ```env
  DISCORD_TOKEN=your_token
  DATABASE_URL=postgresql://user:password@host:port/database
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

## Commands
| Command                     | Description                          |
|-----------------------------|--------------------------------------|
| `.bal`                      | Check your balance                   |
| `.daily`                    | Claim daily coins                    |
| `.work`                     | Earn coins (6h cooldown)             |
| `.flip <amount / all>`      | Gamble coins: (specified amount / all)|
| `.transfer <@user> <amount>`| Send coins to another user           |
| `.add <@user> <amount>`     | Admin: Add coins to a user           |
| `.promote <@user>`          | Admin: Promote a user to admin       |
| `.usd [amount]`             | USD to EGP exchange rate             |
| `.btc`                      | Bitcoin price in USD                 |
| `/quota`                    | Check internet quota (after setup)   |
| `/setup <landline> <password>`| Save WE credentials               |

---
