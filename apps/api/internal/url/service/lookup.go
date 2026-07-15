package service

import (
	"context"
	"fmt"

	"github.com/tapadar13/url-shortener/apps/api/internal/shortcode"
	urlmodel "github.com/tapadar13/url-shortener/apps/api/internal/url"
)

type FindRepository interface {
	FindByShortCode(ctx context.Context, shortCode string) (urlmodel.URL, error)
}

type OwnerFindRepository interface {
	FindByShortCodeForOwner(ctx context.Context, ownerID, shortCode string) (urlmodel.URL, error)
}

type LookupService struct {
	repository FindRepository
}

func NewLookupService(repository FindRepository) (*LookupService, error) {
	if repository == nil {
		return nil, ErrRepositoryRequired
	}

	return &LookupService{
		repository: repository,
	}, nil
}

func (s *LookupService) GetByShortCode(ctx context.Context, shortCode string) (urlmodel.URL, error) {
	if s == nil || s.repository == nil {
		return urlmodel.URL{}, ErrRepositoryRequired
	}

	normalizedShortCode, err := shortcode.Normalize(shortCode)
	if err != nil {
		return urlmodel.URL{}, err
	}

	record, err := s.repository.FindByShortCode(ctx, normalizedShortCode)
	if err != nil {
		return urlmodel.URL{}, fmt.Errorf("find URL by short code: %w", err)
	}

	return record, nil
}

func (s *LookupService) GetByShortCodeForOwner(ctx context.Context, ownerID, shortCode string) (urlmodel.URL, error) {
	if s == nil || s.repository == nil {
		return urlmodel.URL{}, ErrRepositoryRequired
	}
	if ownerID == "" {
		return urlmodel.URL{}, ErrOwnerRequired
	}
	repository, ok := s.repository.(OwnerFindRepository)
	if !ok {
		return urlmodel.URL{}, ErrOwnerRepositoryUnsupported
	}
	normalizedShortCode, err := shortcode.Normalize(shortCode)
	if err != nil {
		return urlmodel.URL{}, err
	}
	record, err := repository.FindByShortCodeForOwner(ctx, ownerID, normalizedShortCode)
	if err != nil {
		return urlmodel.URL{}, fmt.Errorf("find URL by owner: %w", err)
	}
	return record, nil
}
