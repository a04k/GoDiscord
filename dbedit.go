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

	// Add the FPL league ID column if it doesn't exist
	query := `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1
				FROM information_schema.columns
				WHERE table_name='guilds'
				AND column_name='fpl_league_id'
			) THEN
				ALTER TABLE guilds ADD COLUMN fpl_league_id BIGINT;
			END IF;
		END$$;
	`

	_, err = db.Exec(query)
	if err != nil {
		log.Fatal("Failed to add fpl_league_id column:", err)
	}

	fmt.Println("Schema update complete: fpl_league_id column ensured in guilds table.")
}
