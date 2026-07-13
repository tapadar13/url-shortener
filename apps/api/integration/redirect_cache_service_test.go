//go:build integration

package integration_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	urlmodel "github.com/tapadar13/url-shortener/apps/api/internal/url"
	urlcache "github.com/tapadar13/url-shortener/apps/api/internal/url/cache"
	rediscache "github.com/tapadar13/url-shortener/apps/api/internal/url/cache/redis"
	"github.com/tapadar13/url-shortener/apps/api/internal/url/service"
)

func TestCachedRedirectServiceLifecycle(t *testing.T) {
	now := time.Now().UTC()
	shortCode := fmt.Sprintf("Flow%d", now.UnixNano())
	keyPrefix := fmt.Sprintf("url-shortener-flow:%d", now.UnixNano())
	ctx, cancel := context.WithTimeout(context.Background(), integrationTimeout)
	defer cancel()

	urlRepository, _ := newIntegrationURLRepository(t)
	redisClient := newIntegrationRedisClient(t, keyPrefix)
	redirectCache := rediscache.New(redisClient.Driver(), redisClient.KeyPrefix())

	recordErrors := make(chan error, 16)
	accessRecorder, err := service.NewAsyncAccessRecorder(urlRepository, service.AccessRecorderOptions{
		Workers:   1,
		QueueSize: 8,
		Timeout:   integrationTimeout,
		OnError: func(err error) {
			recordErrors <- err
		},
	})
	if err != nil {
		t.Fatalf("create access recorder: %v", err)
	}
	t.Cleanup(func() {
		closeCtx, closeCancel := context.WithTimeout(context.Background(), integrationTimeout)
		defer closeCancel()
		if err := accessRecorder.Close(closeCtx); err != nil {
			t.Errorf("close access recorder: %v", err)
		}
	})

	cacheErrors := make(chan error, 16)
	reportCacheError := func(err error) {
		cacheErrors <- err
	}

	redirectService, err := service.NewRedirectService(urlRepository, service.RedirectOptions{
		Cache:        redirectCache,
		CacheTTL:     time.Minute,
		Recorder:     accessRecorder,
		OnCacheError: reportCacheError,
	})
	if err != nil {
		t.Fatalf("create redirect service: %v", err)
	}

	updateService, err := service.NewUpdateService(urlRepository, service.UpdateOptions{
		Cache:        redirectCache,
		OnCacheError: reportCacheError,
	})
	if err != nil {
		t.Fatalf("create update service: %v", err)
	}

	deleteService, err := service.NewDeleteService(urlRepository, service.DeleteOptions{
		Cache:        redirectCache,
		OnCacheError: reportCacheError,
	})
	if err != nil {
		t.Fatalf("create delete service: %v", err)
	}

	record, err := urlmodel.New(urlmodel.NewParams{
		LongURL:   "https://example.com/original",
		ShortCode: shortCode,
		Now:       now,
	})
	if err != nil {
		t.Fatalf("create URL record: %v", err)
	}
	if _, err := urlRepository.Create(ctx, record); err != nil {
		t.Fatalf("persist URL record: %v", err)
	}

	first, err := redirectService.Resolve(ctx, shortCode)
	if err != nil || first.LongURL != record.LongURL {
		t.Fatalf("expected first redirect from MongoDB, record=%+v err=%v", first, err)
	}

	cached, err := redirectCache.Get(ctx, shortCode)
	if err != nil || cached.LongURL != record.LongURL {
		t.Fatalf("expected redirect cache to be populated, entry=%+v err=%v", cached, err)
	}

	second, err := redirectService.Resolve(ctx, shortCode)
	if err != nil || second.LongURL != record.LongURL {
		t.Fatalf("expected second redirect from cache, record=%+v err=%v", second, err)
	}

	waitForAccessCount(t, ctx, urlRepository, shortCode, 2)

	const updatedLongURL = "https://example.com/updated"
	updated, err := updateService.UpdateLongURL(ctx, service.UpdateParams{
		ShortCode: shortCode,
		LongURL:   updatedLongURL,
	})
	if err != nil || updated.LongURL != updatedLongURL {
		t.Fatalf("expected URL update, record=%+v err=%v", updated, err)
	}

	if _, err := redirectCache.Get(ctx, shortCode); !errors.Is(err, urlcache.ErrMiss) {
		t.Fatalf("expected update to invalidate redirect cache, got %v", err)
	}

	refreshed, err := redirectService.Resolve(ctx, shortCode)
	if err != nil || refreshed.LongURL != updatedLongURL {
		t.Fatalf("expected refreshed redirect, record=%+v err=%v", refreshed, err)
	}

	if err := deleteService.DeleteByShortCode(ctx, shortCode); err != nil {
		t.Fatalf("delete URL: %v", err)
	}

	if _, err := redirectCache.Get(ctx, shortCode); !errors.Is(err, urlcache.ErrMiss) {
		t.Fatalf("expected deletion to invalidate redirect cache, got %v", err)
	}

	if _, err := redirectService.Resolve(ctx, shortCode); !errors.Is(err, urlmodel.ErrNotFound) {
		t.Fatalf("expected deleted URL not to resolve, got %v", err)
	}

	closeCtx, closeCancel := context.WithTimeout(context.Background(), integrationTimeout)
	defer closeCancel()
	if err := accessRecorder.Close(closeCtx); err != nil {
		t.Fatalf("close access recorder: %v", err)
	}

	select {
	case err := <-recordErrors:
		t.Fatalf("unexpected queued access error: %v", err)
	default:
	}

	select {
	case err := <-cacheErrors:
		t.Fatalf("unexpected cache operation error: %v", err)
	default:
	}
}

func waitForAccessCount(
	t *testing.T,
	ctx context.Context,
	repository interface {
		FindByShortCode(context.Context, string) (urlmodel.URL, error)
	},
	shortCode string,
	expected int64,
) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for {
		record, err := repository.FindByShortCode(ctx, shortCode)
		if err != nil {
			t.Fatalf("read URL access count: %v", err)
		}
		if record.AccessCount == expected {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for access count %d, got %d", expected, record.AccessCount)
		}

		time.Sleep(20 * time.Millisecond)
	}
}
