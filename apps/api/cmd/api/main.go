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
	urlrepository "github.com/tapadar13/url-shortener/apps/api/internal/url/repository/mongodb"
	"github.com/tapadar13/url-shortener/apps/api/internal/url/service"
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

	urlRepository := urlrepository.New(mongoClient.URLsCollection())

	urlCreator, err := service.New(
		urlRepository,
		service.DefaultGenerator(),
		service.Options{
			ShortCodeLength: cfg.ShortCode.Length,
			MaxRetries:      cfg.ShortCode.MaxRetries,
		},
	)
	if err != nil {
		return fmt.Errorf("create URL service: %w", err)
	}

	urlFinder, err := service.NewLookupService(urlRepository)
	if err != nil {
		return fmt.Errorf("create URL lookup service: %w", err)
	}

	urlUpdater, err := service.NewUpdateService(urlRepository, service.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("create URL update service: %w", err)
	}

	server := httpserver.New(cfg, httpapi.NewRouter(httpapi.Dependencies{
		URLCreator: urlCreator,
		URLFinder:  urlFinder,
		URLUpdater: urlUpdater,
	}))

	logger.Info("api server starting", "addr", server.Addr)

	return httpserver.Serve(ctx, server, cfg.ShutdownTimeout, logger)
}
