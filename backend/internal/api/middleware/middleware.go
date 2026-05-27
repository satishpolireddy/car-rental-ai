package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Logger logs each request with method, path, status, latency.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.WithFields(log.Fields{
			"method":  c.Request.Method,
			"path":    c.Request.URL.Path,
			"status":  c.Writer.Status(),
			"latency": time.Since(start),
			"ip":      c.ClientIP(),
		}).Info("request")
	}
}

// RateLimiter is a simple per-IP token bucket rate limiter.
// For production, replace with Redis-backed rate limiting for multi-instance support.
type rateLimiter struct {
	mu       sync.Mutex
	clients  map[string]*clientState
	rate     int           // requests allowed per window
	window   time.Duration
}

type clientState struct {
	count    int
	resetAt  time.Time
}

func NewRateLimiter(rate int, window time.Duration) gin.HandlerFunc {
	rl := &rateLimiter{
		clients: make(map[string]*clientState),
		rate:    rate,
		window:  window,
	}
	// Cleanup stale entries every minute
	go func() {
		for range time.Tick(time.Minute) {
			rl.mu.Lock()
			now := time.Now()
			for ip, state := range rl.clients {
				if now.After(state.resetAt) {
					delete(rl.clients, ip)
				}
			}
			rl.mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()
		rl.mu.Lock()
		state, exists := rl.clients[ip]
		if !exists || time.Now().After(state.resetAt) {
			rl.clients[ip] = &clientState{count: 1, resetAt: time.Now().Add(rl.window)}
			rl.mu.Unlock()
			c.Next()
			return
		}
		state.count++
		if state.count > rl.rate {
			rl.mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		rl.mu.Unlock()
		c.Next()
	}
}

// RequestID attaches a unique ID to each request for tracing.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = generateID()
		}
		c.Set("request_id", id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}

var (
	idMu    sync.Mutex
	idCount uint64
)

func generateID() string {
	idMu.Lock()
	idCount++
	v := idCount
	idMu.Unlock()
	return "req-" + uint64ToHex(v)
}

func uint64ToHex(n uint64) string {
	const hex = "0123456789abcdef"
	b := make([]byte, 16)
	for i := 15; i >= 0; i-- {
		b[i] = hex[n&0xf]
		n >>= 4
	}
	return string(b)
}
