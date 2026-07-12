package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/tapadar13/url-shortener/apps/api/internal/shortcode"
	urlmodel "github.com/tapadar13/url-shortener/apps/api/internal/url"
)

func TestNewRedirectServiceRequiresRepository(t *testing.T) {
	t.Parallel()

	_, err := NewRedirectService(nil, RedirectOptions{})
	if !errors.Is(err, ErrRepositoryRequired) {
		t.Fatalf("expected repository error, got %v", err)
	}
}

func TestRedirectServiceRecordsAccessForNormalizedShortCode(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 11, 8, 0, 0, 0, time.FixedZone("IST", 5*60*60+30*60))
	repository := &fakeAccessRepository{
		record: urlmodel.URL{
			ID:        "507f1f77bcf86cd799439011",
			LongURL:   "https://example.com/articles/123",
			ShortCode: "AbC123",
		},
	}
	service, err := NewRedirectService(repository, RedirectOptions{
		Now: func() time.Time {
			return now
		},
	})
	if err != nil {
		t.Fatalf("expected redirect service: %v", err)
	}

	resolved, err := service.Resolve(context.Background(), " AbC123 ")
	if err != nil {
		t.Fatalf("expected short URL to resolve: %v", err)
	}

	if repository.params.ShortCode != "AbC123" {
		t.Fatalf("expected normalized short code, got %q", repository.params.ShortCode)
	}

	if !repository.params.AccessedAt.Equal(now.UTC()) {
		t.Fatalf("expected access timestamp %s, got %s", now.UTC(), repository.params.AccessedAt)
	}

	if resolved != repository.record {
		t.Fatalf("expected resolved URL %#v, got %#v", repository.record, resolved)
	}
}

func TestRedirectServiceRejectsInvalidShortCodeBeforeRepository(t *testing.T) {
	t.Parallel()

	repository := &fakeAccessRepository{}
	service, err := NewRedirectService(repository, RedirectOptions{})
	if err != nil {
		t.Fatalf("expected redirect service: %v", err)
	}

	_, err = service.Resolve(context.Background(), "invalid-code")
	if !errors.Is(err, shortcode.ErrInvalidChars) {
		t.Fatalf("expected short code validation error, got %v", err)
	}

	if repository.called {
		t.Fatal("expected repository not to be called")
	}
}

func TestRedirectServiceReturnsRepositoryError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("database unavailable")
	service, err := NewRedirectService(&fakeAccessRepository{err: expectedErr}, RedirectOptions{})
	if err != nil {
		t.Fatalf("expected redirect service: %v", err)
	}

	_, err = service.Resolve(context.Background(), "AbC123")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected repository error, got %v", err)
	}
}

func TestRedirectServiceRequiresConfiguredRepository(t *testing.T) {
	t.Parallel()

	_, err := (&RedirectService{}).Resolve(context.Background(), "AbC123")
	if !errors.Is(err, ErrRepositoryRequired) {
		t.Fatalf("expected repository error, got %v", err)
	}
}

type fakeAccessRepository struct {
	record urlmodel.URL
	err    error
	params urlmodel.RecordAccessParams
	called bool
}

func (r *fakeAccessRepository) RecordAccess(_ context.Context, params urlmodel.RecordAccessParams) (urlmodel.URL, error) {
	r.called = true
	r.params = params

	if r.err != nil {
		return urlmodel.URL{}, r.err
	}

	return r.record, nil
}
