package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/khiemnguyen15/twitch-watcher/services/notification-dispatcher/internal/config"
	"github.com/khiemnguyen15/twitch-watcher/services/notification-dispatcher/internal/consumer"
	"github.com/khiemnguyen15/twitch-watcher/services/notification-dispatcher/internal/discord"
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

	js, err := jetstream.New(nc)
	if err != nil {
		logger.Error("JetStream init failed", "error", err)
		os.Exit(1)
	}

	sender := discord.New()

	cons, err := consumer.New(js, sender, logger)
	if err != nil {
		logger.Error("create consumer failed", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		logger.Info("shutting down notification-dispatcher")
		cancel()
	}()

	logger.Info("notification-dispatcher started")
	if err := cons.Run(ctx); err != nil {
		logger.Error("consumer error", "error", err)
		os.Exit(1)
	}
	logger.Info("notification-dispatcher stopped")
}
