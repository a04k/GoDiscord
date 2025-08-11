package utils

import (
	"database/sql"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

type ReminderService struct {
	db      *sql.DB
	session *discordgo.Session
}

func NewReminderService(db *sql.DB, session *discordgo.Session) *ReminderService {
	return &ReminderService{
		db:      db,
		session: session,
	}
}

func (rs *ReminderService) Start() {
	log.Println("Starting reminder service...")
	ticker := time.NewTicker(1 * time.Minute) // Check every minute
	defer ticker.Stop()

	for range ticker.C {
		rs.checkAndSendReminders()
	}
}

func (rs *ReminderService) checkAndSendReminders() {
	// Get all unsent reminders that are due
	rows, err := rs.db.Query(`
		SELECT reminder_id, user_id, guild_id, message
		FROM reminders
		WHERE sent = FALSE AND remind_at <= NOW()
	`)
	if err != nil {
		log.Printf("Error querying reminders: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var reminderID int64
		var userID, message string
		var guildID sql.NullString

		if err := rows.Scan(&reminderID, &userID, &guildID, &message); err != nil {
			log.Printf("Error scanning reminder: %v", err)
			continue
		}

		// Send DM to user
		channel, err := rs.session.UserChannelCreate(userID)
		if err != nil {
			log.Printf("Error creating DM channel for user %s: %v", userID, err)
			// Mark reminder as sent even if we can't send it to avoid retrying indefinitely
			rs.markReminderAsSent(reminderID)
			continue
		}

		// Send the reminder message
		_, err = rs.session.ChannelMessageSend(channel.ID, 
			":bell: **Reminder** :bell:\n"+message)
		if err != nil {
			log.Printf("Error sending reminder to user %s: %v", userID, err)
			// Don't mark as sent if we failed to send, so it can be retried
			continue
		}

		// Mark reminder as sent
		rs.markReminderAsSent(reminderID)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating over reminders: %v", err)
	}
}

func (rs *ReminderService) markReminderAsSent(reminderID int64) {
	_, err := rs.db.Exec(`
		UPDATE reminders
		SET sent = TRUE
		WHERE reminder_id = $1
	`, reminderID)
	if err != nil {
		log.Printf("Error marking reminder %d as sent: %v", reminderID, err)
	}
}