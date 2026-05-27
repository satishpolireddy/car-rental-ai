package config

import "os"

type Config struct {
	Port            string
	AuthServiceURL  string
	BookingServiceURL string
	PaymentServiceURL string
	// Rate limit: requests per second per IP
	RateLimitRPS int
	RateLimitBurst int
}

func Load() *Config {
	return &Config{
		Port:              getEnv("PORT", "8080"),
		AuthServiceURL:    getEnv("AUTH_SERVICE_URL", "http://auth-service:8081"),
		BookingServiceURL: getEnv("BOOKING_SERVICE_URL", "http://booking-service:8083"),
		PaymentServiceURL: getEnv("PAYMENT_SERVICE_URL", "http://payment-service:8082"),
		RateLimitRPS:      20,
		RateLimitBurst:    40,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
