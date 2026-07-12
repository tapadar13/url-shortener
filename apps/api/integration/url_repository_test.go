//go:build integration

package integration_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/tapadar13/url-shortener/apps/api/internal/config"
	"github.com/tapadar13/url-shortener/apps/api/internal/platform/mongodb"
	urlmodel "github.com/tapadar13/url-shortener/apps/api/internal/url"
	urlrepository "github.com/tapadar13/url-shortener/apps/api/internal/url/repository/mongodb"
)

const integrationTimeout = 10 * time.Second

func TestURLRepositoryLifecycle(t *testing.T) {
	repository := newIntegrationURLRepository(t)
	ctx, cancel := context.WithTimeout(context.Background(), integrationTimeout)
	defer cancel()

	createdAt := time.Date(2026, 7, 12, 8, 0, 0, 0, time.UTC)
	record, err := urlmodel.New(urlmodel.NewParams{
		LongURL:   "https://example.com/articles/123",
		ShortCode: "Int1234",
		Now:       createdAt,
	})
	if err != nil {
		t.Fatalf("create domain URL: %v", err)
	}

	created, err := repository.Create(ctx, record)
	if err != nil {
		t.Fatalf("create URL: %v", err)
	}

	if created.ID == "" {
		t.Fatal("expected created URL to have an ID")
	}

	found, err := repository.FindByShortCode(ctx, record.ShortCode)
	if err != nil {
		t.Fatalf("find URL: %v", err)
	}

	if found.ID != created.ID || found.LongURL != record.LongURL || found.ShortCode != record.ShortCode {
		t.Fatalf("expected created URL to be found, got %#v", found)
	}

	updatedAt := createdAt.Add(time.Hour)
	updated, err := repository.UpdateLongURL(ctx, urlmodel.UpdateLongURLParams{
		ShortCode: record.ShortCode,
		LongURL:   "https://example.com/new-destination",
		UpdatedAt: updatedAt,
	})
	if err != nil {
		t.Fatalf("update URL: %v", err)
	}

	if updated.LongURL != "https://example.com/new-destination" || !updated.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("expected updated URL record, got %#v", updated)
	}

	accessedAt := updatedAt.Add(time.Hour)
	recorded, err := repository.RecordAccess(ctx, urlmodel.RecordAccessParams{
		ShortCode:  record.ShortCode,
		AccessedAt: accessedAt,
	})
	if err != nil {
		t.Fatalf("record access: %v", err)
	}

	if recorded.AccessCount != 1 || recorded.LastAccessedAt == nil || !recorded.LastAccessedAt.Equal(accessedAt) {
		t.Fatalf("expected recorded access, got %#v", recorded)
	}

	if err := repository.DeleteByShortCode(ctx, record.ShortCode); err != nil {
		t.Fatalf("delete URL: %v", err)
	}

	_, err = repository.FindByShortCode(ctx, record.ShortCode)
	if !errors.Is(err, urlmodel.ErrNotFound) {
		t.Fatalf("expected deleted URL to be missing, got %v", err)
	}
}

func newIntegrationURLRepository(t *testing.T) *urlrepository.Repository {
	t.Helper()

	uri := os.Getenv("MONGODB_INTEGRATION_URI")
	if uri == "" {
		t.Skip("set MONGODB_INTEGRATION_URI to run MongoDB integration tests")
	}

	database := fmt.Sprintf("url_shortener_integration_%d", time.Now().UnixNano())
	ctx, cancel := context.WithTimeout(context.Background(), integrationTimeout)
	defer cancel()

	client, err := mongodb.Connect(ctx, config.MongoDBConfig{
		URI:            uri,
		Database:       database,
		URLsCollection: "urls",
	}, integrationTimeout)
	if err != nil {
		t.Fatalf("connect MongoDB: %v", err)
	}

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), integrationTimeout)
		defer cleanupCancel()

		if database := client.Database(); database != nil {
			if err := database.Drop(cleanupCtx); err != nil {
				t.Errorf("drop integration database: %v", err)
			}
		}

		if err := client.Disconnect(cleanupCtx); err != nil {
			t.Errorf("disconnect MongoDB: %v", err)
		}
	})

	if err := mongodb.EnsureIndexes(ctx, client); err != nil {
		t.Fatalf("ensure MongoDB indexes: %v", err)
	}

	return urlrepository.New(client.URLsCollection())
}
