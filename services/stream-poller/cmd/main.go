package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"github.com/khiemnguyen15/twitch-watcher/services/stream-poller/internal/config"
	"github.com/khiemnguyen15/twitch-watcher/services/stream-poller/internal/poller"
	"github.com/khiemnguyen15/twitch-watcher/services/stream-poller/internal/publisher"
	"github.com/khiemnguyen15/twitch-watcher/services/stream-poller/internal/subscription"
	"github.com/khiemnguyen15/twitch-watcher/services/stream-poller/internal/twitch"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	nc, err := nats.Connect(cfg.NATSUrl)
	if err != nil {
		logger.Error("NATS connect failed", "error", err)
		os.Exit(1)
	}
	defer nc.Drain()

	pub, err := publisher.New(nc)
	if err != nil {
		logger.Error("create publisher failed", "error", err)
		os.Exit(1)
	}

	rdb := redis.NewClient(&redis.Options{Addr: cfg.ValkeyAddr})
	defer rdb.Close()

	tokenMgr := twitch.NewTokenManager(cfg.TwitchClientID, cfg.TwitchClientSecret)
	twitchClient := twitch.NewClient(cfg.TwitchClientID, tokenMgr)
	subClient := subscription.New(cfg.SubscriptionSvcURL, cfg.InternalAPIKey)

	p := poller.New(subClient, twitchClient, pub, rdb, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		logger.Info("shutting down stream-poller")
		cancel()
	}()

	logger.Info("stream-poller started", "interval", cfg.PollInterval)
	p.Run(ctx, cfg.PollInterval)
	logger.Info("stream-poller stopped")
}
