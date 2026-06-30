package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter is an in-memory, per-key token-bucket rate limiter. It is intended
// for protecting auth endpoints against brute force. For multi-instance
// deployments a shared store (e.g. Redis) would be required.
type RateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*rate.Limiter
	rps      rate.Limit
	burst    int
	lastSeen map[string]time.Time
}

// NewRateLimiter builds a limiter allowing `burst` requests with refill of `rps`
// tokens per second per key.
func NewRateLimiter(rps float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		buckets:  make(map[string]*rate.Limiter),
		rps:      rate.Limit(rps),
		burst:    burst,
		lastSeen: make(map[string]time.Time),
	}
	go rl.cleanupLoop()
	return rl
}

func (rl *RateLimiter) limiterFor(key string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.lastSeen[key] = time.Now()
	if l, ok := rl.buckets[key]; ok {
		return l
	}
	l := rate.NewLimiter(rl.rps, rl.burst)
	rl.buckets[key] = l
	return l
}

// cleanupLoop evicts stale buckets to bound memory usage.
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		cutoff := time.Now().Add(-15 * time.Minute)
		for k, seen := range rl.lastSeen {
			if seen.Before(cutoff) {
				delete(rl.lastSeen, k)
				delete(rl.buckets, k)
			}
		}
		rl.mu.Unlock()
	}
}

// Middleware returns a Gin middleware that rate-limits by client IP.
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		if !rl.limiterFor(key).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{
					"code":    "RATE_LIMITED",
					"message": "Too many requests, please slow down",
				},
			})
			return
		}
		c.Next()
	}
}
