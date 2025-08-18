package utils

import (
	"sync"
	"time"
)

// RateLimiter controls the rate of command execution
type RateLimiter struct {
	limits map[string]*userLimit
	mu     sync.Mutex
}

// userLimit tracks rate limiting for a specific user
type userLimit struct {
	lastAccess time.Time
	count      int
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		limits: make(map[string]*userLimit),
	}
}

// Allow checks if a user is allowed to execute a command
// Returns true if allowed, false if rate limited
func (rl *RateLimiter) Allow(userID, command string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	key := userID + ":" + command
	now := time.Now()
	
	limit, exists := rl.limits[key]
	if !exists {
		// First time user is accessing this command
		rl.limits[key] = &userLimit{
			lastAccess: now,
			count:      1,
		}
		return true
	}
	
	// Check if enough time has passed since last access
	if now.Sub(limit.lastAccess) >= time.Minute {
		// Reset the counter if a minute has passed
		limit.lastAccess = now
		limit.count = 1
		return true
	}
	
	// Check if user has exceeded the rate limit (15 commands per minute)
	if limit.count >= 15 {
		return false
	}
	
	// Increment the counter
	limit.count++
	return true
}

// GetRetryAfter returns the time in seconds until the user can try again
func (rl *RateLimiter) GetRetryAfter(userID, command string) int {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	key := userID + ":" + command
	limit, exists := rl.limits[key]
	if !exists {
		return 0
	}
	
	elapsed := time.Since(limit.lastAccess)
	if elapsed >= time.Minute {
		return 0
	}
	
	return int((time.Minute - elapsed).Seconds())
}