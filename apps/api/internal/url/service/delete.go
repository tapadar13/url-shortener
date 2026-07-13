package service

import (
	"context"
	"fmt"

	"github.com/tapadar13/url-shortener/apps/api/internal/shortcode"
)

type DeleteRepository interface {
	DeleteByShortCode(ctx context.Context, shortCode string) error
}

type DeleteOptions struct {
	Cache        RedirectCacheInvalidator
	OnCacheError func(error)
}

type DeleteService struct {
	repository   DeleteRepository
	cache        RedirectCacheInvalidator
	onCacheError func(error)
}

func NewDeleteService(repository DeleteRepository, options DeleteOptions) (*DeleteService, error) {
	if repository == nil {
		return nil, ErrRepositoryRequired
	}

	return &DeleteService{
		repository:   repository,
		cache:        options.Cache,
		onCacheError: options.OnCacheError,
	}, nil
}

func (s *DeleteService) DeleteByShortCode(ctx context.Context, shortCode string) error {
	if s == nil || s.repository == nil {
		return ErrRepositoryRequired
	}

	normalizedShortCode, err := shortcode.Normalize(shortCode)
	if err != nil {
		return err
	}

	if err := s.repository.DeleteByShortCode(ctx, normalizedShortCode); err != nil {
		return fmt.Errorf("delete URL by short code: %w", err)
	}

	invalidateRedirectCache(ctx, s.cache, normalizedShortCode, s.onCacheError)

	return nil
}
