//go:build integration

package integration_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/tapadar13/url-shortener/apps/api/internal/config"
	"github.com/tapadar13/url-shortener/apps/api/internal/platform/mongodb"
	urlmodel "github.com/tapadar13/url-shortener/apps/api/internal/url"
	urlrepository "github.com/tapadar13/url-shortener/apps/api/internal/url/repository/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const integrationTimeout = 10 * time.Second

func TestURLRepositoryLifecycle(t *testing.T) {
	repository, _ := newIntegrationURLRepository(t)
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

func TestURLRepositoryTreatsExpiredURLAsNotFound(t *testing.T) {
	repository, collection := newIntegrationURLRepository(t)
	ctx, cancel := context.WithTimeout(context.Background(), integrationTimeout)
	defer cancel()

	now := time.Now().UTC()
	shortCode := fmt.Sprintf("Exp%d", now.UnixNano())
	_, err := collection.InsertOne(ctx, bson.D{
		{Key: "url", Value: "https://example.com/expired"},
		{Key: "short_code", Value: shortCode},
		{Key: "access_count", Value: 0},
		{Key: "created_at", Value: now.Add(-2 * time.Hour)},
		{Key: "updated_at", Value: now.Add(-2 * time.Hour)},
		{Key: "expires_at", Value: now.Add(-time.Minute)},
	})
	if err != nil {
		t.Fatalf("insert expired URL: %v", err)
	}

	operations := []struct {
		name string
		run  func() error
	}{
		{
			name: "find",
			run: func() error {
				_, err := repository.FindByShortCode(ctx, shortCode)
				return err
			},
		},
		{
			name: "update",
			run: func() error {
				_, err := repository.UpdateLongURL(ctx, urlmodel.UpdateLongURLParams{
					ShortCode: shortCode,
					LongURL:   "https://example.com/updated",
					UpdatedAt: now,
				})
				return err
			},
		},
		{
			name: "record access",
			run: func() error {
				_, err := repository.RecordAccess(ctx, urlmodel.RecordAccessParams{
					ShortCode:  shortCode,
					AccessedAt: now,
				})
				return err
			},
		},
		{
			name: "delete",
			run: func() error {
				return repository.DeleteByShortCode(ctx, shortCode)
			},
		},
	}

	for _, operation := range operations {
		t.Run(operation.name, func(t *testing.T) {
			if err := operation.run(); !errors.Is(err, urlmodel.ErrNotFound) {
				t.Fatalf("expected expired URL to be unavailable, got %v", err)
			}
		})
	}
}

func TestURLRepositoryRejectsDuplicateShortCode(t *testing.T) {
	repository, _ := newIntegrationURLRepository(t)
	ctx, cancel := context.WithTimeout(context.Background(), integrationTimeout)
	defer cancel()

	now := time.Now().UTC()
	shortCode := fmt.Sprintf("Custom%d", now.UnixNano())
	first, err := urlmodel.New(urlmodel.NewParams{
		LongURL:   "https://example.com/first",
		ShortCode: shortCode,
		Now:       now,
	})
	if err != nil {
		t.Fatalf("create first domain URL: %v", err)
	}

	second, err := urlmodel.New(urlmodel.NewParams{
		LongURL:   "https://example.com/second",
		ShortCode: shortCode,
		Now:       now,
	})
	if err != nil {
		t.Fatalf("create second domain URL: %v", err)
	}

	if _, err := repository.Create(ctx, first); err != nil {
		t.Fatalf("create first URL: %v", err)
	}

	if _, err := repository.Create(ctx, second); !errors.Is(err, urlmodel.ErrDuplicateShortCode) {
		t.Fatalf("expected duplicate short code error, got %v", err)
	}
}

func TestURLRepositoryListsOnlyOwnerURLsNewestFirst(t *testing.T) {
	repository, _ := newIntegrationURLRepository(t)
	ctx, cancel := context.WithTimeout(context.Background(), integrationTimeout)
	defer cancel()
	now := time.Now().UTC()
	for _, record := range []urlmodel.NewParams{
		{OwnerID: "owner-a", LongURL: "https://example.com/old", ShortCode: fmt.Sprintf("Old%d", now.UnixNano()), Now: now.Add(-time.Hour)},
		{OwnerID: "owner-a", LongURL: "https://example.com/new", ShortCode: fmt.Sprintf("New%d", now.UnixNano()), Now: now},
		{OwnerID: "owner-b", LongURL: "https://example.com/other", ShortCode: fmt.Sprintf("Other%d", now.UnixNano()), Now: now.Add(time.Hour)},
	} {
		created, err := urlmodel.New(record)
		if err != nil {
			t.Fatalf("create domain URL: %v", err)
		}
		if _, err := repository.Create(ctx, created); err != nil {
			t.Fatalf("create URL: %v", err)
		}
	}
	urls, err := repository.ListByOwner(ctx, "owner-a", 10)
	if err != nil {
		t.Fatalf("list URLs: %v", err)
	}
	if len(urls) != 2 || urls[0].LongURL != "https://example.com/new" || urls[1].LongURL != "https://example.com/old" {
		t.Fatalf("expected owner-scoped newest-first results, got %+v", urls)
	}
}

func TestURLRepositoryPaginatesStableOwnerResults(t *testing.T) {
	repository, _ := newIntegrationURLRepository(t)
	ctx, cancel := context.WithTimeout(context.Background(), integrationTimeout)
	defer cancel()

	const ownerID = "pagination-owner"
	createdAt := time.Date(2026, 7, 16, 8, 0, 0, 0, time.UTC)
	created := make([]urlmodel.URL, 0, 5)
	for index := range 5 {
		record, err := urlmodel.New(urlmodel.NewParams{
			OwnerID:   ownerID,
			LongURL:   fmt.Sprintf("https://example.com/page/%d", index),
			ShortCode: fmt.Sprintf("Page%02d", index),
			Now:       createdAt,
		})
		if err != nil {
			t.Fatalf("create domain URL %d: %v", index, err)
		}
		if index == 4 {
			record.CreatedAt = createdAt.Add(-time.Hour)
			record.UpdatedAt = record.CreatedAt
		}
		stored, err := repository.Create(ctx, record)
		if err != nil {
			t.Fatalf("create URL %d: %v", index, err)
		}
		created = append(created, stored)
	}

	otherOwner, err := urlmodel.New(urlmodel.NewParams{
		OwnerID:   "other-owner",
		LongURL:   "https://example.com/other",
		ShortCode: "Other1",
		Now:       createdAt.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("create other owner domain URL: %v", err)
	}
	if _, err := repository.Create(ctx, otherOwner); err != nil {
		t.Fatalf("create other owner URL: %v", err)
	}

	sort.Slice(created, func(left, right int) bool {
		if created[left].CreatedAt.Equal(created[right].CreatedAt) {
			return created[left].ID > created[right].ID
		}
		return created[left].CreatedAt.After(created[right].CreatedAt)
	})

	var listed []urlmodel.URL
	var after *urlmodel.ListCursor
	for pageNumber := 1; pageNumber <= 3; pageNumber++ {
		page, err := repository.ListPageByOwner(ctx, urlmodel.ListQuery{
			OwnerID: ownerID,
			Limit:   2,
			After:   after,
		})
		if err != nil {
			t.Fatalf("list page %d: %v", pageNumber, err)
		}
		listed = append(listed, page...)
		if len(page) == 0 {
			break
		}
		last := page[len(page)-1]
		after = &urlmodel.ListCursor{CreatedAt: last.CreatedAt, ID: last.ID}
	}

	if len(listed) != len(created) {
		t.Fatalf("expected %d paginated URLs, got %d", len(created), len(listed))
	}
	for index := range created {
		if listed[index].ID != created[index].ID {
			t.Fatalf("expected URL %q at position %d, got %q", created[index].ID, index, listed[index].ID)
		}
	}
}

func newIntegrationURLRepository(t *testing.T) (*urlrepository.Repository, *mongo.Collection) {
	t.Helper()

	client := newIntegrationMongoClient(t)
	collection := client.URLsCollection()
	return urlrepository.New(collection), collection
}

func newIntegrationMongoClient(t *testing.T) *mongodb.Client {
	t.Helper()

	uri := os.Getenv("MONGODB_INTEGRATION_URI")
	if uri == "" {
		t.Skip("set MONGODB_INTEGRATION_URI to run MongoDB integration tests")
	}

	database := fmt.Sprintf("url_shortener_integration_%d", time.Now().UnixNano())
	ctx, cancel := context.WithTimeout(context.Background(), integrationTimeout)
	defer cancel()

	client, err := mongodb.Connect(ctx, config.MongoDBConfig{
		URI:                  uri,
		Database:             database,
		URLsCollection:       "urls",
		UsersCollection:      "users",
		SessionsCollection:   "sessions",
		RateLimitsCollection: "rate_limits",
		AnalyticsCollection:  "click_analytics",
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

	return client
}
