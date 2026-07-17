package url

import (
	"errors"
	"testing"
	"time"

	"github.com/tapadar13/url-shortener/apps/api/internal/shortcode"
)

func TestNewCreatesURLRecord(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 9, 12, 30, 0, 0, time.FixedZone("IST", 5*60*60+30*60))

	record, err := New(NewParams{
		LongURL:   " https://example.com/articles/123 ",
		ShortCode: " AbCd123 ",
		Now:       now,
	})
	if err != nil {
		t.Fatalf("expected record to be valid: %v", err)
	}

	if record.LongURL != "https://example.com/articles/123" {
		t.Fatalf("expected trimmed long URL, got %q", record.LongURL)
	}

	if record.ShortCode != "AbCd123" {
		t.Fatalf("expected trimmed short code, got %q", record.ShortCode)
	}

	if record.AccessCount != 0 {
		t.Fatalf("expected zero access count, got %d", record.AccessCount)
	}

	expectedTime := now.UTC()
	if !record.CreatedAt.Equal(expectedTime) {
		t.Fatalf("expected created time %s, got %s", expectedTime, record.CreatedAt)
	}

	if !record.UpdatedAt.Equal(expectedTime) {
		t.Fatalf("expected updated time %s, got %s", expectedTime, record.UpdatedAt)
	}

	if record.LastAccessedAt != nil {
		t.Fatal("expected last accessed time to be empty")
	}
}

func TestNewNormalizesExpiration(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 9, 7, 0, 0, 0, time.UTC)
	expiresAt := now.Add(24 * time.Hour).In(time.FixedZone("IST", 5*60*60+30*60))

	record, err := New(NewParams{
		LongURL:   "https://example.com",
		ShortCode: "abc123",
		Now:       now,
		ExpiresAt: &expiresAt,
	})
	if err != nil {
		t.Fatalf("expected record to be valid: %v", err)
	}

	if record.ExpiresAt == nil {
		t.Fatal("expected expiration to be set")
	}

	if !record.ExpiresAt.Equal(now.Add(24*time.Hour)) || record.ExpiresAt.Location() != time.UTC {
		t.Fatalf("expected UTC expiration %s, got %s", now.Add(24*time.Hour), *record.ExpiresAt)
	}
}

func TestNewRejectsNonFutureExpiration(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 9, 7, 0, 0, 0, time.UTC)
	for _, expiresAt := range []time.Time{now, now.Add(-time.Nanosecond)} {
		_, err := New(NewParams{
			LongURL:   "https://example.com",
			ShortCode: "abc123",
			Now:       now,
			ExpiresAt: &expiresAt,
		})
		if !errors.Is(err, ErrExpirationNotFuture) {
			t.Fatalf("expected expiration error, got %v", err)
		}
	}
}

func TestNewRejectsInvalidRecord(t *testing.T) {
	t.Parallel()

	_, err := New(NewParams{})
	if err == nil {
		t.Fatal("expected validation error")
	}

	for _, expected := range []error{
		ErrLongURLRequired,
		ErrShortCodeRequired,
		ErrTimestampRequired,
	} {
		if !errors.Is(err, expected) {
			t.Fatalf("expected error %q, got %v", expected, err)
		}
	}
}

func TestValidateRejectsNegativeAccessCount(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 9, 7, 0, 0, 0, time.UTC)
	record := URL{
		LongURL:     "https://example.com",
		ShortCode:   "abc123",
		AccessCount: -1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err := record.Validate()
	if !errors.Is(err, ErrNegativeAccesses) {
		t.Fatalf("expected negative access error, got %v", err)
	}
}

func TestValidateRejectsZeroExpiration(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 9, 7, 0, 0, 0, time.UTC)
	zero := time.Time{}
	record := URL{
		LongURL:   "https://example.com",
		ShortCode: "abc123",
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: &zero,
	}

	err := record.Validate()
	if !errors.Is(err, ErrExpirationInvalid) {
		t.Fatalf("expected invalid expiration error, got %v", err)
	}
}

func TestValidateRejectsInvalidShortCode(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 9, 7, 0, 0, 0, time.UTC)
	record := URL{
		LongURL:     "https://example.com",
		ShortCode:   "abc-123",
		AccessCount: 0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err := record.Validate()
	if !errors.Is(err, shortcode.ErrInvalidChars) {
		t.Fatalf("expected invalid short code error, got %v", err)
	}
}

func TestWithLongURLUpdatesURLAndTimestamp(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 7, 9, 7, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Hour)

	record, err := New(NewParams{
		LongURL:   "https://example.com/old",
		ShortCode: "abc123",
		Now:       createdAt,
	})
	if err != nil {
		t.Fatalf("expected record to be valid: %v", err)
	}

	updated, err := record.WithLongURL(" https://example.com/new ", updatedAt)
	if err != nil {
		t.Fatalf("expected update to be valid: %v", err)
	}

	if updated.LongURL != "https://example.com/new" {
		t.Fatalf("expected updated URL, got %q", updated.LongURL)
	}

	if !updated.CreatedAt.Equal(createdAt) {
		t.Fatalf("expected created time to be unchanged, got %s", updated.CreatedAt)
	}

	if !updated.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("expected updated time %s, got %s", updatedAt, updated.UpdatedAt)
	}
}

func TestWithAccessRecordedIncrementsAccessCount(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 7, 9, 7, 0, 0, 0, time.UTC)
	accessedAt := createdAt.Add(2 * time.Hour)

	record, err := New(NewParams{
		LongURL:   "https://example.com",
		ShortCode: "abc123",
		Now:       createdAt,
	})
	if err != nil {
		t.Fatalf("expected record to be valid: %v", err)
	}

	updated, err := record.WithAccessRecorded(accessedAt)
	if err != nil {
		t.Fatalf("expected access update to be valid: %v", err)
	}

	if updated.AccessCount != 1 {
		t.Fatalf("expected access count 1, got %d", updated.AccessCount)
	}

	if updated.LastAccessedAt == nil {
		t.Fatal("expected last accessed time")
	}

	if !updated.LastAccessedAt.Equal(accessedAt) {
		t.Fatalf("expected last accessed time %s, got %s", accessedAt, *updated.LastAccessedAt)
	}

	if !updated.UpdatedAt.Equal(createdAt) {
		t.Fatalf("expected updated time to remain unchanged, got %s", updated.UpdatedAt)
	}
}
