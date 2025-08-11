package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: No .env file found")
	}

	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test connection
	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	fmt.Println("Connected to database successfully")

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		log.Fatal("Failed to begin transaction:", err)
	}
	defer tx.Rollback()

	// Drop existing tables in correct order to avoid foreign key constraint issues
	fmt.Println("Dropping existing tables...")
	tables := []string{
		"reminders", "scheduled_messages", "transactions", "role_daily_modifiers",
		"guild_members", "disabled_commands", "permissions", "guilds", "users",
	}

	for _, table := range tables {
		_, err = tx.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table))
		if err != nil {
			log.Printf("Warning: Failed to drop table %s: %v", table, err)
		}
	}
	fmt.Println("Existing tables dropped.")

	// Migrate to new schema based on newSchema.md
	fmt.Println("Creating new schema...")

	// Create global users table
	fmt.Println("Creating global users table...")
	_, err = tx.Exec(`
		CREATE TABLE users (
			user_id BIGINT PRIMARY KEY,
			username TEXT,
			avatar TEXT,
			created_at TIMESTAMPTZ DEFAULT now(),
			updated_at TIMESTAMPTZ DEFAULT now()
		)
	`)
	if err != nil {
		log.Fatal("Failed to create global users table:", err)
	}
	fmt.Println("Global users table created successfully.")

	// Create guilds table
	fmt.Println("Creating guilds table...")
	_, err = tx.Exec(`
		CREATE TABLE guilds (
			guild_id BIGINT PRIMARY KEY,
			owner_id BIGINT NOT NULL,
			name TEXT,
			currency_name TEXT DEFAULT 'Coins',
			settings JSONB DEFAULT '{}'::jsonb,
			created_at TIMESTAMPTZ DEFAULT now(),
			updated_at TIMESTAMPTZ DEFAULT now()
		)
	`)
	if err != nil {
		log.Fatal("Failed to create guilds table:", err)
	}
	fmt.Println("guilds table created successfully.")

	// Create guild_members table
	fmt.Println("Creating guild_members table...")
	_, err = tx.Exec(`
		CREATE TABLE guild_members (
			guild_id BIGINT NOT NULL REFERENCES guilds(guild_id) ON DELETE CASCADE,
			user_id BIGINT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			balance BIGINT DEFAULT 0,
			last_daily TIMESTAMPTZ,
			base_daily_hours INT DEFAULT 24,
			joined_at TIMESTAMPTZ DEFAULT now(),
			PRIMARY KEY (guild_id, user_id)
		)
	`)
	if err != nil {
		log.Fatal("Failed to create guild_members table:", err)
	}
	fmt.Println("guild_members table created successfully.")

	// Create role_daily_modifiers table
	fmt.Println("Creating role_daily_modifiers table...")
	_, err = tx.Exec(`
		CREATE TABLE role_daily_modifiers (
			guild_id BIGINT NOT NULL REFERENCES guilds(guild_id) ON DELETE CASCADE,
			role_id BIGINT NOT NULL,
			min_hours INT NOT NULL,
			PRIMARY KEY (guild_id, role_id)
		)
	`)
	if err != nil {
		log.Fatal("Failed to create role_daily_modifiers table:", err)
	}
	fmt.Println("role_daily_modifiers table created successfully.")

	// Create disabled_commands table
	fmt.Println("Creating disabled_commands table...")
	_, err = tx.Exec(`
		CREATE TABLE disabled_commands (
			guild_id BIGINT NOT NULL REFERENCES guilds(guild_id) ON DELETE CASCADE,
			name TEXT NOT NULL,
			type TEXT NOT NULL CHECK (type IN ('command', 'category')),
			PRIMARY KEY (guild_id, name)
		)
	`)
	if err != nil {
		log.Fatal("Failed to create disabled_commands table:", err)
	}
	fmt.Println("disabled_commands table created successfully.")

	// Create unified reminders table
	fmt.Println("Creating reminders table...")
	_, err = tx.Exec(`
		CREATE TABLE reminders (
			reminder_id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			guild_id BIGINT REFERENCES guilds(guild_id) ON DELETE CASCADE,
			message TEXT NOT NULL,
			remind_at TIMESTAMPTZ NOT NULL,
			source TEXT NOT NULL,
			data JSONB,
			sent BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMPTZ DEFAULT now()
		)
	`)
	if err != nil {
		log.Fatal("Failed to create reminders table:", err)
	}
	fmt.Println("reminders table created successfully.")

	// Create scheduled_messages table
	fmt.Println("Creating scheduled_messages table...")
	_, err = tx.Exec(`
		CREATE TABLE scheduled_messages (
			message_id BIGSERIAL PRIMARY KEY,
			guild_id BIGINT NOT NULL REFERENCES guilds(guild_id) ON DELETE CASCADE,
			channel_id BIGINT NOT NULL,
			content TEXT NOT NULL,
			send_at TIMESTAMPTZ NOT NULL,
			sent BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMPTZ DEFAULT now()
		)
	`)
	if err != nil {
		log.Fatal("Failed to create scheduled_messages table:", err)
	}
	fmt.Println("scheduled_messages table created successfully.")

	// Create transactions table
	fmt.Println("Creating transactions table...")
	_, err = tx.Exec(`
		CREATE TABLE transactions (
			transaction_id BIGSERIAL PRIMARY KEY,
			guild_id BIGINT NOT NULL REFERENCES guilds(guild_id) ON DELETE CASCADE,
			user_id BIGINT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			amount BIGINT NOT NULL,
			balance_before BIGINT,
			balance_after BIGINT,
			reason TEXT,
			meta JSONB,
			created_at TIMESTAMPTZ DEFAULT now()
		)
	`)
	if err != nil {
		log.Fatal("Failed to create transactions table:", err)
	}
	fmt.Println("transactions table created successfully.")

	// Create indexes
	fmt.Println("Creating indexes...")
	indexes := []string{
		"CREATE INDEX idx_gm_guild_balance_desc ON guild_members (guild_id, balance DESC)",
		"CREATE INDEX idx_gm_user ON guild_members (user_id)",
		"CREATE INDEX idx_tx_guild_time ON transactions (guild_id, created_at DESC)",
		"CREATE INDEX idx_reminders_due ON reminders (sent, remind_at)",
		"CREATE INDEX idx_scheduled_due ON scheduled_messages (sent, send_at)",
	}
	for _, indexSQL := range indexes {
		_, err = tx.Exec(indexSQL)
		if err != nil {
			log.Printf("Warning: Failed to create index with SQL '%s': %v", indexSQL, err)
		}
	}
	fmt.Println("Indexes created successfully.")

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		log.Fatal("Failed to commit transaction:", err)
	}

	fmt.Println("Database migration completed successfully!")
}