package filter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/khiemnguyen15/twitch-watcher/pkg/models"
	"github.com/khiemnguyen15/twitch-watcher/services/stream-filter/internal/publisher"
)

const seenTTL = 26 * time.Hour

// Filter deduplicates StreamEvents via Valkey SETNX and fans out NotificationPayloads.
type Filter struct {
	cache     *redis.Client
	publisher *publisher.Publisher
}

// New creates a Filter.
func New(cache *redis.Client, pub *publisher.Publisher) *Filter {
	return &Filter{cache: cache, publisher: pub}
}

// Process checks whether the stream is new (not seen in the last 26h) and
// fans out one NotificationPayload per SubscriptionRef if it is.
func (f *Filter) Process(ctx context.Context, event models.StreamEvent) (int, error) {
	key := fmt.Sprintf("seen:%s:%s", event.StreamID, event.UserLogin)

	set, err := f.cache.SetNX(ctx, key, "1", seenTTL).Result()
	if err != nil {
		return 0, fmt.Errorf("valkey SETNX %s: %w", key, err)
	}
	if !set {
		// Already seen; discard.
		return 0, nil
	}

	published := 0
	for _, ref := range event.Subscriptions {
		payload := models.NotificationPayload{
			SubscriptionID: ref.SubscriptionID,
			DiscordWebhook: ref.DiscordWebhook,
			StreamID:       event.StreamID,
			UserLogin:      event.UserLogin,
			UserName:       event.UserName,
			GameName:       event.GameName,
			Title:          event.Title,
			ViewerCount:    event.ViewerCount,
			StartedAt:      event.StartedAt,
			ThumbnailURL:   event.ThumbnailURL,
			StreamURL:      event.StreamURL,
		}
		if err := f.publisher.Publish(ctx, payload); err != nil {
			return published, fmt.Errorf("publish notification for sub %s: %w", ref.SubscriptionID, err)
		}
		published++
	}

	return published, nil
}
