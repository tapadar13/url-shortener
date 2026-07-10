package service

import (
	"context"
	"errors"
	"testing"

	"github.com/tapadar13/url-shortener/apps/api/internal/shortcode"
	urlmodel "github.com/tapadar13/url-shortener/apps/api/internal/url"
)

func TestNewLookupServiceRequiresRepository(t *testing.T) {
	t.Parallel()

	_, err := NewLookupService(nil)
	if !errors.Is(err, ErrRepositoryRequired) {
		t.Fatalf("expected repository error, got %v", err)
	}
}

func TestLookupServiceGetsURLByNormalizedShortCode(t *testing.T) {
	t.Parallel()

	repository := &fakeFindRepository{
		record: urlmodel.URL{
			ID:        "507f1f77bcf86cd799439011",
			LongURL:   "https://example.com/articles/123",
			ShortCode: "AbC123",
		},
	}
	service, err := NewLookupService(repository)
	if err != nil {
		t.Fatalf("expected lookup service: %v", err)
	}

	found, err := service.GetByShortCode(context.Background(), " AbC123 ")
	if err != nil {
		t.Fatalf("expected URL lookup to succeed: %v", err)
	}

	if repository.shortCode != "AbC123" {
		t.Fatalf("expected normalized short code, got %q", repository.shortCode)
	}

	if found != repository.record {
		t.Fatalf("expected found URL %#v, got %#v", repository.record, found)
	}
}

func TestLookupServiceRejectsInvalidShortCodeBeforeRepository(t *testing.T) {
	t.Parallel()

	repository := &fakeFindRepository{}
	service, err := NewLookupService(repository)
	if err != nil {
		t.Fatalf("expected lookup service: %v", err)
	}

	_, err = service.GetByShortCode(context.Background(), "invalid-code")
	if !errors.Is(err, shortcode.ErrInvalidChars) {
		t.Fatalf("expected short code validation error, got %v", err)
	}

	if repository.called {
		t.Fatal("expected repository not to be called")
	}
}

func TestLookupServiceReturnsRepositoryError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("database unavailable")
	service, err := NewLookupService(&fakeFindRepository{err: expectedErr})
	if err != nil {
		t.Fatalf("expected lookup service: %v", err)
	}

	_, err = service.GetByShortCode(context.Background(), "AbC123")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected repository error, got %v", err)
	}
}

func TestLookupServiceRequiresConfiguredRepository(t *testing.T) {
	t.Parallel()

	_, err := (&LookupService{}).GetByShortCode(context.Background(), "AbC123")
	if !errors.Is(err, ErrRepositoryRequired) {
		t.Fatalf("expected repository error, got %v", err)
	}
}

type fakeFindRepository struct {
	record    urlmodel.URL
	err       error
	shortCode string
	called    bool
}

func (r *fakeFindRepository) FindByShortCode(_ context.Context, shortCode string) (urlmodel.URL, error) {
	r.called = true
	r.shortCode = shortCode

	if r.err != nil {
		return urlmodel.URL{}, r.err
	}

	return r.record, nil
}
