package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration loaded from environment variables.
// Designed for 12-factor app compatibility — no hardcoded secrets.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Azure    AzureConfig
	ETL      ETLConfig
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	Environment  string
}

type DatabaseConfig struct {
	Host            string
	Port            string
	Name            string
	User            string
	Password        string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	PoolSize int
}

type AzureConfig struct {
	OpenAIEndpoint   string
	OpenAIKey        string
	DeploymentName   string
	EmbeddingModel   string
}

type ETLConfig struct {
	BatchSize     int
	Workers       int
	PollInterval  time.Duration
}

func Load() (*Config, error) {
	return &Config{
		Server: ServerConfig{
			Port:         getEnv("PORT", "8080"),
			ReadTimeout:  getDuration("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout: getDuration("SERVER_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:  getDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
			Environment:  getEnv("ENVIRONMENT", "development"),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "1433"),
			Name:            getEnv("DB_NAME", "carrental"),
			User:            getEnv("DB_USER", "sa"),
			Password:        getEnv("DB_PASSWORD", ""),
			MaxOpenConns:    getInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: getDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getInt("REDIS_DB", 0),
			PoolSize: getInt("REDIS_POOL_SIZE", 20),
		},
		Azure: AzureConfig{
			OpenAIEndpoint: getEnv("AZURE_OPENAI_ENDPOINT", ""),
			OpenAIKey:      getEnv("AZURE_OPENAI_KEY", ""),
			DeploymentName: getEnv("AZURE_OPENAI_DEPLOYMENT", "gpt-4o"),
			EmbeddingModel: getEnv("AZURE_OPENAI_EMBEDDING", "text-embedding-3-small"),
		},
		ETL: ETLConfig{
			BatchSize:    getInt("ETL_BATCH_SIZE", 500),
			Workers:      getInt("ETL_WORKERS", 4),
			PollInterval: getDuration("ETL_POLL_INTERVAL", 5*time.Minute),
		},
	}, nil
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultVal
}

func getDuration(key string, defaultVal time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return defaultVal
}
