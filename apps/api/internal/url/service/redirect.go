package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/tapadar13/url-shortener/apps/api/internal/analytics"
	"github.com/tapadar13/url-shortener/apps/api/internal/shortcode"
	urlmodel "github.com/tapadar13/url-shortener/apps/api/internal/url"
	urlcache "github.com/tapadar13/url-shortener/apps/api/internal/url/cache"
)

var (
	ErrRedirectCacheRequired  = errors.New("redirect cache is required when an access recorder is configured")
	ErrAccessRecorderRequired = errors.New("access recorder is required when redirect caching is enabled")
)

type AccessRepository interface {
	RecordAccess(ctx context.Context, params urlmodel.RecordAccessParams) (urlmodel.URL, error)
}

type AccessEnqueuer interface {
	Enqueue(shortCode string, accessedAt time.Time) error
}

type RedirectOptions struct {
	Cache            urlcache.Store
	CacheTTL         time.Duration
	Recorder         AccessEnqueuer
	OnCacheError     func(error)
	Analytics        analytics.Enqueuer
	OnAnalyticsError func(error)
	Now              func() time.Time
}

type RedirectService struct {
	repository       AccessRepository
	cache            urlcache.Store
	cacheTTL         time.Duration
	recorder         AccessEnqueuer
	onCacheError     func(error)
	analytics        analytics.Enqueuer
	onAnalyticsError func(error)
	now              func() time.Time
}

func NewRedirectService(repository AccessRepository, options RedirectOptions) (*RedirectService, error) {
	if repository == nil {
		return nil, ErrRepositoryRequired
	}

	if options.Cache == nil && options.Recorder != nil {
		return nil, ErrRedirectCacheRequired
	}

	if options.Cache != nil && options.Recorder == nil {
		return nil, ErrAccessRecorderRequired
	}

	if options.Cache != nil && options.CacheTTL <= 0 {
		return nil, urlcache.ErrTTLInvalid
	}

	if options.Now == nil {
		options.Now = time.Now
	}

	return &RedirectService{
		repository:       repository,
		cache:            options.Cache,
		cacheTTL:         options.CacheTTL,
		recorder:         options.Recorder,
		onCacheError:     options.OnCacheError,
		analytics:        options.Analytics,
		onAnalyticsError: options.OnAnalyticsError,
		now:              options.Now,
	}, nil
}

func (s *RedirectService) Resolve(ctx context.Context, shortCode string) (urlmodel.URL, error) {
	if s == nil || s.repository == nil {
		return urlmodel.URL{}, ErrRepositoryRequired
	}

	if s.now == nil {
		s.now = time.Now
	}

	normalizedShortCode, err := shortcode.Normalize(shortCode)
	if err != nil {
		return urlmodel.URL{}, err
	}

	accessedAt := s.now().UTC()
	if s.cache != nil && s.recorder != nil {
		entry, cacheErr := s.cache.Get(ctx, normalizedShortCode)
		if cacheErr == nil {
			if err := s.recorder.Enqueue(normalizedShortCode, accessedAt); err == nil {
				resolved := urlmodel.URL{
					LongURL:   entry.LongURL,
					ShortCode: normalizedShortCode,
				}
				s.recordAnalytics(normalizedShortCode, accessedAt)
				return resolved, nil
			}

			recorded, err := s.recordAccess(ctx, normalizedShortCode, accessedAt)
			if err != nil {
				return urlmodel.URL{}, err
			}

			s.recordAnalytics(normalizedShortCode, accessedAt)
			return recorded, nil
		}

		if !errors.Is(cacheErr, urlcache.ErrMiss) {
			s.reportCacheError(fmt.Errorf("get redirect cache entry: %w", cacheErr))
		}
	}

	recorded, err := s.recordAccess(ctx, normalizedShortCode, accessedAt)
	if err != nil {
		return urlmodel.URL{}, err
	}

	if s.cache != nil {
		if ttl, ok := redirectCacheTTL(recorded.ExpiresAt, accessedAt, s.cacheTTL); ok {
			if err := s.cache.Set(ctx, normalizedShortCode, urlcache.Entry{LongURL: recorded.LongURL}, ttl); err != nil {
				s.reportCacheError(fmt.Errorf("set redirect cache entry: %w", err))
			}
		}
	}

	s.recordAnalytics(normalizedShortCode, accessedAt)
	return recorded, nil
}

func (s *RedirectService) recordAccess(ctx context.Context, shortCode string, accessedAt time.Time) (urlmodel.URL, error) {
	recorded, err := s.repository.RecordAccess(ctx, urlmodel.RecordAccessParams{
		ShortCode:  shortCode,
		AccessedAt: accessedAt,
	})
	if err != nil {
		return urlmodel.URL{}, fmt.Errorf("record URL access: %w", err)
	}

	return recorded, nil
}

func (s *RedirectService) reportCacheError(err error) {
	if s.onCacheError != nil {
		s.onCacheError(err)
	}
}

func (s *RedirectService) recordAnalytics(shortCode string, clickedAt time.Time) {
	if s.analytics == nil {
		return
	}

	if err := s.analytics.Enqueue(shortCode, clickedAt); err != nil && s.onAnalyticsError != nil {
		s.onAnalyticsError(fmt.Errorf("enqueue click analytics: %w", err))
	}
}

func redirectCacheTTL(expiresAt *time.Time, now time.Time, maximum time.Duration) (time.Duration, bool) {
	if maximum <= 0 {
		return 0, false
	}

	ttl := maximum
	if expiresAt != nil {
		remaining := expiresAt.UTC().Sub(now.UTC())
		if remaining <= 0 {
			return 0, false
		}

		if remaining < ttl {
			ttl = remaining
		}
	}

	return ttl, true
}
