package poller

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/khiemnguyen15/twitch-watcher/pkg/models"
	"github.com/khiemnguyen15/twitch-watcher/services/stream-poller/internal/publisher"
	"github.com/khiemnguyen15/twitch-watcher/services/stream-poller/internal/subscription"
	"github.com/khiemnguyen15/twitch-watcher/services/stream-poller/internal/twitch"
)

const gameCacheTTL = 24 * time.Hour
const gameIDCachePrefix = "game:"

// Poller is the core poll loop.
type Poller struct {
	subClient   *subscription.Client
	twitchClient *twitch.Client
	publisher   *publisher.Publisher
	cache       *redis.Client
	logger      *slog.Logger
}

// New creates a Poller.
func New(
	subClient *subscription.Client,
	twitchClient *twitch.Client,
	pub *publisher.Publisher,
	cache *redis.Client,
	logger *slog.Logger,
) *Poller {
	return &Poller{
		subClient:    subClient,
		twitchClient: twitchClient,
		publisher:    pub,
		cache:        cache,
		logger:       logger,
	}
}

// Run starts the polling loop and blocks until ctx is cancelled.
func (p *Poller) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Poll immediately on start.
	p.poll(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.poll(ctx)
		}
	}
}

func (p *Poller) poll(ctx context.Context) {
	p.logger.Info("poll cycle started")

	subs, err := p.subClient.ListActive(ctx)
	if err != nil {
		p.logger.Error("fetch subscriptions failed", "error", err)
		return
	}
	if len(subs) == 0 {
		p.logger.Info("no active subscriptions")
		return
	}

	// Build dedup maps.
	gameMap := make(map[string][]models.SubscriptionRef)    // gameName → refs
	streamerMap := make(map[string][]models.SubscriptionRef) // userLogin → refs

	for _, s := range subs {
		ref := models.SubscriptionRef{
			SubscriptionID: s.ID,
			DiscordWebhook: s.DiscordWebhook,
		}
		switch s.WatchType {
		case models.WatchTypeGame:
			gameMap[s.WatchTarget] = append(gameMap[s.WatchTarget], ref)
		case models.WatchTypeStreamer:
			streamerMap[s.WatchTarget] = append(streamerMap[s.WatchTarget], ref)
		}
	}

	// Resolve game names → game IDs (Valkey-cached).
	gameNames := make([]string, 0, len(gameMap))
	for name := range gameMap {
		gameNames = append(gameNames, name)
	}

	gameIDMap, err := p.resolveGameIDs(ctx, gameNames)
	if err != nil {
		p.logger.Error("resolve game IDs failed", "error", err)
		return
	}

	// Collect unique game IDs and user logins for batch query.
	gameIDs := make([]string, 0, len(gameIDMap))
	gameIDToName := make(map[string]string)
	for name, id := range gameIDMap {
		gameIDs = append(gameIDs, id)
		gameIDToName[id] = name
	}

	userLogins := make([]string, 0, len(streamerMap))
	for login := range streamerMap {
		userLogins = append(userLogins, login)
	}

	// Batch fetch live streams.
	streams, err := p.twitchClient.GetStreams(ctx, gameIDs, userLogins)
	if err != nil {
		p.logger.Error("fetch streams failed", "error", err)
		return
	}

	polledAt := time.Now().UTC()
	published := 0

	for _, s := range streams {
		refs := p.collectRefs(s, gameMap, streamerMap)
		if len(refs) == 0 {
			continue
		}

		event := models.StreamEvent{
			StreamID:      s.ID,
			UserLogin:     s.UserLogin,
			UserName:      s.UserName,
			GameID:        s.GameID,
			GameName:      s.GameName,
			Title:         s.Title,
			ViewerCount:   s.ViewerCount,
			StartedAt:     s.StartedAt,
			ThumbnailURL:  formatThumbnail(s.ThumbnailURL, 440, 248),
			StreamURL:     "https://twitch.tv/" + s.UserLogin,
			Subscriptions: refs,
			PolledAt:      polledAt,
		}

		if err := p.publisher.Publish(ctx, event); err != nil {
			p.logger.Error("publish stream event failed", "stream_id", s.ID, "error", err)
			continue
		}
		published++
	}

	p.logger.Info("poll cycle complete", "streams", len(streams), "published", published)
}

// collectRefs gathers all SubscriptionRefs for a stream and deduplicates by DiscordWebhook.
func (p *Poller) collectRefs(
	s models.TwitchStream,
	gameMap map[string][]models.SubscriptionRef,
	streamerMap map[string][]models.SubscriptionRef,
) []models.SubscriptionRef {
	seen := make(map[string]struct{})
	var refs []models.SubscriptionRef

	for _, ref := range gameMap[s.GameName] {
		if _, ok := seen[ref.DiscordWebhook]; !ok {
			seen[ref.DiscordWebhook] = struct{}{}
			refs = append(refs, ref)
		}
	}
	for _, ref := range streamerMap[s.UserLogin] {
		if _, ok := seen[ref.DiscordWebhook]; !ok {
			seen[ref.DiscordWebhook] = struct{}{}
			refs = append(refs, ref)
		}
	}
	return refs
}

// resolveGameIDs resolves game names to IDs, using Valkey as a 24h cache.
func (p *Poller) resolveGameIDs(ctx context.Context, names []string) (map[string]string, error) {
	result := make(map[string]string, len(names))
	var toFetch []string

	for _, name := range names {
		id, err := p.cache.Get(ctx, gameIDCachePrefix+name).Result()
		if err == nil {
			result[name] = id
		} else {
			toFetch = append(toFetch, name)
		}
	}

	if len(toFetch) == 0 {
		return result, nil
	}

	fetched, err := p.twitchClient.GetGameIDs(ctx, toFetch)
	if err != nil {
		return nil, fmt.Errorf("GetGameIDs: %w", err)
	}

	for name, id := range fetched {
		result[name] = id
		p.cache.Set(ctx, gameIDCachePrefix+name, id, gameCacheTTL)
	}

	return result, nil
}

// formatThumbnail replaces Twitch thumbnail URL template placeholders.
func formatThumbnail(tmpl string, w, h int) string {
	if tmpl == "" {
		return ""
	}
	out := tmpl
	for _, from := range []string{"{width}", "%7Bwidth%7D"} {
		out = replaceAll(out, from, fmt.Sprintf("%d", w))
	}
	for _, from := range []string{"{height}", "%7Bheight%7D"} {
		out = replaceAll(out, from, fmt.Sprintf("%d", h))
	}
	return out
}

func replaceAll(s, old, new string) string {
	result := []byte{}
	for i := 0; i < len(s); {
		if i+len(old) <= len(s) && s[i:i+len(old)] == old {
			result = append(result, new...)
			i += len(old)
		} else {
			result = append(result, s[i])
			i++
		}
	}
	return string(result)
}
