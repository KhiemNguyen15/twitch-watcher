package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/redis/go-redis/v9"
	"github.com/khiemnguyen15/twitch-watcher/services/stream-filter/internal/config"
	"github.com/khiemnguyen15/twitch-watcher/services/stream-filter/internal/consumer"
	"github.com/khiemnguyen15/twitch-watcher/services/stream-filter/internal/filter"
	"github.com/khiemnguyen15/twitch-watcher/services/stream-filter/internal/publisher"
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

	pub, err := publisher.New(nc)
	if err != nil {
		logger.Error("create publisher failed", "error", err)
		os.Exit(1)
	}

	rdb := redis.NewClient(&redis.Options{Addr: cfg.ValkeyAddr})
	defer rdb.Close()

	f := filter.New(rdb, pub)

	cons, err := consumer.New(js, f, logger)
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
		logger.Info("shutting down stream-filter")
		cancel()
	}()

	logger.Info("stream-filter started")
	if err := cons.Run(ctx); err != nil {
		logger.Error("consumer error", "error", err)
		os.Exit(1)
	}
	logger.Info("stream-filter stopped")
}
