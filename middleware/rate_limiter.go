package middleware

import (
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	requests map[string][]time.Time
	mu       sync.RWMutex
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *RateLimiter) Allow(username string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Clean old requests
	requests := rl.requests[username]
	validRequests := []time.Time{}
	for _, req := range requests {
		if req.After(cutoff) {
			validRequests = append(validRequests, req)
		}
	}

	if len(validRequests) >= rl.limit {
		return false
	}

	validRequests = append(validRequests, now)
	rl.requests[username] = validRequests
	return true
}

var uploadLimiter = NewRateLimiter(10, time.Minute) // 10 uploads per minute

func RateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := GetSessionUser(r)
		if username != "" && !uploadLimiter.Allow(username) {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}
