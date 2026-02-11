package service

import (
	"context"
	"errors"
	"testing"

	"github.com/khiemnguyen15/twitch-watcher/pkg/models"
)

// svc with a nil repo is valid for tests that never reach the repository.
var svc = &SubscriptionService{}

func TestCreate_InvalidWebhook(t *testing.T) {
	cases := []struct {
		name    string
		webhook string
	}{
		{"empty", ""},
		{"plain string", "not-a-url"},
		{"http scheme", "http://discord.com/api/webhooks/1234/token"},
		{"wrong host", "https://example.com/api/webhooks/1234/token"},
		{"missing path prefix", "https://discord.com/webhooks/1234/token"},
		{"discordapp valid host missing path", "https://discordapp.com/not/webhooks"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.Create(context.Background(), tc.webhook, models.WatchTypeGame, "Fortnite")
			if !errors.Is(err, ErrInvalidWebhook) {
				t.Errorf("webhook %q: got error %v, want ErrInvalidWebhook", tc.webhook, err)
			}
		})
	}
}

func TestValidateDiscordWebhook_ValidHosts(t *testing.T) {
	webhooks := []string{
		"https://discord.com/api/webhooks/1234/token",
		"https://discordapp.com/api/webhooks/1234/token",
		"https://discord.com/api/webhooks/123456789/abcdefghijklmnop",
	}
	for _, wh := range webhooks {
		t.Run(wh, func(t *testing.T) {
			if err := validateDiscordWebhook(wh); err != nil {
				t.Errorf("webhook %q should be valid, got %v", wh, err)
			}
		})
	}
}

func TestCreate_InvalidWatchType(t *testing.T) {
	webhook := "https://discord.com/api/webhooks/1234/token"
	_, err := svc.Create(context.Background(), webhook, "channel", "something")
	if err == nil || errors.Is(err, ErrInvalidWebhook) {
		t.Errorf("expected watch_type error, got %v", err)
	}
}

func TestCreate_EmptyWatchTarget(t *testing.T) {
	webhook := "https://discord.com/api/webhooks/1234/token"
	_, err := svc.Create(context.Background(), webhook, models.WatchTypeGame, "   ")
	if err == nil || errors.Is(err, ErrInvalidWebhook) {
		t.Errorf("expected watch_target error, got %v", err)
	}
}
