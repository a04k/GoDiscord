package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
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

	// Check if migration is needed for users table
	var guildIDExists bool
	err = tx.QueryRow(`
		SELECT EXISTS(
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_name = 'users' AND column_name = 'guild_id'
		)
	`).Scan(&guildIDExists)
	if err != nil {
		log.Fatal("Failed to check if guild_id column exists:", err)
	}

	// Check if migration is needed for disabled_commands table
	var nameColumnExists bool
	err = tx.QueryRow(`
		SELECT EXISTS(
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_name = 'disabled_commands' AND column_name = 'name'
		)
	`).Scan(&nameColumnExists)
	if err != nil {
		log.Fatal("Failed to check if name column exists in disabled_commands:", err)
	}

	if guildIDExists && nameColumnExists {
		fmt.Println("Database appears to be already migrated. Exiting.")
		return
	}

	// Perform migration
	fmt.Println("Starting database migration...")

	// Migrate users table if needed
	if !guildIDExists {
		fmt.Println("Migrating users table...")
		// 1. Create a backup of the users table (in case we need to rollback)
		fmt.Println("1. Creating backup of users table...")
		_, err = tx.Exec(`
			CREATE TABLE users_backup AS 
			SELECT * FROM users
		`)
		if err != nil {
			log.Fatal("Failed to create backup of users table:", err)
		}

		// 2. Add guild_id column to users table
		fmt.Println("2. Adding guild_id column to users table...")
		_, err = tx.Exec(`
			ALTER TABLE users 
			ADD COLUMN guild_id TEXT DEFAULT '' NOT NULL
		`)
		if err != nil {
			log.Fatal("Failed to add guild_id column:", err)
		}

		// 3. Update primary key to include guild_id
		fmt.Println("3. Updating primary key to include guild_id...")
		_, err = tx.Exec(`
			ALTER TABLE users 
			DROP CONSTRAINT IF EXISTS users_pkey
		`)
		if err != nil {
			log.Fatal("Failed to drop existing primary key:", err)
		}

		_, err = tx.Exec(`
			ALTER TABLE users 
			ADD PRIMARY KEY (guild_id, user_id)
		`)
		if err != nil {
			log.Fatal("Failed to add new primary key:", err)
		}

		// 4. For existing users, set a default guild_id (this is a limitation of the migration)
		fmt.Println("4. Setting default guild_id for existing users...")
		_, err = tx.Exec(`
			UPDATE users 
			SET guild_id = 'DEFAULT_GUILD' 
			WHERE guild_id = ''
		`)
		if err != nil {
			log.Fatal("Failed to set default guild_id:", err)
		}

		// 5. Make guild_id NOT NULL (it already is, but just to be explicit)
		fmt.Println("5. Ensuring guild_id is NOT NULL...")
		_, err = tx.Exec(`
			ALTER TABLE users 
			ALTER COLUMN guild_id SET NOT NULL
		`)
		if err != nil {
			log.Fatal("Failed to set guild_id as NOT NULL:", err)
		}
	}

	// Migrate disabled_commands table if needed
	if !nameColumnExists {
		fmt.Println("Migrating disabled_commands table...")
		
		// Check if the old table structure exists
		var commandNameColumnExists bool
		err = tx.QueryRow(`
			SELECT EXISTS(
				SELECT 1 
				FROM information_schema.columns 
				WHERE table_name = 'disabled_commands' AND column_name = 'command_name'
			)
		`).Scan(&commandNameColumnExists)
		if err != nil {
			log.Fatal("Failed to check if command_name column exists in disabled_commands:", err)
		}
		
		if commandNameColumnExists {
			// Rename command_name column to name
			fmt.Println("Renaming command_name column to name...")
			_, err = tx.Exec(`
				ALTER TABLE disabled_commands 
				RENAME COLUMN command_name TO name
			`)
			if err != nil {
				log.Fatal("Failed to rename command_name column:", err)
			}
			
			// Add type column if it doesn't exist
			var typeColumnExists bool
			err = tx.QueryRow(`
				SELECT EXISTS(
					SELECT 1 
					FROM information_schema.columns 
					WHERE table_name = 'disabled_commands' AND column_name = 'type'
				)
			`).Scan(&typeColumnExists)
			if err != nil {
				log.Fatal("Failed to check if type column exists in disabled_commands:", err)
			}
			
			if !typeColumnExists {
				fmt.Println("Adding type column...")
				_, err = tx.Exec(`
					ALTER TABLE disabled_commands 
					ADD COLUMN type TEXT DEFAULT 'command' NOT NULL
				`)
				if err != nil {
					log.Fatal("Failed to add type column:", err)
				}
			}
		} else {
			// Recreate the table with the correct schema
			fmt.Println("Recreating disabled_commands table with correct schema...")
			_, err = tx.Exec(`
				DROP TABLE disabled_commands
			`)
			if err != nil {
				log.Fatal("Failed to drop disabled_commands table:", err)
			}
			
			_, err = tx.Exec(`
				CREATE TABLE disabled_commands (
					guild_id TEXT NOT NULL,
					name TEXT NOT NULL,
					type TEXT NOT NULL,
					PRIMARY KEY (guild_id, name)
				)
			`)
			if err != nil {
				log.Fatal("Failed to create disabled_commands table:", err)
			}
		}
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		log.Fatal("Failed to commit transaction:", err)
	}

	fmt.Println("Database migration completed successfully!")
	fmt.Println("Note: Existing users have been assigned to 'DEFAULT_GUILD'.")
	fmt.Println("You may want to update this based on your specific needs.")
}