package publisher

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/khiemnguyen15/twitch-watcher/pkg/messaging"
	"github.com/khiemnguyen15/twitch-watcher/pkg/models"
)

// Publisher publishes NotificationPayloads to NATS JetStream.
type Publisher struct {
	js jetstream.JetStream
}

// New creates a Publisher, ensuring the twitch.streams.new stream exists.
func New(nc *nats.Conn) (*Publisher, error) {
	js, err := jetstream.New(nc)
	if err != nil {
		return nil, fmt.Errorf("create jetstream context: %w", err)
	}

	_, err = js.CreateOrUpdateStream(context.Background(), jetstream.StreamConfig{
		Name:      messaging.StreamTwitchStreamsNew,
		Subjects:  []string{messaging.SubjectStreamsNew},
		Retention: jetstream.WorkQueuePolicy,
		MaxAge:    30 * time.Minute,
		Storage:   jetstream.FileStorage,
		Replicas:  1, // set to 3 via Helm in production
	})
	if err != nil {
		return nil, fmt.Errorf("create stream %s: %w", messaging.StreamTwitchStreamsNew, err)
	}

	return &Publisher{js: js}, nil
}

// Publish serialises and publishes a NotificationPayload envelope.
func (p *Publisher) Publish(ctx context.Context, payload models.NotificationPayload) error {
	env := messaging.NewEnvelope(payload)
	data, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("marshal notification payload: %w", err)
	}

	if _, err := p.js.Publish(ctx, messaging.SubjectStreamsNew, data); err != nil {
		return fmt.Errorf("publish notification payload: %w", err)
	}
	return nil
}
