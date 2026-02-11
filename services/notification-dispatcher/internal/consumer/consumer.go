package consumer

import (
	"context"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/khiemnguyen15/twitch-watcher/pkg/messaging"
	"github.com/khiemnguyen15/twitch-watcher/pkg/models"
	"github.com/khiemnguyen15/twitch-watcher/services/notification-dispatcher/internal/discord"
)

// Consumer pulls NotificationPayloads from twitch.streams.new and dispatches Discord webhooks.
type Consumer struct {
	js     jetstream.JetStream
	sender *discord.Sender
	logger *slog.Logger
}

// New creates a Consumer, ensuring the durable consumer exists on TWITCH_STREAMS_NEW.
func New(js jetstream.JetStream, sender *discord.Sender, logger *slog.Logger) (*Consumer, error) {
	_, err := js.CreateOrUpdateConsumer(context.Background(), messaging.StreamTwitchStreamsNew, jetstream.ConsumerConfig{
		Durable:       messaging.ConsumerNotificationDispatcher,
		AckPolicy:     jetstream.AckExplicitPolicy,
		MaxDeliver:    3,
		AckWait:       30 * time.Second,
		FilterSubject: messaging.SubjectStreamsNew,
	})
	if err != nil {
		return nil, err
	}
	return &Consumer{js: js, sender: sender, logger: logger}, nil
}

// Run starts consuming messages until ctx is cancelled.
func (c *Consumer) Run(ctx context.Context) error {
	cons, err := c.js.Consumer(ctx, messaging.StreamTwitchStreamsNew, messaging.ConsumerNotificationDispatcher)
	if err != nil {
		return err
	}

	cc, err := cons.Consume(func(msg jetstream.Msg) {
		env, err := messaging.Unmarshal[models.NotificationPayload](msg.Data())
		if err != nil {
			c.logger.Error("unmarshal notification payload failed", "error", err)
			msg.Nak()
			return
		}

		if err := c.sender.Send(ctx, env.Payload); err != nil {
			c.logger.Error("discord send failed",
				"subscription_id", env.Payload.SubscriptionID,
				"stream_id", env.Payload.StreamID,
				"error", err,
			)
			msg.Nak()
			return
		}

		c.logger.Info("notification dispatched",
			"subscription_id", env.Payload.SubscriptionID,
			"stream_id", env.Payload.StreamID,
			"user_login", env.Payload.UserLogin,
		)
		msg.Ack()
	})
	if err != nil {
		return err
	}
	defer cc.Stop()

	<-ctx.Done()
	return nil
}
