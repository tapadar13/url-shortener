package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/tapadar13/url-shortener/apps/api/internal/shortcode"
	urlmodel "github.com/tapadar13/url-shortener/apps/api/internal/url"
)

func TestNewUpdateServiceRequiresRepository(t *testing.T) {
	t.Parallel()

	_, err := NewUpdateService(nil, UpdateOptions{})
	if !errors.Is(err, ErrRepositoryRequired) {
		t.Fatalf("expected repository error, got %v", err)
	}
}

func TestUpdateServiceUpdatesNormalizedURL(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 10, 10, 0, 0, 0, time.FixedZone("IST", 5*60*60+30*60))
	repository := &fakeUpdateRepository{
		record: urlmodel.URL{
			ID:        "507f1f77bcf86cd799439011",
			LongURL:   "https://example.com/updated",
			ShortCode: "AbC123",
			UpdatedAt: now.UTC(),
		},
	}
	service, err := NewUpdateService(repository, UpdateOptions{
		Now: func() time.Time {
			return now
		},
	})
	if err != nil {
		t.Fatalf("expected update service: %v", err)
	}

	updated, err := service.UpdateLongURL(context.Background(), UpdateParams{
		ShortCode: " AbC123 ",
		LongURL:   " https://example.com/updated ",
	})
	if err != nil {
		t.Fatalf("expected URL update to succeed: %v", err)
	}

	if repository.params.ShortCode != "AbC123" || repository.params.LongURL != "https://example.com/updated" {
		t.Fatalf("expected normalized update params, got %#v", repository.params)
	}

	if !repository.params.UpdatedAt.Equal(now.UTC()) {
		t.Fatalf("expected update timestamp %s, got %s", now.UTC(), repository.params.UpdatedAt)
	}

	if updated != repository.record {
		t.Fatalf("expected updated URL %#v, got %#v", repository.record, updated)
	}
}

func TestUpdateServiceRejectsInvalidShortCodeBeforeRepository(t *testing.T) {
	t.Parallel()

	repository := &fakeUpdateRepository{}
	service, err := NewUpdateService(repository, UpdateOptions{})
	if err != nil {
		t.Fatalf("expected update service: %v", err)
	}

	_, err = service.UpdateLongURL(context.Background(), UpdateParams{
		ShortCode: "invalid-code",
		LongURL:   "https://example.com/updated",
	})
	if !errors.Is(err, shortcode.ErrInvalidChars) {
		t.Fatalf("expected short code validation error, got %v", err)
	}

	if repository.called {
		t.Fatal("expected repository not to be called")
	}
}

func TestUpdateServiceRejectsInvalidLongURLBeforeRepository(t *testing.T) {
	t.Parallel()

	repository := &fakeUpdateRepository{}
	service, err := NewUpdateService(repository, UpdateOptions{})
	if err != nil {
		t.Fatalf("expected update service: %v", err)
	}

	_, err = service.UpdateLongURL(context.Background(), UpdateParams{
		ShortCode: "AbC123",
		LongURL:   "ftp://example.com",
	})
	if !errors.Is(err, urlmodel.ErrLongURLSchemeUnsupported) {
		t.Fatalf("expected URL validation error, got %v", err)
	}

	if repository.called {
		t.Fatal("expected repository not to be called")
	}
}

func TestUpdateServiceReturnsRepositoryError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("database unavailable")
	service, err := NewUpdateService(&fakeUpdateRepository{err: expectedErr}, UpdateOptions{})
	if err != nil {
		t.Fatalf("expected update service: %v", err)
	}

	_, err = service.UpdateLongURL(context.Background(), UpdateParams{
		ShortCode: "AbC123",
		LongURL:   "https://example.com/updated",
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected repository error, got %v", err)
	}
}

func TestUpdateServiceRequiresConfiguredRepository(t *testing.T) {
	t.Parallel()

	_, err := (&UpdateService{}).UpdateLongURL(context.Background(), UpdateParams{
		ShortCode: "AbC123",
		LongURL:   "https://example.com/updated",
	})
	if !errors.Is(err, ErrRepositoryRequired) {
		t.Fatalf("expected repository error, got %v", err)
	}
}

type fakeUpdateRepository struct {
	record urlmodel.URL
	err    error
	params urlmodel.UpdateLongURLParams
	called bool
}

func (r *fakeUpdateRepository) UpdateLongURL(_ context.Context, params urlmodel.UpdateLongURLParams) (urlmodel.URL, error) {
	r.called = true
	r.params = params

	if r.err != nil {
		return urlmodel.URL{}, r.err
	}

	return r.record, nil
}
