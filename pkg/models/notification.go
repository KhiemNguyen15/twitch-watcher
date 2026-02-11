package models

import "time"

// NotificationPayload is published to twitch.streams.new by stream-filter.
// One message is published per SubscriptionRef (fan-out).
type NotificationPayload struct {
	SubscriptionID string    `json:"subscription_id"`
	DiscordWebhook string    `json:"discord_webhook"`
	StreamID       string    `json:"stream_id"`
	UserLogin      string    `json:"user_login"`
	UserName       string    `json:"user_name"`
	GameName       string    `json:"game_name"`
	Title          string    `json:"title"`
	ViewerCount    int       `json:"viewer_count"`
	StartedAt      time.Time `json:"started_at"`
	ThumbnailURL   string    `json:"thumbnail_url"`
	StreamURL      string    `json:"stream_url"`
}
