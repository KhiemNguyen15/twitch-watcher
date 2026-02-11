package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/khiemnguyen15/twitch-watcher/pkg/models"
	"github.com/khiemnguyen15/twitch-watcher/services/subscription-service/internal/repository"
)

// ErrInvalidWebhook is returned when the provided discord_webhook URL is malformed.
var ErrInvalidWebhook = errors.New("invalid Discord webhook URL")

// ErrDuplicate is forwarded from the repository layer.
var ErrDuplicate = repository.ErrDuplicate

// ErrNotFound is forwarded from the repository layer.
var ErrNotFound = repository.ErrNotFound

// SubscriptionService contains business logic for managing subscriptions.
type SubscriptionService struct {
	repo *repository.Repository
}

// New creates a new SubscriptionService.
func New(repo *repository.Repository) *SubscriptionService {
	return &SubscriptionService{repo: repo}
}

// Create validates inputs and creates a new subscription.
func (s *SubscriptionService) Create(ctx context.Context, webhook string, watchType models.WatchType, watchTarget string) (*models.Subscription, error) {
	if err := validateDiscordWebhook(webhook); err != nil {
		return nil, err
	}
	if watchType != models.WatchTypeGame && watchType != models.WatchTypeStreamer {
		return nil, fmt.Errorf("watch_type must be 'game' or 'streamer'")
	}
	if strings.TrimSpace(watchTarget) == "" {
		return nil, fmt.Errorf("watch_target must not be empty")
	}
	return s.repo.Create(ctx, webhook, watchType, watchTarget)
}

// GetByID retrieves a subscription by ID.
func (s *SubscriptionService) GetByID(ctx context.Context, id string) (*models.Subscription, error) {
	return s.repo.GetByID(ctx, id)
}

// Delete soft-deletes a subscription.
func (s *SubscriptionService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// ListActive returns all active subscriptions (used by stream-poller internal endpoint).
func (s *SubscriptionService) ListActive(ctx context.Context) ([]models.Subscription, error) {
	return s.repo.ListActive(ctx)
}

// validateDiscordWebhook ensures the URL is a well-formed Discord webhook.
func validateDiscordWebhook(raw string) error {
	u, err := url.ParseRequestURI(raw)
	if err != nil {
		return ErrInvalidWebhook
	}
	if u.Scheme != "https" {
		return ErrInvalidWebhook
	}
	if u.Host != "discord.com" && u.Host != "discordapp.com" {
		return ErrInvalidWebhook
	}
	if !strings.HasPrefix(u.Path, "/api/webhooks/") {
		return ErrInvalidWebhook
	}
	return nil
}
