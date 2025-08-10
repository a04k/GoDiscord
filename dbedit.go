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

	// Check if permissions table exists
	var permissionsTableExists bool
	err = tx.QueryRow(`
		SELECT EXISTS(
			SELECT 1
			FROM information_schema.tables
			WHERE table_name = 'permissions'
		)
	`).Scan(&permissionsTableExists)
	if err != nil {
		log.Fatal("Failed to check if permissions table exists:", err)
	}

	if !permissionsTableExists {
		fmt.Println("Creating permissions table...")
		_, err = tx.Exec(`
			CREATE TABLE permissions (
				guild_id TEXT NOT NULL,
				user_id TEXT NOT NULL,
				role TEXT NOT NULL,
				PRIMARY KEY (guild_id, user_id)
			)
		`)
		if err != nil {
			log.Fatal("Failed to create permissions table:", err)
		}
		fmt.Println("permissions table created successfully.")
	} else {
		fmt.Println("permissions table already exists.")
	}

	// Check if is_owner column exists in users table
	var isOwnerExists bool
	err = tx.QueryRow(`
		SELECT EXISTS(
			SELECT 1
			FROM information_schema.columns
			WHERE table_name = 'users' AND column_name = 'is_owner'
		)
	`).Scan(&isOwnerExists)
	if err != nil {
		log.Fatal("Failed to check if is_owner column exists:", err)
	}

	if isOwnerExists {
		fmt.Println("Dropping is_owner column from users table...")
		_, err = tx.Exec(`
			ALTER TABLE users
			DROP COLUMN is_owner
		`)
		if err != nil {
			log.Fatal("Failed to drop is_owner column:", err)
		}
		fmt.Println("is_owner column dropped successfully.")
	} else {
		fmt.Println("is_owner column does not exist in users table.")
	}
	
	// Check if is_admin column exists in users table
	var isAdminExists bool
	err = tx.QueryRow(`
		SELECT EXISTS(
			SELECT 1
			FROM information_schema.columns
			WHERE table_name = 'users' AND column_name = 'is_admin'
		)
	`).Scan(&isAdminExists)
	if err != nil {
		log.Fatal("Failed to check if is_admin column exists:", err)
	}

	if isAdminExists {
		fmt.Println("Dropping is_admin column from users table...")
		_, err = tx.Exec(`
			ALTER TABLE users
			DROP COLUMN is_admin
		`)
		if err != nil {
			log.Fatal("Failed to drop is_admin column:", err)
		}
		fmt.Println("is_admin column dropped successfully.")
	} else {
		fmt.Println("is_admin column does not exist in users table.")
	}

	// Check if is_mod column exists in users table
	var isModExists bool
	err = tx.QueryRow(`
		SELECT EXISTS(
			SELECT 1
			FROM information_schema.columns
			WHERE table_name = 'users' AND column_name = 'is_mod'
		)
	`).Scan(&isModExists)
	if err != nil {
		log.Fatal("Failed to check if is_mod column exists:", err)
	}

	if isModExists {
		fmt.Println("Dropping is_mod column from users table...")
		_, err = tx.Exec(`
			ALTER TABLE users
			DROP COLUMN is_mod
		`)
		if err != nil {
			log.Fatal("Failed to drop is_mod column:", err)
		}
		fmt.Println("is_mod column dropped successfully.")
	} else {
		fmt.Println("is_mod column does not exist in users table.")
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		log.Fatal("Failed to commit transaction:", err)
	}

	fmt.Println("Database migration completed successfully!")
}
