package service

import (
	"context"
	"fmt"
	"time"

	"github.com/tapadar13/url-shortener/apps/api/internal/shortcode"
	urlmodel "github.com/tapadar13/url-shortener/apps/api/internal/url"
)

type AccessRepository interface {
	RecordAccess(ctx context.Context, params urlmodel.RecordAccessParams) (urlmodel.URL, error)
}

type RedirectOptions struct {
	Now func() time.Time
}

type RedirectService struct {
	repository AccessRepository
	now        func() time.Time
}

func NewRedirectService(repository AccessRepository, options RedirectOptions) (*RedirectService, error) {
	if repository == nil {
		return nil, ErrRepositoryRequired
	}

	if options.Now == nil {
		options.Now = time.Now
	}

	return &RedirectService{
		repository: repository,
		now:        options.Now,
	}, nil
}

func (s *RedirectService) Resolve(ctx context.Context, shortCode string) (urlmodel.URL, error) {
	if s == nil || s.repository == nil {
		return urlmodel.URL{}, ErrRepositoryRequired
	}

	if s.now == nil {
		s.now = time.Now
	}

	normalizedShortCode, err := shortcode.Normalize(shortCode)
	if err != nil {
		return urlmodel.URL{}, err
	}

	recorded, err := s.repository.RecordAccess(ctx, urlmodel.RecordAccessParams{
		ShortCode:  normalizedShortCode,
		AccessedAt: s.now().UTC(),
	})
	if err != nil {
		return urlmodel.URL{}, fmt.Errorf("record URL access: %w", err)
	}

	return recorded, nil
}
