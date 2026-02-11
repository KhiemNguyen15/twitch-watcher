package publisher

import (
	"context"
	"encoding/json/v2"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/khiemnguyen15/twitch-watcher/pkg/messaging"
	"github.com/khiemnguyen15/twitch-watcher/pkg/models"
)

// Publisher publishes StreamEvents to NATS JetStream.
type Publisher struct {
	js jetstream.JetStream
}

// New creates a Publisher, ensuring the JetStream stream exists.
func New(nc *nats.Conn) (*Publisher, error) {
	js, err := jetstream.New(nc)
	if err != nil {
		return nil, fmt.Errorf("create jetstream context: %w", err)
	}

	_, err = js.CreateOrUpdateStream(context.Background(), jetstream.StreamConfig{
		Name:       messaging.StreamTwitchStreamsRaw,
		Subjects:   []string{messaging.SubjectStreamsRaw},
		Retention:  jetstream.WorkQueuePolicy,
		MaxAge:     5 * 60 * 1000000000, // 5 minutes in nanoseconds
		Storage:    jetstream.FileStorage,
		Replicas:   1, // set to 3 in production via Helm values
	})
	if err != nil {
		return nil, fmt.Errorf("create stream %s: %w", messaging.StreamTwitchStreamsRaw, err)
	}

	return &Publisher{js: js}, nil
}

// Publish serialises and publishes a StreamEvent envelope.
func (p *Publisher) Publish(ctx context.Context, event models.StreamEvent) error {
	env := messaging.NewEnvelope(event)
	data, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("marshal stream event: %w", err)
	}

	if _, err := p.js.Publish(ctx, messaging.SubjectStreamsRaw, data); err != nil {
		return fmt.Errorf("publish stream event: %w", err)
	}
	return nil
}
