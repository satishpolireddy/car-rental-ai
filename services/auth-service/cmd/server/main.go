package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/satishpolireddy/car-rental-ai/auth-service/config"
	"github.com/satishpolireddy/car-rental-ai/auth-service/internal/api/handlers"
	"github.com/satishpolireddy/car-rental-ai/auth-service/internal/api/middleware"
	"github.com/satishpolireddy/car-rental-ai/auth-service/internal/models"
	"github.com/satishpolireddy/car-rental-ai/auth-service/internal/repository"
	"github.com/satishpolireddy/car-rental-ai/auth-service/internal/services"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	log.SetFormatter(&log.JSONFormatter{})

	// Database
	dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
	db, err := gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
	if err != nil {
		log.WithError(err).Fatal("db connect failed")
	}
	db.AutoMigrate(&models.User{}, &models.RefreshToken{})

	// Redis
	rdb := redis.NewClient(&redis.Options{Addr: cfg.Redis.Addr, Password: cfg.Redis.Password})

	// Wire dependencies
	userRepo := repository.NewUserRepository(db)
	refreshRepo := repository.NewRefreshTokenRepository(db)
	otpSvc := services.NewOTPService(rdb, cfg.Email)
	tokenSvc := services.NewTokenService(cfg.JWT, refreshRepo)
	h := handlers.NewAuthHandler(otpSvc, tokenSvc, userRepo)
	jwtMiddleware := middleware.NewJWTMiddleware(tokenSvc)

	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"service": "auth", "status": "ok"}) })

	auth := r.Group("/auth")
	{
		// Public endpoints
		auth.POST("/send-otp", h.SendOTP)
		auth.POST("/verify-otp", h.VerifyOTP)
		auth.POST("/refresh", h.Refresh)
		auth.POST("/logout", h.Logout)

		// Internal endpoint — called by API Gateway, not exposed publicly
		auth.POST("/validate", h.ValidateToken)

		// Protected
		auth.GET("/me", jwtMiddleware.Authenticate(), h.Me)
	}

	srv := &http.Server{Addr: ":" + cfg.Server.Port, Handler: r}
	go func() {
		log.WithField("port", cfg.Server.Port).Info("auth-service starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}
