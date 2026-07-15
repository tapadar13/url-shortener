package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/tapadar13/url-shortener/apps/api/internal/analytics"
	analyticsrepository "github.com/tapadar13/url-shortener/apps/api/internal/analytics/repository/mongodb"
	"github.com/tapadar13/url-shortener/apps/api/internal/auth"
	authrepository "github.com/tapadar13/url-shortener/apps/api/internal/auth/repository/mongodb"
	"github.com/tapadar13/url-shortener/apps/api/internal/config"
	"github.com/tapadar13/url-shortener/apps/api/internal/metrics"
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
		"users_collection", cfg.MongoDB.UsersCollection,
		"rate_limits_collection", cfg.MongoDB.RateLimitsCollection,
		"analytics_collection", cfg.MongoDB.AnalyticsCollection,
	)

	urlRepository := urlrepository.New(mongoClient.URLsCollection())
	urlListService, err := service.NewListService(urlRepository)
	if err != nil {
		return fmt.Errorf("create URL list service: %w", err)
	}
	authRepository := authrepository.New(mongoClient.UsersCollection())
	authService, err := auth.NewService(authRepository)
	if err != nil {
		return fmt.Errorf("create authentication service: %w", err)
	}
	tokenService, err := auth.NewTokenService(auth.TokenOptions{
		Secret:   cfg.Auth.TokenSecret,
		Issuer:   cfg.Auth.TokenIssuer,
		Audience: cfg.Auth.TokenAudience,
		TTL:      cfg.Auth.TokenTTL,
	})
	if err != nil {
		return fmt.Errorf("create authentication token service: %w", err)
	}
	rateLimitRepository := ratelimitrepository.New(mongoClient.RateLimitsCollection())
	analyticsRepository := analyticsrepository.New(mongoClient.AnalyticsCollection())
	analyticsReporter, err := analytics.NewReporter(analyticsRepository, analytics.ReporterOptions{})
	if err != nil {
		return fmt.Errorf("create click analytics reporter: %w", err)
	}

	analyticsRecorder, err := analytics.NewAsyncRecorder(analyticsRepository, analytics.AsyncRecorderOptions{
		Workers:   cfg.Analytics.Workers,
		QueueSize: cfg.Analytics.QueueSize,
		Timeout:   cfg.Analytics.WriteTimeout,
		OnError: func(err error) {
			logger.Error("queued click analytics recording failed", "error", err)
		},
	})
	if err != nil {
		return fmt.Errorf("create click analytics recorder: %w", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()

		if err := analyticsRecorder.Close(shutdownCtx); err != nil {
			logger.Error("click analytics recorder shutdown failed", "error", err)
			return
		}

		logger.Info("click analytics recorder stopped")
	}()

	logger.Info("click analytics configured",
		"workers", cfg.Analytics.Workers,
		"queue_size", cfg.Analytics.QueueSize,
		"write_timeout", cfg.Analytics.WriteTimeout,
		"max_report_range_days", analytics.DefaultMaxRangeDays,
	)

	updateOptions := service.UpdateOptions{}
	deleteOptions := service.DeleteOptions{}
	redirectOptions := service.RedirectOptions{
		Analytics: analyticsRecorder,
		OnAnalyticsError: func(err error) {
			logger.Warn("click analytics enqueue failed", "error", err)
		},
	}

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
		ReadinessChecker:    mongoClient,
		URLCreator:          urlCreator,
		URLLister:           urlListService,
		URLFinder:           urlFinder,
		URLUpdater:          urlUpdater,
		URLDeleter:          urlDeleter,
		URLRedirector:       urlRedirector,
		AnalyticsReporter:   analyticsReporter,
		RedirectStatusCode:  cfg.Redirect.StatusCode,
		Metrics:             metrics.New(),
		AuthService:         authService,
		AccessTokenIssuer:   tokenService,
		AccessTokenVerifier: tokenService,
	})
	handler = httpapi.RateLimit(requestLimiter, cfg.HTTP.TrustedProxyCIDRs...)(handler)
	handler = httpserver.CORS(cfg.HTTP.AllowedOrigins)(handler)
	handler = httpserver.SecurityHeaders(handler)
	handler = httpserver.Timeout(cfg.RequestTimeout)(handler)
	handler = httpserver.MaxRequestBody(cfg.MaxRequestBodyBytes)(handler)
	handler = httpapi.Recovery(logger)(handler)
	handler = httpserver.RequestLogger(logger)(handler)
	handler = httpserver.RequestID(handler)

	server := httpserver.New(cfg, handler)

	logger.Info("api server starting", "addr", server.Addr)

	return httpserver.Serve(ctx, server, cfg.ShutdownTimeout, logger)
}
