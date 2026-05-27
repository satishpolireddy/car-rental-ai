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
	"github.com/satishpolireddy/car-rental-ai/payment-service/config"
	"github.com/satishpolireddy/car-rental-ai/payment-service/internal/api/handlers"
	"github.com/satishpolireddy/car-rental-ai/payment-service/internal/api/middleware"
	"github.com/satishpolireddy/car-rental-ai/payment-service/internal/models"
	"github.com/satishpolireddy/car-rental-ai/payment-service/internal/repository"
	"github.com/satishpolireddy/car-rental-ai/payment-service/internal/services"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

func main() {
	_ = godotenv.Load()

	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	// Database
	db, err := gorm.Open(sqlserver.Open(cfg.DBDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("db connect error: %v", err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	if err := db.AutoMigrate(&models.Payment{}); err != nil {
		log.Fatalf("db migrate error: %v", err)
	}

	// Wire dependencies
	repo := repository.NewPaymentRepository(db)
	svc := services.NewPaymentService(repo, cfg.StripeSecretKey)
	h := handlers.NewPaymentHandler(svc, cfg.StripeWebhookSecret)

	// Router
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.WithFields(log.Fields{
			"method":   c.Request.Method,
			"path":     c.Request.URL.Path,
			"status":   c.Writer.Status(),
			"latency":  time.Since(start).String(),
		}).Info("request")
	})

	r.GET("/health", handlers.Health)

	// Stripe webhooks — NO auth, but signature-verified inside handler
	r.POST("/payments/webhook", h.Webhook)

	// Authenticated payment routes
	auth := r.Group("/")
	auth.Use(middleware.AuthMiddleware(cfg.AuthServiceURL))
	{
		auth.POST("/payments", h.CreatePayment)
		auth.POST("/payments/refund", h.Refund)
		auth.GET("/payments/booking/:booking_id", h.GetByBooking)
	}

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Infof("payment-service listening on :%s", cfg.Port)
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
	log.Info("payment-service stopped")
}
