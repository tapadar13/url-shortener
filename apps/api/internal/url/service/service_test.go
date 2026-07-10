package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/tapadar13/url-shortener/apps/api/internal/shortcode"
	urlmodel "github.com/tapadar13/url-shortener/apps/api/internal/url"
)

func TestNewRequiresRepository(t *testing.T) {
	t.Parallel()

	_, err := New(nil, &fakeGenerator{}, Options{})
	if !errors.Is(err, ErrRepositoryRequired) {
		t.Fatalf("expected repository error, got %v", err)
	}
}

func TestNewRequiresShortCodeGenerator(t *testing.T) {
	t.Parallel()

	_, err := New(&fakeRepository{}, nil, Options{})
	if !errors.Is(err, ErrShortCodeGeneratorMissing) {
		t.Fatalf("expected generator error, got %v", err)
	}
}

func TestServiceCreateRequiresConfiguredDependencies(t *testing.T) {
	t.Parallel()

	_, err := (&Service{}).Create(context.Background(), CreateParams{
		LongURL: "https://example.com",
	})
	if !errors.Is(err, ErrRepositoryRequired) {
		t.Fatalf("expected repository error, got %v", err)
	}

	_, err = (&Service{repository: &fakeRepository{}}).Create(context.Background(), CreateParams{
		LongURL: "https://example.com",
	})
	if !errors.Is(err, ErrShortCodeGeneratorMissing) {
		t.Fatalf("expected generator error, got %v", err)
	}
}

func TestServiceCreateCreatesURL(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 10, 8, 0, 0, 0, time.UTC)
	repository := &fakeRepository{}

	service := newTestService(t, repository, &fakeGenerator{
		codes: []string{"AbC1234"},
	}, Options{
		ShortCodeLength: 7,
		MaxRetries:      3,
		Now: func() time.Time {
			return now
		},
	})

	created, err := service.Create(context.Background(), CreateParams{
		LongURL: " https://example.com/articles/123 ",
	})
	if err != nil {
		t.Fatalf("expected URL to be created: %v", err)
	}

	if repository.createCount != 1 {
		t.Fatalf("expected one create call, got %d", repository.createCount)
	}

	if created.ShortCode != "AbC1234" {
		t.Fatalf("expected short code AbC1234, got %q", created.ShortCode)
	}

	if created.LongURL != "https://example.com/articles/123" {
		t.Fatalf("expected normalized URL, got %q", created.LongURL)
	}

	if !created.CreatedAt.Equal(now) || !created.UpdatedAt.Equal(now) {
		t.Fatalf("expected service timestamp %s, got created=%s updated=%s", now, created.CreatedAt, created.UpdatedAt)
	}
}

func TestServiceCreateRetriesDuplicateShortCodes(t *testing.T) {
	t.Parallel()

	repository := &fakeRepository{
		errors: []error{
			urlmodel.ErrDuplicateShortCode,
			nil,
		},
	}

	service := newTestService(t, repository, &fakeGenerator{
		codes: []string{"Dup1234", "OkC1234"},
	}, Options{
		ShortCodeLength: 7,
		MaxRetries:      3,
		Now:             fixedTime,
	})

	created, err := service.Create(context.Background(), CreateParams{
		LongURL: "https://example.com",
	})
	if err != nil {
		t.Fatalf("expected retry to succeed: %v", err)
	}

	if repository.createCount != 2 {
		t.Fatalf("expected two create attempts, got %d", repository.createCount)
	}

	if created.ShortCode != "OkC1234" {
		t.Fatalf("expected retried short code, got %q", created.ShortCode)
	}
}

func TestServiceCreateStopsAfterRetryExhaustion(t *testing.T) {
	t.Parallel()

	repository := &fakeRepository{
		errors: []error{
			urlmodel.ErrDuplicateShortCode,
			urlmodel.ErrDuplicateShortCode,
		},
	}

	service := newTestService(t, repository, &fakeGenerator{
		codes: []string{"Dup1234", "Dup5678"},
	}, Options{
		ShortCodeLength: 7,
		MaxRetries:      2,
		Now:             fixedTime,
	})

	_, err := service.Create(context.Background(), CreateParams{
		LongURL: "https://example.com",
	})
	if !errors.Is(err, ErrShortCodeRetriesExhausted) {
		t.Fatalf("expected retries exhausted error, got %v", err)
	}

	if repository.createCount != 2 {
		t.Fatalf("expected two create attempts, got %d", repository.createCount)
	}
}

func TestServiceCreateReturnsGeneratorError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("entropy unavailable")
	service := newTestService(t, &fakeRepository{}, &fakeGenerator{
		err: expectedErr,
	}, Options{
		Now: fixedTime,
	})

	_, err := service.Create(context.Background(), CreateParams{
		LongURL: "https://example.com",
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected generator error, got %v", err)
	}
}

func TestServiceCreateReturnsValidationError(t *testing.T) {
	t.Parallel()

	repository := &fakeRepository{}
	service := newTestService(t, repository, &fakeGenerator{
		codes: []string{"AbC1234"},
	}, Options{
		Now: fixedTime,
	})

	_, err := service.Create(context.Background(), CreateParams{
		LongURL: "not-a-url",
	})
	if !errors.Is(err, urlmodel.ErrLongURLSchemeUnsupported) {
		t.Fatalf("expected URL validation error, got %v", err)
	}

	if repository.createCount != 0 {
		t.Fatalf("expected no create calls, got %d", repository.createCount)
	}
}

func TestServiceCreateReturnsRepositoryError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("database unavailable")
	repository := &fakeRepository{
		errors: []error{expectedErr},
	}

	service := newTestService(t, repository, &fakeGenerator{
		codes: []string{"AbC1234"},
	}, Options{
		Now: fixedTime,
	})

	_, err := service.Create(context.Background(), CreateParams{
		LongURL: "https://example.com",
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected repository error, got %v", err)
	}
}

func TestDefaultGeneratorProducesValidGenerator(t *testing.T) {
	t.Parallel()

	code, err := DefaultGenerator().Generate(DefaultShortCodeLength)
	if err != nil {
		t.Fatalf("expected code to be generated: %v", err)
	}

	if err := shortcode.Validate(code); err != nil {
		t.Fatalf("expected generated code to be valid: %v", err)
	}
}

func newTestService(t *testing.T, repository CreateRepository, generator ShortCodeGenerator, options Options) *Service {
	t.Helper()

	service, err := New(repository, generator, options)
	if err != nil {
		t.Fatalf("expected service to be created: %v", err)
	}

	return service
}

func fixedTime() time.Time {
	return time.Date(2026, 7, 10, 8, 0, 0, 0, time.UTC)
}

type fakeGenerator struct {
	codes []string
	err   error
	index int
}

func (g *fakeGenerator) Generate(_ int) (string, error) {
	if g.err != nil {
		return "", g.err
	}

	if g.index >= len(g.codes) {
		return "", errors.New("no more fake codes")
	}

	code := g.codes[g.index]
	g.index++

	return code, nil
}

type fakeRepository struct {
	createCount int
	errors      []error
}

func (r *fakeRepository) Create(_ context.Context, record urlmodel.URL) (urlmodel.URL, error) {
	r.createCount++

	var err error
	if len(r.errors) >= r.createCount {
		err = r.errors[r.createCount-1]
	}

	if err != nil {
		return urlmodel.URL{}, err
	}

	record.ID = "507f1f77bcf86cd799439011"
	return record, nil
}
