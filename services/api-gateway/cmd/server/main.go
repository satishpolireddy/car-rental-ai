package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/satishpolireddy/car-rental-ai/api-gateway/config"
	"github.com/satishpolireddy/car-rental-ai/api-gateway/internal/middleware"
	"github.com/satishpolireddy/car-rental-ai/api-gateway/internal/proxy"
	log "github.com/sirupsen/logrus"
)

func main() {
	_ = godotenv.Load()

	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)

	cfg := config.Load()

	r := gin.New()
	r.Use(gin.Recovery())

	// Request logging
	r.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.WithFields(log.Fields{
			"method":  c.Request.Method,
			"path":    c.Request.URL.Path,
			"status":  c.Writer.Status(),
			"latency": time.Since(start).String(),
			"ip":      c.ClientIP(),
		}).Info("gateway request")
	})

	// Global rate limiter (per IP)
	rl := middleware.NewRateLimiter(cfg.RateLimitRPS, cfg.RateLimitBurst)
	r.Use(rl.Middleware())

	// Health check — no auth required
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "api-gateway"})
	})

	// ─── Public auth routes (pass-through to auth service) ────────────────────
	// These are unauthenticated — the auth service itself validates OTPs & issues tokens.
	authPassthrough := r.Group("/auth")
	{
		authPassthrough.Any("/*path", proxy.NewReverseProxy(cfg.AuthServiceURL))
	}

	// ─── Public payment webhook (no auth — Stripe signature verified downstream) ─
	r.POST("/payments/webhook", proxy.NewReverseProxy(cfg.PaymentServiceURL))

	// ─── Authenticated routes ──────────────────────────────────────────────────
	protected := r.Group("/")
	protected.Use(middleware.JWTAuth(cfg.AuthServiceURL))
	{
		// Booking service routes
		protected.Any("/api/v1/cars/*path", proxy.NewReverseProxy(cfg.BookingServiceURL))
		protected.Any("/api/v1/bookings/*path", proxy.NewReverseProxy(cfg.BookingServiceURL))
		protected.Any("/api/v1/customers/*path", proxy.NewReverseProxy(cfg.BookingServiceURL))

		// Payment service routes
		protected.Any("/payments/*path", proxy.NewReverseProxy(cfg.PaymentServiceURL))
	}

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Infof("api-gateway listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}
	log.Info("api-gateway stopped")
}
