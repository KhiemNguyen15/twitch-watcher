package models

import "time"

// WatchType indicates whether a subscription watches a game or a streamer.
type WatchType string

const (
	WatchTypeGame     WatchType = "game"
	WatchTypeStreamer WatchType = "streamer"
)

// Subscription represents an active subscription stored in PostgreSQL.
type Subscription struct {
	ID             string    `json:"id"`
	DiscordWebhook string    `json:"discord_webhook"`
	WatchType      WatchType `json:"watch_type"`
	WatchTarget    string    `json:"watch_target"`
	Active         bool      `json:"active"`
	CreatedAt      time.Time `json:"created_at"`
}
