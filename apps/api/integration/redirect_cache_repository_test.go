//go:build integration

package integration_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	urlcache "github.com/tapadar13/url-shortener/apps/api/internal/url/cache"
	rediscache "github.com/tapadar13/url-shortener/apps/api/internal/url/cache/redis"
)

func TestRedirectCacheRepositoryLifecycle(t *testing.T) {
	now := time.Now()
	shortCode := fmt.Sprintf("Cache%d", now.UnixNano())
	keyPrefix := fmt.Sprintf("url-shortener-integration:%d", now.UnixNano())
	ctx, cancel := context.WithTimeout(context.Background(), integrationTimeout)
	defer cancel()

	client := newIntegrationRedisClient(t, keyPrefix)

	repository := rediscache.New(client.Driver(), client.KeyPrefix())
	entry := urlcache.Entry{LongURL: "https://example.com/cached-destination"}
	const ttl = 250 * time.Millisecond

	if err := repository.Set(ctx, shortCode, entry, ttl); err != nil {
		t.Fatalf("set redirect cache entry: %v", err)
	}

	found, err := repository.Get(ctx, shortCode)
	if err != nil {
		t.Fatalf("get redirect cache entry: %v", err)
	}

	if found != entry {
		t.Fatalf("expected cached entry %+v, got %+v", entry, found)
	}

	key := fmt.Sprintf("%s:redirect:%s", keyPrefix, shortCode)
	remaining, err := client.Driver().PTTL(ctx, key).Result()
	if err != nil {
		t.Fatalf("read redirect cache TTL: %v", err)
	}

	if remaining <= 0 || remaining > ttl {
		t.Fatalf("expected positive TTL no greater than %s, got %s", ttl, remaining)
	}

	if err := repository.Delete(ctx, shortCode); err != nil {
		t.Fatalf("delete redirect cache entry: %v", err)
	}

	if _, err := repository.Get(ctx, shortCode); !errors.Is(err, urlcache.ErrMiss) {
		t.Fatalf("expected deleted entry to miss, got %v", err)
	}

	if err := repository.Set(ctx, shortCode, entry, ttl); err != nil {
		t.Fatalf("set expiring redirect cache entry: %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for {
		_, err := repository.Get(ctx, shortCode)
		if errors.Is(err, urlcache.ErrMiss) {
			break
		}
		if err != nil {
			t.Fatalf("get expiring redirect cache entry: %v", err)
		}
		if time.Now().After(deadline) {
			t.Fatal("timed out waiting for redirect cache entry to expire")
		}

		time.Sleep(20 * time.Millisecond)
	}
}
