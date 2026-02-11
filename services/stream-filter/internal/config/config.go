package config

import (
	"fmt"
	"os"
)

// Config holds all stream-filter configuration.
type Config struct {
	NATSUrl    string
	ValkeyAddr string
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{
		NATSUrl:    getEnv("NATS_URL", "nats://localhost:4222"),
		ValkeyAddr: getEnv("VALKEY_ADDR", "localhost:6379"),
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
