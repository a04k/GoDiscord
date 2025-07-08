package f1notifier

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	f1APIURL           = "https://api.openf1.org/v1/meetings"
	checkInterval      = 1 * time.Hour    // Check for new schedules every hour
	notificationOffset = 30 * time.Minute // Notify 30 minutes before a session
)

// Meeting represents a single F1 meeting (Grand Prix weekend) from OpenF1 API.
type Meeting struct {
	MeetingName string    `json:"meeting_name"`
	Location    string    `json:"location"`
	CountryName string    `json:"country_name"`
	DateStart   string    `json:"date_start"`
	DateEnd     string    `json:"date_end"`
	Sessions    []Session `json:"sessions"`
}

// Session represents a single session within an F1 meeting.
type Session struct {
	SessionName string `json:"session_name"`
	DateStart   string `json:"date_start"`
	DateEnd     string `json:"date_end"`
}

// F1Notifier holds the Discord session, database connection, and notification state.
type F1Notifier struct {
	Session             *discordgo.Session
	DB                  *sql.DB
	lastNotifiedSession map[string]time.Time // Map session ID to last notification time
	lastNotifiedWeekend map[string]time.Time // Map meeting ID to last notification time
}

// NewF1Notifier creates a new F1Notifier instance.
func NewF1Notifier(s *discordgo.Session, db *sql.DB) *F1Notifier {
	return &F1Notifier{
		Session:             s,
		DB:                  db,
		lastNotifiedSession: make(map[string]time.Time),
		lastNotifiedWeekend: make(map[string]time.Time),
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

	meetings, err := fetchF1Meetings()
	if err != nil {
		log.Printf("Error fetching F1 meetings: %v", err)
		return
	}

	now := time.Now().UTC()

	// Notify for upcoming weekends
	for _, meeting := range meetings {
		meetingStartTime, err := time.Parse(time.RFC3339, meeting.DateStart)
		if err != nil {
			log.Printf("Error parsing meeting start date for %s: %v", meeting.MeetingName, err)
			continue
		}

		// Notify 24 hours before the weekend starts
		if meetingStartTime.After(now) && meetingStartTime.Before(now.Add(24*time.Hour)) {
			if _, ok := fn.lastNotifiedWeekend[meeting.MeetingName]; !ok {
				fn.notifySubscribers(fmt.Sprintf(`
Upcoming F1 Weekend: **%s** in %s, %s
Starts: <t:%d:F> (<t:%d:R>)`,
					meeting.MeetingName, meeting.Location, meeting.CountryName, meetingStartTime.Unix(), meetingStartTime.Unix(),
				))
				fn.lastNotifiedWeekend[meeting.MeetingName] = now
			}
		}
	}

	// Notify for upcoming sessions
	for _, meeting := range meetings {
		for _, session := range meeting.Sessions {
			sessionStartTime, err := time.Parse(time.RFC3339, session.DateStart)
			if err != nil {
				log.Printf("Error parsing session start date for %s: %v", session.SessionName, err)
				continue
			}

			// Notify 30 minutes before each session
			if sessionStartTime.After(now) && sessionStartTime.Before(now.Add(notificationOffset)) {
				sessionID := fmt.Sprintf("%s-%s", meeting.MeetingName, session.SessionName)
				if _, ok := fn.lastNotifiedSession[sessionID]; !ok {
					fn.notifySubscribers(fmt.Sprintf(`
F1 Session Reminder: **%s** of %s starts in 30 minutes! (<t:%d:R>)`,
						session.SessionName, meeting.MeetingName, sessionStartTime.Unix(),
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

		user, err := fn.Session.User(userID)
		if err != nil {
			log.Printf("Error fetching user %s: %v", userID, err)
			continue
		}

		channel, err := fn.Session.UserChannelCreate(user.ID)
		if err != nil {
			log.Printf("Error creating DM channel for user %s: %v", user.Username, err)
			continue
		}

		_, err = fn.Session.ChannelMessageSend(channel.ID, message)
		if err != nil {
			log.Printf("Error sending DM to user %s: %v", user.Username, err)
		} else {
			log.Printf("Sent F1 notification to %s", user.Username)
		}
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating F1 subscriptions: %v", err)
	}
}

func fetchF1Meetings() ([]Meeting, error) {
	resp, err := http.Get(f1APIURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch F1 meetings: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("F1 API returned non-200 status: %d", resp.StatusCode)
	}

	var meetings []Meeting
	err = json.NewDecoder(resp.Body).Decode(&meetings)
	if err != nil {
		return nil, fmt.Errorf("error decoding F1 meetings: %w", err)
	}

	// Sort meetings by start date to easily find the next one
	sort.Slice(meetings, func(i, j int) bool {
		timeI, _ := time.Parse(time.RFC3339, meetings[i].DateStart)
		timeJ, _ := time.Parse(time.RFC3339, meetings[j].DateStart)
		return timeI.Before(timeJ)
	})

	return meetings, nil
}
