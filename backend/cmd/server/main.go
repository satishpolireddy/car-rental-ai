package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/satishpolireddy/car-rental-ai/config"
	"github.com/satishpolireddy/car-rental-ai/internal/api/handlers"
	"github.com/satishpolireddy/car-rental-ai/internal/api/middleware"
	"github.com/satishpolireddy/car-rental-ai/internal/etl"
	"github.com/satishpolireddy/car-rental-ai/internal/models"
	"github.com/satishpolireddy/car-rental-ai/internal/repository"
	"github.com/satishpolireddy/car-rental-ai/internal/services"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Load .env in development
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.WithError(err).Fatal("failed to load config")
	}

	log.SetFormatter(&log.JSONFormatter{})
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
		log.SetLevel(log.InfoLevel)
	}

	// Database — connection pool tuned for horizontal scale
	dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s",
		cfg.Database.User, cfg.Database.Password,
		cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)

	db, err := gorm.Open(sqlserver.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.WithError(err).Fatal("failed to connect to database")
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	// Auto-migrate schema
	if err := db.AutoMigrate(
		&models.Car{},
		&models.Customer{},
		&models.Booking{},
		&models.AIRecommendation{},
		&models.ETLLog{},
	); err != nil {
		log.WithError(err).Fatal("database migration failed")
	}

	// Redis — pooled connections for caching
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.WithError(err).Warn("redis unavailable — caching disabled")
	}

	// Repositories & services (dependency injection)
	carRepo := repository.NewCarRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	customerRepo := repository.NewCustomerRepository(db)
	bookingService := services.NewBookingService(bookingRepo, carRepo)
	aiService, err := services.NewAIService(cfg.Azure, rdb, carRepo)
	if err != nil {
		log.WithError(err).Fatal("failed to initialise AI service")
	}

	// ETL pipeline — runs on a background goroutine
	pipeline := etl.NewPipeline(db, carRepo, cfg.ETL)
	ctx, cancel := context.WithCancel(context.Background())
	pipeline.Start(ctx)

	// HTTP router
	h := handlers.NewHandler(carRepo, bookingService, aiService, customerRepo, bookingRepo)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.RequestID())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", os.Getenv("FRONTEND_URL")},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// Rate limiting: 100 req/min per IP
	r.Use(middleware.NewRateLimiter(100, time.Minute))

	r.GET("/health", h.HealthCheck)

	v1 := r.Group("/api/v1")
	{
		v1.POST("/cars/search", h.SearchCars)
		v1.GET("/cars/:id", h.GetCar)
		v1.GET("/locations", h.ListLocations)

		v1.POST("/bookings", h.CreateBooking)
		v1.GET("/bookings/:id", h.GetBooking)
		v1.DELETE("/bookings/:id", h.CancelBooking)
		v1.GET("/customers/:id/bookings", h.GetCustomerBookings)
		v1.GET("/customers/:id/dashboard", h.GetCustomerDashboard)

		v1.POST("/ai/recommend", h.AIRecommend)
	}

	// Graceful shutdown
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		log.WithField("port", cfg.Server.Port).Info("server starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Fatal("server error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down gracefully...")
	cancel() // stop ETL
	pipeline.Stop()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.WithError(err).Error("server forced to shutdown")
	}
	log.Info("server exited")
}
