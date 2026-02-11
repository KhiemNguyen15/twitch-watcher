package models

import "time"

// SubscriptionRef links a subscription to its Discord webhook.
type SubscriptionRef struct {
	SubscriptionID string `json:"subscription_id"`
	DiscordWebhook string `json:"discord_webhook"`
}

// TwitchStream represents a raw Twitch stream from the Helix API.
type TwitchStream struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	UserLogin    string    `json:"user_login"`
	UserName     string    `json:"user_name"`
	GameID       string    `json:"game_id"`
	GameName     string    `json:"game_name"`
	Type         string    `json:"type"`
	Title        string    `json:"title"`
	ViewerCount  int       `json:"viewer_count"`
	StartedAt    time.Time `json:"started_at"`
	Language     string    `json:"language"`
	ThumbnailURL string    `json:"thumbnail_url"`
}

// StreamEvent is published to twitch.streams.raw by stream-poller.
type StreamEvent struct {
	StreamID      string           `json:"stream_id"`
	UserLogin     string           `json:"user_login"`
	UserName      string           `json:"user_name"`
	GameID        string           `json:"game_id"`
	GameName      string           `json:"game_name"`
	Title         string           `json:"title"`
	ViewerCount   int              `json:"viewer_count"`
	StartedAt     time.Time        `json:"started_at"`
	ThumbnailURL  string           `json:"thumbnail_url"`
	StreamURL     string           `json:"stream_url"`
	Subscriptions []SubscriptionRef `json:"subscriptions"`
	PolledAt      time.Time        `json:"polled_at"`
}
