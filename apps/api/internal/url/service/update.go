package service

import (
	"context"
	"fmt"
	"time"

	"github.com/tapadar13/url-shortener/apps/api/internal/shortcode"
	urlmodel "github.com/tapadar13/url-shortener/apps/api/internal/url"
)

type UpdateRepository interface {
	UpdateLongURL(ctx context.Context, params urlmodel.UpdateLongURLParams) (urlmodel.URL, error)
}

type UpdateOptions struct {
	Cache        RedirectCacheInvalidator
	OnCacheError func(error)
	Now          func() time.Time
}

type UpdateService struct {
	repository   UpdateRepository
	cache        RedirectCacheInvalidator
	onCacheError func(error)
	now          func() time.Time
}

type UpdateParams struct {
	ShortCode string
	LongURL   string
}

func NewUpdateService(repository UpdateRepository, options UpdateOptions) (*UpdateService, error) {
	if repository == nil {
		return nil, ErrRepositoryRequired
	}

	if options.Now == nil {
		options.Now = time.Now
	}

	return &UpdateService{
		repository:   repository,
		cache:        options.Cache,
		onCacheError: options.OnCacheError,
		now:          options.Now,
	}, nil
}

func (s *UpdateService) UpdateLongURL(ctx context.Context, params UpdateParams) (urlmodel.URL, error) {
	if s == nil || s.repository == nil {
		return urlmodel.URL{}, ErrRepositoryRequired
	}

	if s.now == nil {
		s.now = time.Now
	}

	normalizedShortCode, err := shortcode.Normalize(params.ShortCode)
	if err != nil {
		return urlmodel.URL{}, err
	}

	normalizedLongURL, err := urlmodel.NormalizeLongURL(params.LongURL)
	if err != nil {
		return urlmodel.URL{}, err
	}

	updated, err := s.repository.UpdateLongURL(ctx, urlmodel.UpdateLongURLParams{
		ShortCode: normalizedShortCode,
		LongURL:   normalizedLongURL,
		UpdatedAt: s.now().UTC(),
	})
	if err != nil {
		return urlmodel.URL{}, fmt.Errorf("update URL: %w", err)
	}

	invalidateRedirectCache(ctx, s.cache, normalizedShortCode, s.onCacheError)

	return updated, nil
}
