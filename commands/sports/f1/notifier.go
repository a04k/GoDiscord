package f1

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	checkInterval      = 1 * time.Hour    // Check for new schedules every hour
	notificationOffset = 30 * time.Minute // Notify 30 minutes before a session
)

// Event represents a single F1 event (Grand Prix) from our schedule.
// This struct is defined in notifier.go and used by other f1 commands.

// F1Notifier holds the Discord session, database connection, and notification state.
type F1Notifier struct {
	Session             *discordgo.Session
	DB                  *sql.DB
	lastNotifiedSession map[string]time.Time // Map session ID to last notification time
	lastNotifiedEvent   map[string]time.Time // Map event ID to last notification time
}

// NewF1Notifier creates a new F1Notifier instance.
func NewF1Notifier(s *discordgo.Session, db *sql.DB) *F1Notifier {
	return &F1Notifier{
		Session:             s,
		DB:                  db,
		lastNotifiedSession: make(map[string]time.Time),
		lastNotifiedEvent:   make(map[string]time.Time),
	}
}

// Start begins the F1 notification routine.
func (fn *F1Notifier) Start() {
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	// Run immediately on start
	fn.checkAndNotify()

	for range ticker.C {
		fn.checkAndNotify()
	}
}

func (fn *F1Notifier) checkAndNotify() {
	log.Println("Checking for upcoming F1 sessions...")

	events, err := FetchF1Events()
	if err != nil {
		log.Printf("Error fetching F1 events: %v", err)
		return
	}

	now := time.Now().UTC()

	// Notify for upcoming events (race weekends) with full schedule
	for _, event := range events {
		if len(event.Sessions) == 0 {
			continue
		}
		
		// Get the first session of the event (typically Practice 1)
		firstSession := event.Sessions[0]
		eventStartTime, err := time.Parse(time.RFC3339, firstSession.Date)
		if err != nil {
			log.Printf("Error parsing event start date for %s: %v", event.Name, err)
			continue
		}

		// Notify 24 hours before the event starts
		if eventStartTime.After(now) && eventStartTime.Before(now.Add(24*time.Hour)) {
			eventID := fmt.Sprintf("%s-%s", event.Name, firstSession.Date)
			if _, ok := fn.lastNotifiedEvent[eventID]; !ok {
				// Build the full weekend schedule message
				scheduleMessage := fmt.Sprintf(`
**ITS %s WEEKEND!** Here's the timetable (in your timezone) for each session:

`, event.Name)
				
				for _, session := range event.Sessions {
					sessionTime, err := time.Parse(time.RFC3339, session.Date)
					if err != nil {
						continue
					}
					scheduleMessage += fmt.Sprintf("**%s**: <t:%d:F>\n", session.Name, sessionTime.Unix())
				}
				
				fn.notifySubscribers(scheduleMessage)
				fn.lastNotifiedEvent[eventID] = now
			}
		}
	}

	// Notify for upcoming sessions (except the first one which is already covered above)
	for _, event := range events {
		for i, session := range event.Sessions {
			// Skip the first session as it's already covered in the event notification
			if i == 0 {
				continue
			}
			
			sessionTime, err := time.Parse(time.RFC3339, session.Date)
			if err != nil {
				log.Printf("Error parsing session date for %s: %v", session.Name, err)
				continue
			}

			// Notify 30 minutes before each session
			if sessionTime.After(now) && sessionTime.Before(now.Add(notificationOffset)) {
				sessionID := fmt.Sprintf("%s-%s", event.Name, session.Name)
				if _, ok := fn.lastNotifiedSession[sessionID]; !ok {
					fn.notifySubscribers(fmt.Sprintf(`
F1 Session Reminder: **%s** (%s) starts in 30 minutes! (<t:%d:R>)`,
						session.Name, event.Name, sessionTime.Unix(),
					))
					fn.lastNotifiedSession[sessionID] = now
				}
			}
		}
	}
}

func (fn *F1Notifier) notifySubscribers(message string) {
	rows, err := fn.DB.Query("SELECT user_id FROM f1_subscriptions")
	if err != nil {
		log.Printf("Error querying F1 subscriptions: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			log.Printf("Error scanning user ID from F1 subscriptions: %v", err)
			continue
		}

		// Create DM channel and send message
		channel, err := fn.Session.UserChannelCreate(userID)
		if err != nil {
			log.Printf("Error creating DM channel for user %s: %v", userID, err)
			continue
		}

		_, err = fn.Session.ChannelMessageSend(channel.ID, message)
		if err != nil {
			log.Printf("Error sending DM to user %s: %v", userID, err)
		} else {
			log.Printf("Sent F1 notification to user %s", userID)
		}
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating F1 subscriptions: %v", err)
	}
}

// FetchF1Events reads the F1 schedule from the local JSON file.
func FetchF1Events() ([]Event, error) {
	// Get the directory of the current file
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	
	// Read the JSON file
	jsonFile, err := os.ReadFile(filepath.Join(dir, "f1_schedule_2025.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to read F1 schedule file: %w", err)
	}

	var events []Event
	err = json.Unmarshal(jsonFile, &events)
	if err != nil {
		return nil, fmt.Errorf("error decoding F1 events: %w", err)
	}

	// Sort events by the date of their first session
	sort.Slice(events, func(i, j int) bool {
		if len(events[i].Sessions) == 0 || len(events[j].Sessions) == 0 {
			return false
		}
		timeI, _ := time.Parse(time.RFC3339, events[i].Sessions[0].Date)
		timeJ, _ := time.Parse(time.RFC3339, events[j].Sessions[0].Date)
		return timeI.Before(timeJ)
	})

	return events, nil
}
