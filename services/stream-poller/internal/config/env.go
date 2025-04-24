package config

import (
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	ClientID     string
	ClientSecret string
	PollInterval time.Duration
	RabbitMQURL  string
	RedisURL     string
	LogLevel     string
}

func Load() Config {
	return Config{
		ClientID:     mustGet("TWITCH_CLIENT_ID"),
		ClientSecret: mustGet("TWITCH_CLIENT_SECRET"),
		PollInterval: mustGetDuration("POLL_INTERVAL_SECONDS"),
		RabbitMQURL:  mustGet("RABBITMQ_URL"),
		RedisURL:     mustGet("REDIS_URL"),
		LogLevel:     getDefault("LOG_LEVEL", "info"),
	}
}

func mustGet(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("missing required environment variable: %s", key)
	}
	return val
}

func mustGetDuration(key string) time.Duration {
	seconds := mustGet(key)
	sec, err := strconv.Atoi(seconds)
	if err != nil {
		log.Fatalf("invalid duration value for %s: %v", key, err)
	}
	return time.Duration(sec) * time.Second
}

func getDefault(key string, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
