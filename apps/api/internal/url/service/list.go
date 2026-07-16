package service

import (
	"context"
	"errors"
	"fmt"

	urlmodel "github.com/tapadar13/url-shortener/apps/api/internal/url"
)

const (
	DefaultURLListLimit int64 = 25
	MaxURLListLimit     int64 = 100
)

type ListRepository interface {
	ListByOwner(ctx context.Context, ownerID string, limit int64) ([]urlmodel.URL, error)
	ListPageByOwner(ctx context.Context, query urlmodel.ListQuery) ([]urlmodel.URL, error)
}

type ListParams struct {
	OwnerID string
	Limit   int64
	Cursor  string
}

type ListPage struct {
	Items      []urlmodel.URL
	NextCursor string
}

type ListService struct{ repository ListRepository }

func NewListService(repository ListRepository) (*ListService, error) {
	if repository == nil {
		return nil, ErrRepositoryRequired
	}
	return &ListService{repository: repository}, nil
}

func (s *ListService) ListByOwner(ctx context.Context, ownerID string, limit int64) ([]urlmodel.URL, error) {
	page, err := s.ListPageByOwner(ctx, ListParams{OwnerID: ownerID, Limit: limit})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (s *ListService) ListPageByOwner(ctx context.Context, params ListParams) (ListPage, error) {
	if s == nil || s.repository == nil {
		return ListPage{}, ErrRepositoryRequired
	}
	if params.OwnerID == "" {
		return ListPage{}, ErrOwnerRequired
	}
	if params.Limit == 0 {
		params.Limit = DefaultURLListLimit
	}
	if params.Limit < 1 || params.Limit > MaxURLListLimit {
		return ListPage{}, errors.New("URL list limit must be between 1 and 100")
	}
	var after *urlmodel.ListCursor
	if params.Cursor != "" {
		decoded, err := urlmodel.DecodeListCursor(params.Cursor)
		if err != nil {
			return ListPage{}, err
		}
		after = &decoded
	}
	urls, err := s.repository.ListPageByOwner(ctx, urlmodel.ListQuery{OwnerID: params.OwnerID, Limit: params.Limit + 1, After: after})
	if err != nil {
		return ListPage{}, fmt.Errorf("list URLs by owner: %w", err)
	}
	page := ListPage{Items: urls}
	if int64(len(urls)) <= params.Limit {
		return page, nil
	}
	page.Items = urls[:params.Limit]
	last := page.Items[len(page.Items)-1]
	page.NextCursor, err = urlmodel.EncodeListCursor(urlmodel.ListCursor{CreatedAt: last.CreatedAt, ID: last.ID})
	if err != nil {
		return ListPage{}, fmt.Errorf("encode next URL list cursor: %w", err)
	}
	return page, nil
}
