package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

type ipLimiter struct {
	limiter *rate.Limiter
}

type RateLimiter struct {
	mu      sync.Mutex
	ips     map[string]*ipLimiter
	rps     rate.Limit
	burst   int
}

func NewRateLimiter(rps, burst int) *RateLimiter {
	return &RateLimiter{
		ips:   make(map[string]*ipLimiter),
		rps:   rate.Limit(rps),
		burst: burst,
	}
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	if l, ok := rl.ips[ip]; ok {
		return l.limiter
	}
	l := &ipLimiter{limiter: rate.NewLimiter(rl.rps, rl.burst)}
	rl.ips[ip] = l
	return l.limiter
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !rl.getLimiter(ip).Allow() {
			log.WithField("ip", ip).Warn("rate limit exceeded")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests, please slow down",
			})
			return
		}
		c.Next()
	}
}
