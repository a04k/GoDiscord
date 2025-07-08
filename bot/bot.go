package bot

import (
	"database/sql"

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

const schema = `
CREATE TABLE IF NOT EXISTS users (
    user_id TEXT PRIMARY KEY,
    balance INTEGER DEFAULT 0,
    last_daily TIMESTAMP,
    is_admin BOOLEAN DEFAULT FALSE,
    is_mod BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS credentials (
    user_id TEXT PRIMARY KEY,
    landline TEXT NOT NULL,
    password TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS f1_subscriptions (
    user_id TEXT PRIMARY KEY
);`

func NewBot(token string, dbURL string) (*Bot, error) {
	client, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}
	client.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(schema); err != nil {
		return nil, err
	}

	return &Bot{Db: db, Client: client}, nil
}
