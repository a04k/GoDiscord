# Discord Bot - Feature-Rich Economy and Utility Bot

This Discord bot provides an economy system, utility commands, and admin tools. It uses PostgreSQL and integrates with external APIs.

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
| `.flip <amount|all>`        | Gamble coins                         |
| `.transfer <@user> <amount>`| Send coins to another user           |
| `.add <@user> <amount>`     | Admin: Add coins to a user           |
| `.promote <@user>`          | Admin: Promote a user to admin       |
| `.usd [amount]`             | USD to EGP exchange rate             |
| `.btc`                      | Bitcoin price in USD                 |
| `/quota`                    | Check internet quota (after setup)   |
| `/setup <landline> <password>`| Save WE credentials               |

---
