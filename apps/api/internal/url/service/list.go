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
}

type ListService struct{ repository ListRepository }

func NewListService(repository ListRepository) (*ListService, error) {
	if repository == nil {
		return nil, ErrRepositoryRequired
	}
	return &ListService{repository: repository}, nil
}

func (s *ListService) ListByOwner(ctx context.Context, ownerID string, limit int64) ([]urlmodel.URL, error) {
	if s == nil || s.repository == nil {
		return nil, ErrRepositoryRequired
	}
	if ownerID == "" {
		return nil, ErrOwnerRequired
	}
	if limit == 0 {
		limit = DefaultURLListLimit
	}
	if limit < 1 || limit > MaxURLListLimit {
		return nil, errors.New("URL list limit must be between 1 and 100")
	}
	urls, err := s.repository.ListByOwner(ctx, ownerID, limit)
	if err != nil {
		return nil, fmt.Errorf("list URLs by owner: %w", err)
	}
	return urls, nil
}
