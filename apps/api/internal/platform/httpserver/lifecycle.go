package httpserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

func Serve(ctx context.Context, server *http.Server, shutdownTimeout time.Duration, logger *slog.Logger) error {
	if server == nil {
		return errors.New("server is required")
	}

	return serve(ctx, shutdownTimeout, logger, server.ListenAndServe, server.Shutdown)
}

func serve(
	ctx context.Context,
	shutdownTimeout time.Duration,
	logger *slog.Logger,
	listenAndServe func() error,
	shutdown func(context.Context) error,
) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if logger == nil {
		logger = slog.Default()
	}

	serverErr := make(chan error, 1)
	go func() {
		err := listenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			serverErr <- nil
			return
		}

		serverErr <- err
	}()

	select {
	case err := <-serverErr:
		if err != nil {
			return fmt.Errorf("listen and serve: %w", err)
		}

		return nil
	case <-ctx.Done():
		logger.Info("api server shutting down")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown server: %w", err)
		}

		if err := <-serverErr; err != nil {
			return fmt.Errorf("listen and serve: %w", err)
		}

		logger.Info("api server stopped")

		return nil
	}
}
