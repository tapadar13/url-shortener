package mongodb

import (
	"errors"
	"strings"
	"testing"
	"time"

	urlmodel "github.com/tapadar13/url-shortener/apps/api/internal/url"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestNewURLDocumentMapsDomainRecord(t *testing.T) {
	t.Parallel()

	id := bson.NewObjectID()
	createdAt := time.Date(2026, 7, 9, 7, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Hour)
	lastAccessedAt := updatedAt.Add(time.Hour)
	expiresAt := updatedAt.Add(24 * time.Hour)

	record := urlmodel.URL{
		ID:             id.Hex(),
		LongURL:        "https://example.com/articles/123",
		ShortCode:      "AbC123",
		AccessCount:    42,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
		LastAccessedAt: &lastAccessedAt,
		ExpiresAt:      &expiresAt,
	}

	doc, err := newURLDocument(record)
	if err != nil {
		t.Fatalf("expected document to be created: %v", err)
	}

	if doc.ID != id {
		t.Fatalf("expected id %s, got %s", id.Hex(), doc.ID.Hex())
	}

	if doc.LongURL != record.LongURL {
		t.Fatalf("expected URL %q, got %q", record.LongURL, doc.LongURL)
	}

	if doc.ShortCode != record.ShortCode {
		t.Fatalf("expected short code %q, got %q", record.ShortCode, doc.ShortCode)
	}

	if doc.AccessCount != record.AccessCount {
		t.Fatalf("expected access count %d, got %d", record.AccessCount, doc.AccessCount)
	}

	if !doc.CreatedAt.Equal(createdAt) || !doc.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("expected timestamps to be mapped")
	}

	if doc.LastAccessedAt == nil || !doc.LastAccessedAt.Equal(lastAccessedAt) {
		t.Fatalf("expected last accessed time to be mapped")
	}

	if doc.ExpiresAt == nil || !doc.ExpiresAt.Equal(expiresAt) {
		t.Fatalf("expected expiration time to be mapped")
	}
}

func TestNewURLDocumentAllowsEmptyIDForInsert(t *testing.T) {
	t.Parallel()

	record, err := urlmodel.New(urlmodel.NewParams{
		LongURL:   "https://example.com",
		ShortCode: "AbC123",
		Now:       time.Date(2026, 7, 9, 7, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("expected domain record to be valid: %v", err)
	}

	doc, err := newURLDocument(record)
	if err != nil {
		t.Fatalf("expected document to be created: %v", err)
	}

	if !doc.ID.IsZero() {
		t.Fatalf("expected empty object id for insert, got %s", doc.ID.Hex())
	}
}

func TestNewURLDocumentRejectsInvalidID(t *testing.T) {
	t.Parallel()

	_, err := newURLDocument(urlmodel.URL{ID: "invalid-object-id"})
	if err == nil {
		t.Fatal("expected invalid id error")
	}

	if !strings.Contains(err.Error(), "parse URL id") {
		t.Fatalf("expected id parsing error, got %q", err.Error())
	}
}

func TestURLDocumentToDomainMapsRecord(t *testing.T) {
	t.Parallel()

	id := bson.NewObjectID()
	createdAt := time.Date(2026, 7, 9, 7, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Hour)
	lastAccessedAt := updatedAt.Add(time.Hour)
	expiresAt := updatedAt.Add(24 * time.Hour)

	doc := urlDocument{
		ID:             id,
		LongURL:        "https://example.com/articles/123",
		ShortCode:      "AbC123",
		AccessCount:    3,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
		LastAccessedAt: &lastAccessedAt,
		ExpiresAt:      &expiresAt,
	}

	record, err := doc.toDomain()
	if err != nil {
		t.Fatalf("expected domain record to be created: %v", err)
	}

	if record.ID != id.Hex() {
		t.Fatalf("expected id %q, got %q", id.Hex(), record.ID)
	}

	if record.LongURL != doc.LongURL {
		t.Fatalf("expected URL %q, got %q", doc.LongURL, record.LongURL)
	}

	if record.ShortCode != doc.ShortCode {
		t.Fatalf("expected short code %q, got %q", doc.ShortCode, record.ShortCode)
	}

	if record.AccessCount != doc.AccessCount {
		t.Fatalf("expected access count %d, got %d", doc.AccessCount, record.AccessCount)
	}

	if !record.CreatedAt.Equal(createdAt) || !record.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("expected timestamps to be mapped")
	}

	if record.LastAccessedAt == nil || !record.LastAccessedAt.Equal(lastAccessedAt) {
		t.Fatalf("expected last accessed time to be mapped")
	}

	if record.ExpiresAt == nil || !record.ExpiresAt.Equal(expiresAt) {
		t.Fatalf("expected expiration time to be mapped")
	}
}

func TestURLDocumentToDomainRejectsInvalidDocument(t *testing.T) {
	t.Parallel()

	_, err := (urlDocument{}).toDomain()
	if err == nil {
		t.Fatal("expected validation error")
	}

	if !errors.Is(err, urlmodel.ErrLongURLRequired) {
		t.Fatalf("expected long URL error, got %v", err)
	}
}
