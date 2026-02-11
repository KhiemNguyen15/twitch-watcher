package config

import (
	"fmt"
	"os"
)

// Config holds all notification-dispatcher configuration.
type Config struct {
	NATSUrl string
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{
		NATSUrl: getEnv("NATS_URL", "nats://localhost:4222"),
	}

	if cfg.NATSUrl == "" {
		return nil, fmt.Errorf("NATS_URL is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
