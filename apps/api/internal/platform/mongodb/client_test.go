package mongodb

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/tapadar13/url-shortener/apps/api/internal/config"
)

func TestConnectRejectsInvalidTimeout(t *testing.T) {
	t.Parallel()

	_, err := Connect(context.Background(), config.MongoDBConfig{
		URI:            "mongodb://localhost:27017",
		Database:       "url_shortener",
		URLsCollection: "urls",
	}, 0)
	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), "timeout") {
		t.Fatalf("expected timeout error, got %q", err.Error())
	}
}

func TestConnectRejectsInvalidURI(t *testing.T) {
	t.Parallel()

	_, err := Connect(context.Background(), config.MongoDBConfig{
		URI:            "://invalid",
		Database:       "url_shortener",
		URLsCollection: "urls",
	}, time.Second)
	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), "create MongoDB client") {
		t.Fatalf("expected client creation error, got %q", err.Error())
	}
}

func TestDisconnectHandlesNilClient(t *testing.T) {
	t.Parallel()

	var client *Client
	if err := client.Disconnect(context.Background()); err != nil {
		t.Fatalf("expected nil client disconnect to be a no-op: %v", err)
	}
}

func TestPingRejectsNilClient(t *testing.T) {
	t.Parallel()

	var client *Client
	err := client.Ping(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), "client") {
		t.Fatalf("expected client error, got %q", err.Error())
	}
}

func TestNilClientAccessorsReturnNil(t *testing.T) {
	t.Parallel()

	var client *Client

	if client.Database() != nil {
		t.Fatal("expected nil database")
	}

	if client.URLsCollection() != nil {
		t.Fatal("expected nil URLs collection")
	}

	if client.RateLimitsCollection() != nil {
		t.Fatal("expected nil rate limits collection")
	}
}
