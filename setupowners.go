package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"DiscordBot/bot"

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
	db, err := bot.NewBot("", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Db.Close()

	fmt.Println("Guild Owner Setup Tool")
	fmt.Println("======================")
	fmt.Println("This tool helps you manually set guild owners for existing guilds.")
	fmt.Println("")

	// Get input from user
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Enter guild ID (or 'quit' to exit): ")
		guildID, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal("Error reading input:", err)
		}
		guildID = strings.TrimSpace(guildID)

		if guildID == "quit" {
			break
		}

		fmt.Print("Enter user ID to set as owner: ")
		userID, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal("Error reading input:", err)
		}
		userID = strings.TrimSpace(userID)

		// Set the user as owner in the specified guild
		_, err = db.Db.Exec("INSERT INTO users (guild_id, user_id, is_owner) VALUES ($1, $2, TRUE) ON CONFLICT (guild_id, user_id) DO UPDATE SET is_owner = TRUE", guildID, userID)
		if err != nil {
			log.Printf("Error setting owner: %v", err)
			fmt.Println("Failed to set owner. Please try again.")
			continue
		}

		fmt.Printf("Successfully set <@%s> as owner in guild %s\n\n", userID, guildID)
	}

	fmt.Println("Done!")
}