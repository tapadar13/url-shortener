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
	redisplatform "github.com/tapadar13/url-shortener/apps/api/internal/platform/redis"
	"github.com/tapadar13/url-shortener/apps/api/internal/ratelimit"
	ratelimitrepository "github.com/tapadar13/url-shortener/apps/api/internal/ratelimit/repository/mongodb"
	"github.com/tapadar13/url-shortener/apps/api/internal/transport/httpapi"
	rediscache "github.com/tapadar13/url-shortener/apps/api/internal/url/cache/redis"
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

	logger.Info("MongoDB indexes ready",
		"urls_collection", cfg.MongoDB.URLsCollection,
		"rate_limits_collection", cfg.MongoDB.RateLimitsCollection,
	)

	urlRepository := urlrepository.New(mongoClient.URLsCollection())
	rateLimitRepository := ratelimitrepository.New(mongoClient.RateLimitsCollection())

	updateOptions := service.UpdateOptions{}
	deleteOptions := service.DeleteOptions{}
	redirectOptions := service.RedirectOptions{}

	if cfg.RedirectCache.Enabled {
		redisClient, err := redisplatform.Connect(ctx, cfg.Redis)
		if err != nil {
			return fmt.Errorf("connect Redis redirect cache: %w", err)
		}
		defer func() {
			if err := redisClient.Close(); err != nil {
				logger.Error("Redis disconnect failed", "error", err)
				return
			}

			logger.Info("Redis disconnected")
		}()

		logger.Info("Redis redirect cache connected")

		redirectCache := rediscache.New(redisClient.Driver(), redisClient.KeyPrefix())
		accessRecorder, err := service.NewAsyncAccessRecorder(urlRepository, service.AccessRecorderOptions{
			Workers:   cfg.RedirectCache.AccessWorkers,
			QueueSize: cfg.RedirectCache.AccessQueueSize,
			Timeout:   cfg.RedirectCache.AccessTimeout,
			OnError: func(err error) {
				logger.Error("queued URL access recording failed", "error", err)
			},
		})
		if err != nil {
			return fmt.Errorf("create cached access recorder: %w", err)
		}
		defer func() {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
			defer cancel()

			if err := accessRecorder.Close(shutdownCtx); err != nil {
				logger.Error("cached access recorder shutdown failed", "error", err)
				return
			}

			logger.Info("cached access recorder stopped")
		}()

		reportCacheError := func(err error) {
			logger.Warn("redirect cache operation failed", "error", err)
		}

		updateOptions.Cache = redirectCache
		updateOptions.OnCacheError = reportCacheError
		deleteOptions.Cache = redirectCache
		deleteOptions.OnCacheError = reportCacheError
		redirectOptions.Cache = redirectCache
		redirectOptions.CacheTTL = cfg.RedirectCache.TTL
		redirectOptions.Recorder = accessRecorder
		redirectOptions.OnCacheError = reportCacheError
	}

	logger.Info("redirect cache configured",
		"enabled", cfg.RedirectCache.Enabled,
		"ttl", cfg.RedirectCache.TTL,
		"access_workers", cfg.RedirectCache.AccessWorkers,
		"access_queue_size", cfg.RedirectCache.AccessQueueSize,
		"access_timeout", cfg.RedirectCache.AccessTimeout,
	)

	requestLimiter, err := ratelimit.New(rateLimitRepository, ratelimit.Options{
		Requests: cfg.RateLimit.Requests,
		Window:   cfg.RateLimit.Window,
	})
	if err != nil {
		return fmt.Errorf("create request rate limiter: %w", err)
	}

	logger.Info("request rate limiting configured",
		"requests", cfg.RateLimit.Requests,
		"window", cfg.RateLimit.Window,
		"enabled", cfg.RateLimit.Requests > 0,
		"trusted_proxy_cidrs", len(cfg.HTTP.TrustedProxyCIDRs),
	)

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

	urlUpdater, err := service.NewUpdateService(urlRepository, updateOptions)
	if err != nil {
		return fmt.Errorf("create URL update service: %w", err)
	}

	urlDeleter, err := service.NewDeleteService(urlRepository, deleteOptions)
	if err != nil {
		return fmt.Errorf("create URL delete service: %w", err)
	}

	urlRedirector, err := service.NewRedirectService(urlRepository, redirectOptions)
	if err != nil {
		return fmt.Errorf("create URL redirect service: %w", err)
	}

	handler := httpapi.NewRouter(httpapi.Dependencies{
		ReadinessChecker:   mongoClient,
		URLCreator:         urlCreator,
		URLFinder:          urlFinder,
		URLUpdater:         urlUpdater,
		URLDeleter:         urlDeleter,
		URLRedirector:      urlRedirector,
		RedirectStatusCode: cfg.Redirect.StatusCode,
	})
	handler = httpapi.RateLimit(requestLimiter, cfg.HTTP.TrustedProxyCIDRs...)(handler)
	handler = httpserver.CORS(cfg.HTTP.AllowedOrigins)(handler)
	handler = httpserver.SecurityHeaders(handler)
	handler = httpserver.Timeout(cfg.RequestTimeout)(handler)
	handler = httpapi.Recovery(logger)(handler)
	handler = httpserver.RequestLogger(logger)(handler)
	handler = httpserver.RequestID(handler)

	server := httpserver.New(cfg, handler)

	logger.Info("api server starting", "addr", server.Addr)

	return httpserver.Serve(ctx, server, cfg.ShutdownTimeout, logger)
}
