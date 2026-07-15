package service

import (
	"context"
	"testing"

	urlmodel "github.com/tapadar13/url-shortener/apps/api/internal/url"
)

type fakeListRepository struct {
	limit int64
	urls  []urlmodel.URL
}

func (r *fakeListRepository) ListByOwner(_ context.Context, _ string, limit int64) ([]urlmodel.URL, error) {
	r.limit = limit
	return r.urls, nil
}

func TestListByOwnerUsesBoundedDefaultLimit(t *testing.T) {
	repository := &fakeListRepository{}
	service, err := NewListService(repository)
	if err != nil {
		t.Fatalf("create list service: %v", err)
	}
	if _, err := service.ListByOwner(context.Background(), "owner-1", 0); err != nil {
		t.Fatalf("list URLs: %v", err)
	}
	if repository.limit != DefaultURLListLimit {
		t.Fatalf("expected default limit %d, got %d", DefaultURLListLimit, repository.limit)
	}
}

func TestListByOwnerRejectsInvalidLimit(t *testing.T) {
	service, err := NewListService(&fakeListRepository{})
	if err != nil {
		t.Fatalf("create list service: %v", err)
	}
	if _, err := service.ListByOwner(context.Background(), "owner-1", MaxURLListLimit+1); err == nil {
		t.Fatal("expected invalid limit error")
	}
}
