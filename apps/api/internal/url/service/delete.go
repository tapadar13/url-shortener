package service

import (
	"context"
	"fmt"

	"github.com/tapadar13/url-shortener/apps/api/internal/shortcode"
)

type DeleteRepository interface {
	DeleteByShortCode(ctx context.Context, shortCode string) error
}

type DeleteService struct {
	repository DeleteRepository
}

func NewDeleteService(repository DeleteRepository) (*DeleteService, error) {
	if repository == nil {
		return nil, ErrRepositoryRequired
	}

	return &DeleteService{
		repository: repository,
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

	return nil
}
