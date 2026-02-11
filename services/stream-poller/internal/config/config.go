package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all stream-poller configuration.
type Config struct {
	TwitchClientID     string
	TwitchClientSecret string
	NATSUrl            string
	ValkeyAddr         string
	SubscriptionSvcURL string
	InternalAPIKey     string
	PollInterval       time.Duration
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	pollSec, _ := strconv.Atoi(getEnv("POLL_INTERVAL_SECONDS", "60"))
	if pollSec < 1 {
		pollSec = 60
	}

	cfg := &Config{
		TwitchClientID:     os.Getenv("TWITCH_CLIENT_ID"),
		TwitchClientSecret: os.Getenv("TWITCH_CLIENT_SECRET"),
		NATSUrl:            getEnv("NATS_URL", "nats://localhost:4222"),
		ValkeyAddr:         getEnv("VALKEY_ADDR", "localhost:6379"),
		SubscriptionSvcURL: getEnv("SUBSCRIPTION_SVC_URL", "http://localhost:8080"),
		InternalAPIKey:     os.Getenv("INTERNAL_API_KEY"),
		PollInterval:       time.Duration(pollSec) * time.Second,
	}

	if cfg.TwitchClientID == "" {
		return nil, fmt.Errorf("TWITCH_CLIENT_ID is required")
	}
	if cfg.TwitchClientSecret == "" {
		return nil, fmt.Errorf("TWITCH_CLIENT_SECRET is required")
	}
	if cfg.InternalAPIKey == "" {
		return nil, fmt.Errorf("INTERNAL_API_KEY is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
