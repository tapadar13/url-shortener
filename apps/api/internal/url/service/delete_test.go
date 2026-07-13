package service

import (
	"context"
	"errors"
	"testing"

	"github.com/tapadar13/url-shortener/apps/api/internal/shortcode"
)

func TestNewDeleteServiceRequiresRepository(t *testing.T) {
	t.Parallel()

	_, err := NewDeleteService(nil, DeleteOptions{})
	if !errors.Is(err, ErrRepositoryRequired) {
		t.Fatalf("expected repository error, got %v", err)
	}
}

func TestDeleteServiceDeletesNormalizedShortCode(t *testing.T) {
	t.Parallel()

	repository := &fakeDeleteRepository{}
	service, err := NewDeleteService(repository, DeleteOptions{})
	if err != nil {
		t.Fatalf("expected delete service: %v", err)
	}

	if err := service.DeleteByShortCode(context.Background(), " AbC123 "); err != nil {
		t.Fatalf("expected deletion to succeed: %v", err)
	}

	if repository.shortCode != "AbC123" {
		t.Fatalf("expected normalized short code, got %q", repository.shortCode)
	}
}

func TestDeleteServiceRejectsInvalidShortCodeBeforeRepository(t *testing.T) {
	t.Parallel()

	repository := &fakeDeleteRepository{}
	service, err := NewDeleteService(repository, DeleteOptions{})
	if err != nil {
		t.Fatalf("expected delete service: %v", err)
	}

	err = service.DeleteByShortCode(context.Background(), "invalid-code")
	if !errors.Is(err, shortcode.ErrInvalidChars) {
		t.Fatalf("expected short code validation error, got %v", err)
	}

	if repository.called {
		t.Fatal("expected repository not to be called")
	}
}

func TestDeleteServiceReturnsRepositoryError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("database unavailable")
	service, err := NewDeleteService(&fakeDeleteRepository{err: expectedErr}, DeleteOptions{})
	if err != nil {
		t.Fatalf("expected delete service: %v", err)
	}

	err = service.DeleteByShortCode(context.Background(), "AbC123")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected repository error, got %v", err)
	}
}

func TestDeleteServiceInvalidatesRedirectCache(t *testing.T) {
	t.Parallel()

	cache := &fakeDeleteCacheInvalidator{}
	service, err := NewDeleteService(&fakeDeleteRepository{}, DeleteOptions{Cache: cache})
	if err != nil {
		t.Fatalf("expected delete service: %v", err)
	}

	if err := service.DeleteByShortCode(context.Background(), " AbC123 "); err != nil {
		t.Fatalf("expected deletion to succeed: %v", err)
	}

	if cache.shortCode != "AbC123" {
		t.Fatalf("expected normalized cache invalidation code, got %q", cache.shortCode)
	}
}

func TestDeleteServiceReportsCacheInvalidationError(t *testing.T) {
	t.Parallel()

	cacheErr := errors.New("Redis unavailable")
	reported := make(chan error, 1)
	cache := &fakeDeleteCacheInvalidator{err: cacheErr}
	service, err := NewDeleteService(&fakeDeleteRepository{}, DeleteOptions{
		Cache: cache,
		OnCacheError: func(err error) {
			reported <- err
		},
	})
	if err != nil {
		t.Fatalf("expected delete service: %v", err)
	}

	if err := service.DeleteByShortCode(context.Background(), "AbC123"); err != nil {
		t.Fatalf("expected committed deletion despite cache error: %v", err)
	}

	select {
	case err := <-reported:
		if !errors.Is(err, cacheErr) {
			t.Fatalf("expected reported cache error, got %v", err)
		}
	default:
		t.Fatal("expected cache invalidation error to be reported")
	}
}

func TestDeleteServiceDoesNotInvalidateCacheOnRepositoryError(t *testing.T) {
	t.Parallel()

	cache := &fakeDeleteCacheInvalidator{}
	service, err := NewDeleteService(&fakeDeleteRepository{err: errors.New("database unavailable")}, DeleteOptions{Cache: cache})
	if err != nil {
		t.Fatalf("expected delete service: %v", err)
	}

	_ = service.DeleteByShortCode(context.Background(), "AbC123")
	if cache.called {
		t.Fatal("expected cache not to be invalidated")
	}
}

func TestDeleteServiceRequiresConfiguredRepository(t *testing.T) {
	t.Parallel()

	err := (&DeleteService{}).DeleteByShortCode(context.Background(), "AbC123")
	if !errors.Is(err, ErrRepositoryRequired) {
		t.Fatalf("expected repository error, got %v", err)
	}
}

type fakeDeleteRepository struct {
	err       error
	shortCode string
	called    bool
}

func (r *fakeDeleteRepository) DeleteByShortCode(_ context.Context, shortCode string) error {
	r.called = true
	r.shortCode = shortCode

	return r.err
}

type fakeDeleteCacheInvalidator struct {
	shortCode string
	err       error
	called    bool
}

func (c *fakeDeleteCacheInvalidator) Delete(_ context.Context, shortCode string) error {
	c.called = true
	c.shortCode = shortCode
	return c.err
}
