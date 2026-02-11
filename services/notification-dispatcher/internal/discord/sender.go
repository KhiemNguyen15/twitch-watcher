package discord

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"fmt"
	"net/http"
	"time"

	"github.com/khiemnguyen15/twitch-watcher/pkg/models"
)

// embed represents a Discord rich embed.
type embed struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	Color       int       `json:"color"`
	Timestamp   time.Time `json:"timestamp"`
	Image       *image    `json:"image,omitempty"`
	Footer      *footer   `json:"footer,omitempty"`
	Fields      []field   `json:"fields,omitempty"`
}

type image struct {
	URL string `json:"url"`
}

type footer struct {
	Text string `json:"text"`
}

type field struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type webhookPayload struct {
	Embeds []embed `json:"embeds"`
}

// Sender sends Discord webhook notifications.
type Sender struct {
	httpClient *http.Client
}

// New creates a Sender.
func New() *Sender {
	return &Sender{
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// Send posts a rich embed to the given Discord webhook URL.
// Retries up to 3 times with exponential backoff on non-2xx responses.
func (s *Sender) Send(ctx context.Context, payload models.NotificationPayload) error {
	e := buildEmbed(payload)
	body, err := json.Marshal(webhookPayload{Embeds: []embed{e}})
	if err != nil {
		return fmt.Errorf("marshal discord payload: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			wait := time.Duration(1<<uint(attempt)) * time.Second
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, payload.DiscordWebhook,
			bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("build discord request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := s.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("discord webhook POST: %w", err)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}
		lastErr = fmt.Errorf("discord webhook returned %d", resp.StatusCode)
	}

	return lastErr
}

// buildEmbed constructs the Discord rich embed for a stream notification.
func buildEmbed(p models.NotificationPayload) embed {
	return embed{
		Title:       fmt.Sprintf("%s is live on Twitch!", p.UserName),
		Description: p.Title,
		URL:         p.StreamURL,
		Color:       0x9146FF, // Twitch purple
		Timestamp:   p.StartedAt,
		Image:       &image{URL: p.ThumbnailURL},
		Footer:      &footer{Text: "Twitch Watcher"},
		Fields: []field{
			{Name: "Game", Value: p.GameName, Inline: true},
			{Name: "Viewers", Value: fmt.Sprintf("%d", p.ViewerCount), Inline: true},
		},
	}
}
