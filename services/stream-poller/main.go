package main

import (
	"net/http"
	"os"
	"time"

	"github.com/khiemnguyen15/twitch-watcher/services/stream-poller/internal/config"
	"github.com/khiemnguyen15/twitch-watcher/services/stream-poller/internal/twitch"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg := config.Load()

	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Enable pretty logging in development
	if zerolog.GlobalLevel() == zerolog.DebugLevel {
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.RFC3339,
		})
	}

	log.Info().Msg("Loaded app config")

	tokenManager := twitch.NewTokenManager(cfg.ClientID, cfg.ClientSecret)

	token, err := tokenManager.GetAccessToken()
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Error getting Twitch access token")
	}
	log.Info().Msg("Fetched Twitch access token")

	// DEBUG: Test the access token
	req, err := http.NewRequest(http.MethodGet, "https://api.twitch.tv/helix/streams", nil)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Error creating streams request")
	}
	req.Header.Set("Client-Id", cfg.ClientID)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Error getting Twitch streams")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatal().
			Str("status", resp.Status).
			Msg("Unexpected status code")
	}
	log.Info().Msg("Successfully fetched Twitch streams")
}
