package redis

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/tapadar13/url-shortener/apps/api/internal/config"
)

func TestConnectRejectsInvalidTimeout(t *testing.T) {
	t.Parallel()

	_, err := Connect(context.Background(), config.RedisConfig{
		URL:       "redis://localhost:6379/0",
		KeyPrefix: "url-shortener",
	})
	if err == nil || !strings.Contains(err.Error(), "timeout") {
		t.Fatalf("expected timeout error, got %v", err)
	}
}

func TestConnectRejectsMissingKeyPrefix(t *testing.T) {
	t.Parallel()

	_, err := Connect(context.Background(), config.RedisConfig{
		URL:            "redis://localhost:6379/0",
		ConnectTimeout: time.Second,
	})
	if err == nil || !strings.Contains(err.Error(), "key prefix") {
		t.Fatalf("expected key prefix error, got %v", err)
	}
}

func TestConnectRejectsInvalidURL(t *testing.T) {
	t.Parallel()

	_, err := Connect(context.Background(), config.RedisConfig{
		URL:            "://invalid",
		KeyPrefix:      "url-shortener",
		ConnectTimeout: time.Second,
	})
	if err == nil || !strings.Contains(err.Error(), "parse Redis URL") {
		t.Fatalf("expected Redis URL error, got %v", err)
	}
}

func TestConnectReturnsPingError(t *testing.T) {
	t.Parallel()

	_, err := Connect(context.Background(), config.RedisConfig{
		URL:            "redis://127.0.0.1:1/0",
		KeyPrefix:      "url-shortener",
		ConnectTimeout: 50 * time.Millisecond,
	})
	if err == nil || !strings.Contains(err.Error(), "ping Redis") {
		t.Fatalf("expected Redis ping error, got %v", err)
	}
}

func TestNilClientMethodsAreSafe(t *testing.T) {
	t.Parallel()

	var client *Client

	if err := client.Close(); err != nil {
		t.Fatalf("expected nil client close to be a no-op: %v", err)
	}

	if err := client.Ping(nil); err == nil || !strings.Contains(err.Error(), "client") {
		t.Fatalf("expected nil client ping error, got %v", err)
	}

	if client.Driver() != nil {
		t.Fatal("expected nil Redis driver")
	}

	if client.KeyPrefix() != "" {
		t.Fatalf("expected empty key prefix, got %q", client.KeyPrefix())
	}
}
