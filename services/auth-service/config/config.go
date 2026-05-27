package config

import "os"

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Email    EmailConfig
}

type ServerConfig struct {
	Port string
}
type DatabaseConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
}
type RedisConfig struct {
	Addr     string
	Password string
}
type JWTConfig struct {
	Secret string
}
type EmailConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	From     string
}

func Load() *Config {
	return &Config{
		Server:   ServerConfig{Port: getEnv("PORT", "8081")},
		Database: DatabaseConfig{
			Host: getEnv("DB_HOST", "localhost"), Port: getEnv("DB_PORT", "1433"),
			Name: getEnv("DB_NAME", "carrental"), User: getEnv("DB_USER", "sa"),
			Password: getEnv("DB_PASSWORD", ""),
		},
		Redis: RedisConfig{Addr: getEnv("REDIS_ADDR", "localhost:6379"), Password: getEnv("REDIS_PASSWORD", "")},
		JWT:   JWTConfig{Secret: getEnv("JWT_SECRET", "change-me-in-production-use-32-char-min")},
		Email: EmailConfig{
			Host: getEnv("SMTP_HOST", "smtp.gmail.com"), Port: getEnv("SMTP_PORT", "587"),
			User: getEnv("SMTP_USER", ""), Password: getEnv("SMTP_PASSWORD", ""),
			From: getEnv("SMTP_FROM", "noreply@driveai.com"),
		},
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
