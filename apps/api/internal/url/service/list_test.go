package service

import (
	"context"
	"testing"
	"time"

	urlmodel "github.com/tapadar13/url-shortener/apps/api/internal/url"
)

type fakeListRepository struct {
	query urlmodel.ListQuery
	urls  []urlmodel.URL
}

func (r *fakeListRepository) ListPageByOwner(_ context.Context, query urlmodel.ListQuery) ([]urlmodel.URL, error) {
	r.query = query
	return r.urls, nil
}

func (r *fakeListRepository) ListByOwner(_ context.Context, _ string, limit int64) ([]urlmodel.URL, error) {
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
	if repository.query.Limit != DefaultURLListLimit+1 {
		t.Fatalf("expected lookahead limit %d, got %d", DefaultURLListLimit+1, repository.query.Limit)
	}
}

func TestListPageByOwnerReturnsNextCursor(t *testing.T) {
	now := time.Date(2026, time.July, 16, 10, 0, 0, 0, time.UTC)
	repository := &fakeListRepository{urls: []urlmodel.URL{
		{ID: "507f1f77bcf86cd799439011", CreatedAt: now},
		{ID: "507f1f77bcf86cd799439012", CreatedAt: now.Add(-time.Minute)},
	}}
	service, err := NewListService(repository)
	if err != nil {
		t.Fatalf("create list service: %v", err)
	}
	page, err := service.ListPageByOwner(context.Background(), ListParams{OwnerID: "owner-1", Limit: 1})
	if err != nil {
		t.Fatalf("list page: %v", err)
	}
	if len(page.Items) != 1 || page.NextCursor == "" {
		t.Fatalf("expected one item and next cursor, got %+v", page)
	}
	cursor, err := urlmodel.DecodeListCursor(page.NextCursor)
	if err != nil || cursor.ID != page.Items[0].ID {
		t.Fatalf("decode next cursor: cursor=%+v err=%v", cursor, err)
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
