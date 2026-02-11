package config

import (
	"fmt"
	"os"
)

// Config holds all service configuration loaded from environment variables.
type Config struct {
	HTTPAddr       string
	DatabaseURL    string
	InternalAPIKey string
}

// Load reads configuration from environment variables, returning an error for any missing required value.
func Load() (*Config, error) {
	cfg := &Config{
		HTTPAddr:       getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		InternalAPIKey: os.Getenv("INTERNAL_API_KEY"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
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
