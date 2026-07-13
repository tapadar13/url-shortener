package redis

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	redisdriver "github.com/redis/go-redis/v9"
	"github.com/tapadar13/url-shortener/apps/api/internal/shortcode"
	urlmodel "github.com/tapadar13/url-shortener/apps/api/internal/url"
	urlcache "github.com/tapadar13/url-shortener/apps/api/internal/url/cache"
)

func TestRepositoryGetReturnsCachedRedirect(t *testing.T) {
	t.Parallel()

	payload := marshalEntry(t, entryDocument{
		Version: entryVersion,
		LongURL: "https://example.com/articles/123",
	})
	client := &fakeCommandClient{getResult: redisdriver.NewStringResult(payload, nil)}
	repository := newRepository(client, " url-shortener: ")

	entry, err := repository.Get(context.Background(), " AbC123 ")
	if err != nil {
		t.Fatalf("expected cache lookup to succeed: %v", err)
	}

	if entry.LongURL != "https://example.com/articles/123" {
		t.Fatalf("expected cached long URL, got %q", entry.LongURL)
	}

	if client.key != "url-shortener:redirect:AbC123" {
		t.Fatalf("expected namespaced cache key, got %q", client.key)
	}
}

func TestRepositoryGetMapsMissingKeyToCacheMiss(t *testing.T) {
	t.Parallel()

	client := &fakeCommandClient{getResult: redisdriver.NewStringResult("", redisdriver.Nil)}
	repository := newRepository(client, "url-shortener")

	_, err := repository.Get(context.Background(), "AbC123")
	if !errors.Is(err, urlcache.ErrMiss) {
		t.Fatalf("expected cache miss, got %v", err)
	}
}

func TestRepositoryGetRejectsInvalidPayload(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		payload string
	}{
		{name: "malformed JSON", payload: "{"},
		{name: "unsupported version", payload: marshalEntry(t, entryDocument{Version: 2, LongURL: "https://example.com"})},
		{name: "invalid long URL", payload: marshalEntry(t, entryDocument{Version: entryVersion, LongURL: "javascript:alert(1)"})},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := &fakeCommandClient{getResult: redisdriver.NewStringResult(tt.payload, nil)}
			_, err := newRepository(client, "url-shortener").Get(context.Background(), "AbC123")
			if err == nil {
				t.Fatal("expected invalid payload error")
			}
		})
	}
}

func TestRepositoryGetWrapsRedisError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("Redis unavailable")
	client := &fakeCommandClient{getResult: redisdriver.NewStringResult("", expectedErr)}

	_, err := newRepository(client, "url-shortener").Get(context.Background(), "AbC123")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected Redis error, got %v", err)
	}
}

func TestRepositorySetStoresVersionedEntryWithTTL(t *testing.T) {
	t.Parallel()

	client := &fakeCommandClient{setResult: redisdriver.NewStatusResult("OK", nil)}
	repository := newRepository(client, "url-shortener")
	ttl := 5 * time.Minute

	err := repository.Set(context.Background(), " AbC123 ", urlcache.Entry{
		LongURL: " https://example.com/articles/123 ",
	}, ttl)
	if err != nil {
		t.Fatalf("expected cache write to succeed: %v", err)
	}

	if client.key != "url-shortener:redirect:AbC123" || client.expiration != ttl {
		t.Fatalf("expected namespaced key and TTL, got key=%q ttl=%s", client.key, client.expiration)
	}

	payload, ok := client.value.([]byte)
	if !ok {
		t.Fatalf("expected byte payload, got %T", client.value)
	}

	var document entryDocument
	if err := json.Unmarshal(payload, &document); err != nil {
		t.Fatalf("decode stored entry: %v", err)
	}

	if document.Version != entryVersion || document.LongURL != "https://example.com/articles/123" {
		t.Fatalf("expected normalized versioned entry, got %+v", document)
	}
}

func TestRepositorySetValidatesEntryBeforeWriting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		shortCode string
		entry     urlcache.Entry
		ttl       time.Duration
		expected  error
	}{
		{name: "invalid short code", shortCode: "invalid-code", entry: urlcache.Entry{LongURL: "https://example.com"}, ttl: time.Minute, expected: shortcode.ErrInvalidChars},
		{name: "invalid TTL", shortCode: "AbC123", entry: urlcache.Entry{LongURL: "https://example.com"}, expected: urlcache.ErrTTLInvalid},
		{name: "invalid long URL", shortCode: "AbC123", entry: urlcache.Entry{LongURL: "ftp://example.com"}, ttl: time.Minute, expected: urlmodel.ErrLongURLSchemeUnsupported},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := &fakeCommandClient{}
			err := newRepository(client, "url-shortener").Set(context.Background(), tt.shortCode, tt.entry, tt.ttl)
			if !errors.Is(err, tt.expected) {
				t.Fatalf("expected %v, got %v", tt.expected, err)
			}

			if client.setCalled {
				t.Fatal("expected Redis not to be called")
			}
		})
	}
}

func TestRepositorySetWrapsRedisError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("Redis unavailable")
	client := &fakeCommandClient{setResult: redisdriver.NewStatusResult("", expectedErr)}

	err := newRepository(client, "url-shortener").Set(context.Background(), "AbC123", urlcache.Entry{LongURL: "https://example.com"}, time.Minute)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected Redis error, got %v", err)
	}
}

func TestRepositoryDeleteRemovesNamespacedEntry(t *testing.T) {
	t.Parallel()

	client := &fakeCommandClient{deleteResult: redisdriver.NewIntResult(1, nil)}
	err := newRepository(client, "url-shortener").Delete(context.Background(), " AbC123 ")
	if err != nil {
		t.Fatalf("expected cache delete to succeed: %v", err)
	}

	if client.key != "url-shortener:redirect:AbC123" {
		t.Fatalf("expected namespaced cache key, got %q", client.key)
	}
}

func TestRepositoryDeleteWrapsRedisError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("Redis unavailable")
	client := &fakeCommandClient{deleteResult: redisdriver.NewIntResult(0, expectedErr)}

	err := newRepository(client, "url-shortener").Delete(context.Background(), "AbC123")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected Redis error, got %v", err)
	}
}

func TestRepositoryRequiresClientAndKeyPrefix(t *testing.T) {
	t.Parallel()

	_, err := New(nil, "url-shortener").Get(context.Background(), "AbC123")
	if err == nil || !strings.Contains(err.Error(), "client") {
		t.Fatalf("expected client error, got %v", err)
	}

	_, err = newRepository(&fakeCommandClient{}, "::").Get(context.Background(), "AbC123")
	if err == nil || !strings.Contains(err.Error(), "key prefix") {
		t.Fatalf("expected key prefix error, got %v", err)
	}
}

func marshalEntry(t *testing.T, document entryDocument) string {
	t.Helper()

	payload, err := json.Marshal(document)
	if err != nil {
		t.Fatalf("encode test entry: %v", err)
	}

	return string(payload)
}

type fakeCommandClient struct {
	getResult    *redisdriver.StringCmd
	setResult    *redisdriver.StatusCmd
	deleteResult *redisdriver.IntCmd
	key          string
	value        any
	expiration   time.Duration
	setCalled    bool
}

func (c *fakeCommandClient) Get(_ context.Context, key string) *redisdriver.StringCmd {
	c.key = key
	if c.getResult == nil {
		return redisdriver.NewStringResult("", nil)
	}

	return c.getResult
}

func (c *fakeCommandClient) Set(_ context.Context, key string, value any, expiration time.Duration) *redisdriver.StatusCmd {
	c.key = key
	c.value = value
	c.expiration = expiration
	c.setCalled = true
	if c.setResult == nil {
		return redisdriver.NewStatusResult("OK", nil)
	}

	return c.setResult
}

func (c *fakeCommandClient) Del(_ context.Context, keys ...string) *redisdriver.IntCmd {
	if len(keys) > 0 {
		c.key = keys[0]
	}
	if c.deleteResult == nil {
		return redisdriver.NewIntResult(0, nil)
	}

	return c.deleteResult
}
