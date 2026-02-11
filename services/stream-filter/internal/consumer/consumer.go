package consumer

import (
	"context"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/khiemnguyen15/twitch-watcher/pkg/messaging"
	"github.com/khiemnguyen15/twitch-watcher/pkg/models"
	"github.com/khiemnguyen15/twitch-watcher/services/stream-filter/internal/filter"
)

// Consumer pulls from twitch.streams.raw and delegates to Filter.
type Consumer struct {
	js     jetstream.JetStream
	filter *filter.Filter
	logger *slog.Logger
}

// New creates a Consumer, ensuring the durable consumer exists.
func New(js jetstream.JetStream, f *filter.Filter, logger *slog.Logger) (*Consumer, error) {
	_, err := js.CreateOrUpdateConsumer(context.Background(), messaging.StreamTwitchStreamsRaw, jetstream.ConsumerConfig{
		Durable:       messaging.ConsumerStreamFilter,
		AckPolicy:     jetstream.AckExplicitPolicy,
		MaxDeliver:    5,
		AckWait:       30 * time.Second,
		FilterSubject: messaging.SubjectStreamsRaw,
	})
	if err != nil {
		return nil, err
	}
	return &Consumer{js: js, filter: f, logger: logger}, nil
}

// Run starts consuming messages until ctx is cancelled.
func (c *Consumer) Run(ctx context.Context) error {
	cons, err := c.js.Consumer(ctx, messaging.StreamTwitchStreamsRaw, messaging.ConsumerStreamFilter)
	if err != nil {
		return err
	}

	cc, err := cons.Consume(func(msg jetstream.Msg) {
		env, err := messaging.Unmarshal[models.StreamEvent](msg.Data())
		if err != nil {
			c.logger.Error("unmarshal stream event failed", "error", err)
			msg.Nak()
			return
		}

		published, err := c.filter.Process(ctx, env.Payload)
		if err != nil {
			c.logger.Error("filter process failed", "stream_id", env.Payload.StreamID, "error", err)
			msg.Nak()
			return
		}

		c.logger.Info("stream processed", "stream_id", env.Payload.StreamID, "notifications_published", published)
		msg.Ack()
	})
	if err != nil {
		return err
	}
	defer cc.Stop()

	<-ctx.Done()
	return nil
}
