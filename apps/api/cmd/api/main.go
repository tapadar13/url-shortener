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

	server := httpserver.New(cfg, httpapi.NewRouter())

	logger.Info("api server starting", "addr", server.Addr)

	return httpserver.Serve(ctx, server, cfg.ShutdownTimeout, logger)
}
