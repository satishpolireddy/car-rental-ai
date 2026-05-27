package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port            string
	DBDSN           string
	StripeSecretKey string
	StripeWebhookSecret string
	AuthServiceURL  string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:                getEnv("PORT", "8082"),
		StripeSecretKey:     os.Getenv("STRIPE_SECRET_KEY"),
		StripeWebhookSecret: os.Getenv("STRIPE_WEBHOOK_SECRET"),
		AuthServiceURL:      getEnv("AUTH_SERVICE_URL", "http://auth-service:8081"),
	}

	// Build SQL Server DSN
	host := os.Getenv("DB_HOST")
	port := getEnv("DB_PORT", "1433")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	dbName := getEnv("DB_NAME", "carrental")

	if host == "" || user == "" || pass == "" {
		return nil, fmt.Errorf("DB_HOST, DB_USER, DB_PASSWORD are required")
	}
	cfg.DBDSN = fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s",
		user, pass, host, port, dbName)

	if cfg.StripeSecretKey == "" {
		return nil, fmt.Errorf("STRIPE_SECRET_KEY is required")
	}
	if cfg.StripeWebhookSecret == "" {
		return nil, fmt.Errorf("STRIPE_WEBHOOK_SECRET is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
