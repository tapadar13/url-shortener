package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/tapadar13/url-shortener/apps/api/internal/shortcode"
	urlmodel "github.com/tapadar13/url-shortener/apps/api/internal/url"
)

const (
	DefaultShortCodeLength = 7
	DefaultMaxRetries      = 5
)

var (
	ErrRepositoryRequired        = errors.New("URL repository is required")
	ErrShortCodeGeneratorMissing = errors.New("short code generator is required")
	ErrShortCodeRetriesExhausted = errors.New("short code generation retries exhausted")
)

type ShortCodeGenerator interface {
	Generate(length int) (string, error)
}

type CreateRepository interface {
	Create(ctx context.Context, record urlmodel.URL) (urlmodel.URL, error)
}

type Options struct {
	ShortCodeLength int
	MaxRetries      int
	Now             func() time.Time
}

type Service struct {
	repository      CreateRepository
	generator       ShortCodeGenerator
	shortCodeLength int
	maxRetries      int
	now             func() time.Time
}

type CreateParams struct {
	LongURL   string
	ExpiresAt *time.Time
	ShortCode *string
}

func New(repository CreateRepository, generator ShortCodeGenerator, options Options) (*Service, error) {
	if repository == nil {
		return nil, ErrRepositoryRequired
	}

	if generator == nil {
		return nil, ErrShortCodeGeneratorMissing
	}

	if options.ShortCodeLength == 0 {
		options.ShortCodeLength = DefaultShortCodeLength
	}

	if options.MaxRetries == 0 {
		options.MaxRetries = DefaultMaxRetries
	}

	if options.Now == nil {
		options.Now = time.Now
	}

	return &Service{
		repository:      repository,
		generator:       generator,
		shortCodeLength: options.ShortCodeLength,
		maxRetries:      options.MaxRetries,
		now:             options.Now,
	}, nil
}

func (s *Service) Create(ctx context.Context, params CreateParams) (urlmodel.URL, error) {
	if s == nil {
		return urlmodel.URL{}, ErrRepositoryRequired
	}

	if s.repository == nil {
		return urlmodel.URL{}, ErrRepositoryRequired
	}

	if s.generator == nil {
		return urlmodel.URL{}, ErrShortCodeGeneratorMissing
	}

	if s.now == nil {
		s.now = time.Now
	}

	if params.ShortCode != nil {
		record, err := urlmodel.New(urlmodel.NewParams{
			LongURL:   params.LongURL,
			ShortCode: *params.ShortCode,
			Now:       s.now(),
			ExpiresAt: params.ExpiresAt,
		})
		if err != nil {
			return urlmodel.URL{}, err
		}

		created, err := s.repository.Create(ctx, record)
		if err != nil {
			return urlmodel.URL{}, fmt.Errorf("create URL: %w", err)
		}

		return created, nil
	}

	if s.maxRetries < 1 {
		return urlmodel.URL{}, ErrShortCodeRetriesExhausted
	}

	var lastDuplicate error
	for attempt := 0; attempt < s.maxRetries; attempt++ {
		shortCode, err := s.generator.Generate(s.shortCodeLength)
		if err != nil {
			return urlmodel.URL{}, fmt.Errorf("generate short code: %w", err)
		}

		record, err := urlmodel.New(urlmodel.NewParams{
			LongURL:   params.LongURL,
			ShortCode: shortCode,
			Now:       s.now(),
			ExpiresAt: params.ExpiresAt,
		})
		if err != nil {
			return urlmodel.URL{}, err
		}

		created, err := s.repository.Create(ctx, record)
		if errors.Is(err, urlmodel.ErrDuplicateShortCode) {
			lastDuplicate = err
			continue
		}

		if err != nil {
			return urlmodel.URL{}, fmt.Errorf("create URL: %w", err)
		}

		return created, nil
	}

	if lastDuplicate != nil {
		return urlmodel.URL{}, fmt.Errorf("%w: %w", ErrShortCodeRetriesExhausted, lastDuplicate)
	}

	return urlmodel.URL{}, ErrShortCodeRetriesExhausted
}

func DefaultGenerator() shortcode.Generator {
	return shortcode.NewGenerator()
}
