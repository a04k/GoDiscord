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

func (b *Bot) IsOwner(guildID, userID string) (bool, error) {
	// Get the guild to check the actual owner ID
	guild, err := b.Client.State.Guild(guildID)
	if err != nil {
		// If not in state, fetch from Discord
		guild, err = b.Client.Guild(guildID)
		if err != nil {
			return false, err
		}
	}
	
	return guild.OwnerID == userID, nil
}

func (b *Bot) IsAdmin(guildID, userID string) (bool, error) {
	// Check if user is the guild owner (highest permission)
	isOwner, err := b.IsOwner(guildID, userID)
	if err != nil || isOwner {
		return isOwner, err
	}
	
	// Check if user has Administrator permission
	member, err := b.Client.State.Member(guildID, userID)
	if err != nil {
		// If not in state, fetch from Discord
		member, err = b.Client.GuildMember(guildID, userID)
		if err != nil {
			return false, err
		}
	}
	
	// Get user's permissions in the guild
	perms, err := b.Client.State.UserChannelPermissions(userID, guildID)
	if err != nil {
		// Calculate permissions manually if not in state
		perms = b.calculatePermissions(guildID, member)
	}
	
	// Check if user has Administrator permission (0x8 is the permission bit for Administrator)
	return perms&0x8 != 0, nil
}

func (b *Bot) IsMod(guildID, userID string) (bool, error) {
	// Check if user is owner or admin first
	isAdmin, err := b.IsAdmin(guildID, userID)
	if err != nil || isAdmin {
		return isAdmin, err
	}
	
	// Check if user has Manage Messages permission (0x2000 is the permission bit for Manage Messages)
	member, err := b.Client.State.Member(guildID, userID)
	if err != nil {
		// If not in state, fetch from Discord
		member, err = b.Client.GuildMember(guildID, userID)
		if err != nil {
			return false, err
		}
	}
	
	// Get user's permissions in the guild
	perms, err := b.Client.State.UserChannelPermissions(userID, guildID)
	if err != nil {
		// Calculate permissions manually if not in state
		perms = b.calculatePermissions(guildID, member)
	}
	
	// Check if user has Manage Messages permission
	hasManageMessages := perms&0x2000 != 0
	if hasManageMessages {
		return true, nil
	}
	
	// Check if user is marked as a moderator in our database
	var count int
	err = b.Db.QueryRow("SELECT COUNT(*) FROM permissions WHERE guild_id = $1 AND user_id = $2 AND role = 'moderator'", guildID, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}

// calculatePermissions manually calculates a member's permissions in a guild
// This is a simplified version and might not cover all edge cases
func (b *Bot) calculatePermissions(guildID string, member *discordgo.Member) int64 {
	// Get the guild
	guild, err := b.Client.State.Guild(guildID)
	if err != nil {
		// If not in state, fetch from Discord
		guild, err = b.Client.Guild(guildID)
		if err != nil {
			return 0
		}
	}
	
	var permissions int64
	
	// Start with @everyone role permissions
	for _, role := range guild.Roles {
		if role.ID == guild.ID { // @everyone role has the same ID as the guild
			permissions |= role.Permissions
			break
		}
	}
	
	// Add permissions from member's roles
	for _, roleID := range member.Roles {
		for _, role := range guild.Roles {
			if role.ID == roleID {
				permissions |= role.Permissions
				break
			}
		}
	}
	
	// If the user is the owner, they have all permissions
	if guild.OwnerID == member.User.ID {
		permissions = discordgo.PermissionAll
	}
	
	return permissions
}

func (b *Bot) guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	// Insert or update the guild in the database
	_, err := b.Db.Exec("INSERT INTO guilds (guild_id, owner_id, name) VALUES ($1, $2, $3) ON CONFLICT (guild_id) DO UPDATE SET owner_id = $2, name = $3", 
		event.Guild.ID, event.Guild.OwnerID, event.Guild.Name)
	if err != nil {
		log.Printf("Failed to insert/update guild: %v", err)
		return
	}
	
	// Insert the guild owner into the users table if not exists
	_, err = b.Db.Exec("INSERT INTO users (guild_id, user_id) VALUES ($1, $2) ON CONFLICT (guild_id, user_id) DO NOTHING", 
		event.Guild.ID, event.Guild.OwnerID)
	if err != nil {
		log.Printf("Failed to insert guild owner into users table: %v", err)
		return
	}
	
	log.Printf("Bot joined guild %s (ID: %s)", event.Guild.Name, event.Guild.ID)
}