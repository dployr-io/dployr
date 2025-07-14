package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// Simple in-memory rate limiter
type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}

// IsAuthenticated is a middleware that checks if
// the user has already been authenticated previously.
func IsAuthenticated(ctx *gin.Context) {
	if sessions.Default(ctx).Get("profile") == nil {
		ctx.Redirect(http.StatusSeeOther, "/")
	} else {
		ctx.Next()
	}
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *RateLimiter) IsAllowed(key string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	
	// Clean old requests
	if attempts, exists := rl.requests[key]; exists {
		var validTimes []time.Time
		for _, t := range attempts {
			if now.Sub(t) < rl.window {
				validTimes = append(validTimes, t)
			}
		}
		rl.requests[key] = validTimes
	}

	// Check if under limit
	if len(rl.requests[key]) >= rl.limit {
		return false
	}

	// Add current request
	rl.requests[key] = append(rl.requests[key], now)
	return true
}

// CleanupRateLimit periodically cleans up old rate limit entries
func CleanupRateLimit(rl *RateLimiter) {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			rl.mutex.Lock()
			now := time.Now()
			for email, times := range rl.requests {
				var validTimes []time.Time
				for _, t := range times {
					if now.Sub(t) < rl.window {
						validTimes = append(validTimes, t)
					}
				}
				if len(validTimes) == 0 {
					delete(rl.requests, email)
				} else {
					rl.requests[email] = validTimes
				}
			}
			rl.mutex.Unlock()
		}
	}()
}
