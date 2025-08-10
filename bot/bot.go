package bot

import (
	"database/sql"
	"log"

	"github.com/bwmarrin/discordgo"
	_ "github.com/lib/pq"
)

type Bot struct {
	Db     *sql.DB
	Client *discordgo.Session
}

type ExchangeResponse struct {
	Result float64 `json:"result"`
}

type BTCResponse struct {
	Bitcoin struct {
		USD float64 `json:"usd"`
	} `json:"bitcoin"`
}

func NewBot(token string, dbURL string) (*Bot, error) {
	client, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}
	client.Identify.Intents = discordgo.IntentsAll

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	bot := &Bot{Db: db, Client: client}
	client.AddHandler(bot.guildCreate)

	return bot, nil
}

func (b *Bot) GetUserRole(guildID, userID string) (string, error) {
	var role string
	err := b.Db.QueryRow("SELECT role FROM permissions WHERE guild_id = $1 AND user_id = $2", guildID, userID).Scan(&role)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil // No specific role, but not an error
		}
		return "", err
	}
	return role, nil
}

func (b *Bot) IsOwner(guildID, userID string) (bool, error) {
	role, err := b.GetUserRole(guildID, userID)
	if err != nil {
		return false, err
	}
	return role == "owner", nil
}

func (b *Bot) IsAdmin(guildID, userID string) (bool, error) {
	role, err := b.GetUserRole(guildID, userID)
	if err != nil {
		return false, err
	}
	return role == "admin" || role == "owner", nil
}

func (b *Bot) IsMod(guildID, userID string) (bool, error) {
	role, err := b.GetUserRole(guildID, userID)
	if err != nil {
		return false, err
	}
	return role == "moderator" || role == "admin" || role == "owner", nil
}

func (b *Bot) guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	// Check if the owner already has a role
	var exists bool
	err := b.Db.QueryRow("SELECT EXISTS(SELECT 1 FROM permissions WHERE guild_id = $1 AND user_id = $2)", event.Guild.ID, event.Guild.OwnerID).Scan(&exists)
	if err != nil {
		log.Printf("Failed to check if guild owner exists: %v", err)
		return
	}

	if !exists {
		// Set the guild owner as the owner in the permissions table
		_, err := b.Db.Exec("INSERT INTO permissions (guild_id, user_id, role) VALUES ($1, $2, $3)", event.Guild.ID, event.Guild.OwnerID, "owner")
		if err != nil {
			log.Printf("Failed to set guild owner: %v", err)
		}
	}
}