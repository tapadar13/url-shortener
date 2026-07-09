package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/tapadar13/url-shortener/apps/api/internal/config"
	"github.com/tapadar13/url-shortener/apps/api/internal/platform/httpserver"
	"github.com/tapadar13/url-shortener/apps/api/internal/platform/logging"
	"github.com/tapadar13/url-shortener/apps/api/internal/platform/mongodb"
	"github.com/tapadar13/url-shortener/apps/api/internal/transport/httpapi"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "startup failed: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger := logging.New(cfg.Log, os.Stdout)
	slog.SetDefault(logger)

	logger.Info("api service configured", "environment", cfg.Environment)

	mongoClient, err := mongodb.Connect(ctx, cfg.MongoDB, cfg.RequestTimeout)
	if err != nil {
		return fmt.Errorf("connect MongoDB: %w", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()

		if err := mongoClient.Disconnect(shutdownCtx); err != nil {
			logger.Error("MongoDB disconnect failed", "error", err)
			return
		}

		logger.Info("MongoDB disconnected")
	}()

	logger.Info("MongoDB connected", "database", cfg.MongoDB.Database)

	if err := mongodb.EnsureIndexes(ctx, mongoClient); err != nil {
		return fmt.Errorf("ensure MongoDB indexes: %w", err)
	}

	logger.Info("MongoDB indexes ready", "collection", cfg.MongoDB.URLsCollection)

	server := httpserver.New(cfg, httpapi.NewRouter())

	logger.Info("api server starting", "addr", server.Addr)

	return httpserver.Serve(ctx, server, cfg.ShutdownTimeout, logger)
}
