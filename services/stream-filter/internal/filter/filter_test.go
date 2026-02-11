package filter

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/khiemnguyen15/twitch-watcher/pkg/models"
)

// mockPublisher records published payloads and can be told to return an error.
type mockPublisher struct {
	published []models.NotificationPayload
	err       error
}

func (m *mockPublisher) Publish(_ context.Context, p models.NotificationPayload) error {
	if m.err != nil {
		return m.err
	}
	m.published = append(m.published, p)
	return nil
}

func newTestFilter(t *testing.T, pub notificationPublisher) *Filter {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return New(rdb, pub)
}

func testEvent(streamID, userLogin string, refs ...models.SubscriptionRef) models.StreamEvent {
	return models.StreamEvent{
		StreamID:    streamID,
		UserLogin:   userLogin,
		UserName:    "TestUser",
		GameName:    "Fortnite",
		Title:       "Test stream",
		ViewerCount: 100,
		StartedAt:   time.Now(),
		ThumbnailURL: "https://example.com/thumb.jpg",
		StreamURL:   "https://twitch.tv/" + userLogin,
		Subscriptions: refs,
	}
}

func TestProcess_NewStream_PublishesOnePerRef(t *testing.T) {
	pub := &mockPublisher{}
	f := newTestFilter(t, pub)

	event := testEvent("stream-1", "streamer1",
		models.SubscriptionRef{SubscriptionID: "sub-1", DiscordWebhook: "https://discord.com/api/webhooks/1/a"},
		models.SubscriptionRef{SubscriptionID: "sub-2", DiscordWebhook: "https://discord.com/api/webhooks/2/b"},
	)

	n, err := f.Process(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 2 {
		t.Errorf("published count = %d, want 2", n)
	}
	if len(pub.published) != 2 {
		t.Errorf("mock received %d payloads, want 2", len(pub.published))
	}
}

func TestProcess_DuplicateStream_Discarded(t *testing.T) {
	pub := &mockPublisher{}
	f := newTestFilter(t, pub)

	event := testEvent("stream-1", "streamer1",
		models.SubscriptionRef{SubscriptionID: "sub-1", DiscordWebhook: "https://discord.com/api/webhooks/1/a"},
	)

	// First call: new stream — should publish.
	n1, err := f.Process(context.Background(), event)
	if err != nil {
		t.Fatalf("first call error: %v", err)
	}
	if n1 != 1 {
		t.Errorf("first call: published = %d, want 1", n1)
	}

	// Second call: same stream — should be discarded.
	n2, err := f.Process(context.Background(), event)
	if err != nil {
		t.Fatalf("second call error: %v", err)
	}
	if n2 != 0 {
		t.Errorf("second call: published = %d, want 0 (duplicate)", n2)
	}
	if len(pub.published) != 1 {
		t.Errorf("total payloads published = %d, want 1", len(pub.published))
	}
}

func TestProcess_DifferentStreams_BothPublished(t *testing.T) {
	pub := &mockPublisher{}
	f := newTestFilter(t, pub)

	ref := models.SubscriptionRef{SubscriptionID: "sub-1", DiscordWebhook: "https://discord.com/api/webhooks/1/a"}

	_, err := f.Process(context.Background(), testEvent("stream-1", "streamer1", ref))
	if err != nil {
		t.Fatalf("first error: %v", err)
	}
	_, err = f.Process(context.Background(), testEvent("stream-2", "streamer2", ref))
	if err != nil {
		t.Fatalf("second error: %v", err)
	}

	if len(pub.published) != 2 {
		t.Errorf("total published = %d, want 2", len(pub.published))
	}
}

func TestProcess_PayloadFieldsMatchEvent(t *testing.T) {
	pub := &mockPublisher{}
	f := newTestFilter(t, pub)

	ref := models.SubscriptionRef{SubscriptionID: "sub-42", DiscordWebhook: "https://discord.com/api/webhooks/9/z"}
	event := testEvent("stream-99", "mycaster", ref)

	if _, err := f.Process(context.Background(), event); err != nil {
		t.Fatalf("error: %v", err)
	}

	if len(pub.published) != 1 {
		t.Fatalf("expected 1 payload, got %d", len(pub.published))
	}
	p := pub.published[0]
	if p.SubscriptionID != "sub-42" {
		t.Errorf("SubscriptionID = %q, want sub-42", p.SubscriptionID)
	}
	if p.StreamID != "stream-99" {
		t.Errorf("StreamID = %q, want stream-99", p.StreamID)
	}
	if p.UserLogin != "mycaster" {
		t.Errorf("UserLogin = %q, want mycaster", p.UserLogin)
	}
	if p.StreamURL != event.StreamURL {
		t.Errorf("StreamURL = %q, want %q", p.StreamURL, event.StreamURL)
	}
}

func TestProcess_PublisherError_ReturnsError(t *testing.T) {
	pub := &mockPublisher{err: errors.New("nats unavailable")}
	f := newTestFilter(t, pub)

	event := testEvent("stream-1", "streamer1",
		models.SubscriptionRef{SubscriptionID: "sub-1", DiscordWebhook: "https://discord.com/api/webhooks/1/a"},
	)

	_, err := f.Process(context.Background(), event)
	if err == nil {
		t.Error("expected error when publisher fails, got nil")
	}
}

func TestProcess_KeyFormat(t *testing.T) {
	// Verify the Valkey key encodes both stream ID and user login so that
	// the same stream ID on a different login is treated as a distinct stream.
	pub := &mockPublisher{}
	f := newTestFilter(t, pub)

	ref := models.SubscriptionRef{SubscriptionID: "sub-1", DiscordWebhook: "https://discord.com/api/webhooks/1/a"}

	// Same stream ID, different user login — should both publish.
	_, _ = f.Process(context.Background(), testEvent("stream-1", "streamer-a", ref))
	_, _ = f.Process(context.Background(), testEvent("stream-1", "streamer-b", ref))

	if len(pub.published) != 2 {
		t.Errorf("expected 2 published (different logins), got %d", len(pub.published))
	}
}
