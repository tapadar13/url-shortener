package service

import (
	"context"
	"errors"
	"testing"

	"github.com/tapadar13/url-shortener/apps/api/internal/shortcode"
)

func TestNewDeleteServiceRequiresRepository(t *testing.T) {
	t.Parallel()

	_, err := NewDeleteService(nil)
	if !errors.Is(err, ErrRepositoryRequired) {
		t.Fatalf("expected repository error, got %v", err)
	}
}

func TestDeleteServiceDeletesNormalizedShortCode(t *testing.T) {
	t.Parallel()

	repository := &fakeDeleteRepository{}
	service, err := NewDeleteService(repository)
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
	service, err := NewDeleteService(repository)
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
	service, err := NewDeleteService(&fakeDeleteRepository{err: expectedErr})
	if err != nil {
		t.Fatalf("expected delete service: %v", err)
	}

	err = service.DeleteByShortCode(context.Background(), "AbC123")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected repository error, got %v", err)
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
