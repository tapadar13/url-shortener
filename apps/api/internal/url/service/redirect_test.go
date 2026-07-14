package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/tapadar13/url-shortener/apps/api/internal/shortcode"
	urlmodel "github.com/tapadar13/url-shortener/apps/api/internal/url"
	urlcache "github.com/tapadar13/url-shortener/apps/api/internal/url/cache"
)

func TestNewRedirectServiceRequiresRepository(t *testing.T) {
	t.Parallel()

	_, err := NewRedirectService(nil, RedirectOptions{})
	if !errors.Is(err, ErrRepositoryRequired) {
		t.Fatalf("expected repository error, got %v", err)
	}
}

func TestNewRedirectServiceValidatesCacheDependencies(t *testing.T) {
	t.Parallel()

	repository := &fakeAccessRepository{}
	cache := &fakeRedirectCache{}
	recorder := &fakeAccessEnqueuer{}

	_, err := NewRedirectService(repository, RedirectOptions{Recorder: recorder})
	if !errors.Is(err, ErrRedirectCacheRequired) {
		t.Fatalf("expected redirect cache error, got %v", err)
	}

	_, err = NewRedirectService(repository, RedirectOptions{Cache: cache, CacheTTL: time.Minute})
	if !errors.Is(err, ErrAccessRecorderRequired) {
		t.Fatalf("expected access recorder error, got %v", err)
	}

	_, err = NewRedirectService(repository, RedirectOptions{Cache: cache, Recorder: recorder})
	if !errors.Is(err, urlcache.ErrTTLInvalid) {
		t.Fatalf("expected cache TTL error, got %v", err)
	}
}

func TestRedirectServiceRecordsAccessForNormalizedShortCode(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 11, 8, 0, 0, 0, time.FixedZone("IST", 5*60*60+30*60))
	repository := &fakeAccessRepository{
		record: urlmodel.URL{
			ID:        "507f1f77bcf86cd799439011",
			LongURL:   "https://example.com/articles/123",
			ShortCode: "AbC123",
		},
	}
	service, err := NewRedirectService(repository, RedirectOptions{
		Now: func() time.Time {
			return now
		},
	})
	if err != nil {
		t.Fatalf("expected redirect service: %v", err)
	}

	resolved, err := service.Resolve(context.Background(), " AbC123 ")
	if err != nil {
		t.Fatalf("expected short URL to resolve: %v", err)
	}

	if repository.params.ShortCode != "AbC123" {
		t.Fatalf("expected normalized short code, got %q", repository.params.ShortCode)
	}

	if !repository.params.AccessedAt.Equal(now.UTC()) {
		t.Fatalf("expected access timestamp %s, got %s", now.UTC(), repository.params.AccessedAt)
	}

	if resolved != repository.record {
		t.Fatalf("expected resolved URL %#v, got %#v", repository.record, resolved)
	}
}

func TestRedirectServiceRejectsInvalidShortCodeBeforeRepository(t *testing.T) {
	t.Parallel()

	repository := &fakeAccessRepository{}
	service, err := NewRedirectService(repository, RedirectOptions{})
	if err != nil {
		t.Fatalf("expected redirect service: %v", err)
	}

	_, err = service.Resolve(context.Background(), "invalid-code")
	if !errors.Is(err, shortcode.ErrInvalidChars) {
		t.Fatalf("expected short code validation error, got %v", err)
	}

	if repository.called {
		t.Fatal("expected repository not to be called")
	}
}

func TestRedirectServiceReturnsRepositoryError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("database unavailable")
	service, err := NewRedirectService(&fakeAccessRepository{err: expectedErr}, RedirectOptions{})
	if err != nil {
		t.Fatalf("expected redirect service: %v", err)
	}

	_, err = service.Resolve(context.Background(), "AbC123")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected repository error, got %v", err)
	}
}

func TestRedirectServiceQueuesAnalyticsAfterDatabaseRedirect(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 14, 10, 0, 0, 0, time.UTC)
	analyticsRecorder := &fakeAnalyticsEnqueuer{}
	service, err := NewRedirectService(&fakeAccessRepository{record: urlmodel.URL{
		LongURL:   "https://example.com/database",
		ShortCode: "AbC123",
	}}, RedirectOptions{
		Analytics: analyticsRecorder,
		Now:       func() time.Time { return now },
	})
	if err != nil {
		t.Fatalf("expected redirect service: %v", err)
	}

	if _, err := service.Resolve(context.Background(), " AbC123 "); err != nil {
		t.Fatalf("expected database redirect: %v", err)
	}

	if analyticsRecorder.calls != 1 || analyticsRecorder.shortCode != "AbC123" || !analyticsRecorder.clickedAt.Equal(now) {
		t.Fatalf("expected one analytics click, got %+v", analyticsRecorder)
	}
}

func TestRedirectServiceQueuesAnalyticsAfterCachedRedirect(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 14, 10, 0, 0, 0, time.UTC)
	analyticsRecorder := &fakeAnalyticsEnqueuer{}
	service, err := NewRedirectService(&fakeAccessRepository{}, RedirectOptions{
		Cache:     &fakeRedirectCache{entry: urlcache.Entry{LongURL: "https://example.com/cached"}},
		CacheTTL:  time.Minute,
		Recorder:  &fakeAccessEnqueuer{},
		Analytics: analyticsRecorder,
		Now:       func() time.Time { return now },
	})
	if err != nil {
		t.Fatalf("expected cached redirect service: %v", err)
	}

	if _, err := service.Resolve(context.Background(), "AbC123"); err != nil {
		t.Fatalf("expected cached redirect: %v", err)
	}

	if analyticsRecorder.calls != 1 || analyticsRecorder.shortCode != "AbC123" || !analyticsRecorder.clickedAt.Equal(now) {
		t.Fatalf("expected one analytics click, got %+v", analyticsRecorder)
	}
}

func TestRedirectServiceFailsOpenWhenAnalyticsQueueRejectsClick(t *testing.T) {
	t.Parallel()

	analyticsErr := errors.New("analytics queue full")
	reported := make(chan error, 1)
	service, err := NewRedirectService(&fakeAccessRepository{record: urlmodel.URL{
		LongURL:   "https://example.com/database",
		ShortCode: "AbC123",
	}}, RedirectOptions{
		Analytics: &fakeAnalyticsEnqueuer{err: analyticsErr},
		OnAnalyticsError: func(err error) {
			reported <- err
		},
	})
	if err != nil {
		t.Fatalf("expected redirect service: %v", err)
	}

	if _, err := service.Resolve(context.Background(), "AbC123"); err != nil {
		t.Fatalf("expected redirect despite analytics error: %v", err)
	}

	select {
	case err := <-reported:
		if !errors.Is(err, analyticsErr) {
			t.Fatalf("expected reported analytics error, got %v", err)
		}
	default:
		t.Fatal("expected analytics error to be reported")
	}
}

func TestRedirectServiceSkipsAnalyticsWhenRedirectFails(t *testing.T) {
	t.Parallel()

	analyticsRecorder := &fakeAnalyticsEnqueuer{}
	service, err := NewRedirectService(&fakeAccessRepository{err: errors.New("database unavailable")}, RedirectOptions{
		Analytics: analyticsRecorder,
	})
	if err != nil {
		t.Fatalf("expected redirect service: %v", err)
	}

	if _, err := service.Resolve(context.Background(), "AbC123"); err == nil {
		t.Fatal("expected redirect error")
	}

	if analyticsRecorder.calls != 0 {
		t.Fatalf("expected no analytics click, got %d", analyticsRecorder.calls)
	}
}

func TestRedirectServiceReturnsCacheHitAndQueuesAccess(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 14, 10, 0, 0, 0, time.UTC)
	repository := &fakeAccessRepository{}
	cache := &fakeRedirectCache{entry: urlcache.Entry{LongURL: "https://example.com/cached"}}
	recorder := &fakeAccessEnqueuer{}
	service := newCachedRedirectService(t, repository, cache, recorder, now, 10*time.Minute, nil)

	resolved, err := service.Resolve(context.Background(), " AbC123 ")
	if err != nil {
		t.Fatalf("expected cached redirect: %v", err)
	}

	if resolved.LongURL != "https://example.com/cached" || resolved.ShortCode != "AbC123" {
		t.Fatalf("expected cached destination, got %+v", resolved)
	}

	if recorder.shortCode != "AbC123" || !recorder.accessedAt.Equal(now) {
		t.Fatalf("expected queued access, got code=%q time=%s", recorder.shortCode, recorder.accessedAt)
	}

	if repository.called {
		t.Fatal("expected cache hit not to call repository synchronously")
	}
}

func TestRedirectServiceFallsBackToSynchronousRecordingUnderBackpressure(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 14, 10, 0, 0, 0, time.UTC)
	repository := &fakeAccessRepository{record: urlmodel.URL{
		LongURL:   "https://example.com/database",
		ShortCode: "AbC123",
	}}
	cache := &fakeRedirectCache{entry: urlcache.Entry{LongURL: "https://example.com/cached"}}
	recorder := &fakeAccessEnqueuer{err: ErrAccessQueueFull}
	service := newCachedRedirectService(t, repository, cache, recorder, now, time.Minute, nil)

	resolved, err := service.Resolve(context.Background(), "AbC123")
	if err != nil {
		t.Fatalf("expected synchronous fallback: %v", err)
	}

	if resolved != repository.record || !repository.called {
		t.Fatalf("expected repository result, got resolved=%+v called=%t", resolved, repository.called)
	}
}

func TestRedirectServicePopulatesCacheAfterMiss(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 14, 10, 0, 0, 0, time.UTC)
	repository := &fakeAccessRepository{record: urlmodel.URL{
		LongURL:   "https://example.com/database",
		ShortCode: "AbC123",
	}}
	cache := &fakeRedirectCache{getErr: urlcache.ErrMiss}
	service := newCachedRedirectService(t, repository, cache, &fakeAccessEnqueuer{}, now, 10*time.Minute, nil)

	resolved, err := service.Resolve(context.Background(), "AbC123")
	if err != nil {
		t.Fatalf("expected redirect after cache miss: %v", err)
	}

	if resolved != repository.record {
		t.Fatalf("expected repository result, got %+v", resolved)
	}

	if cache.setShortCode != "AbC123" || cache.setEntry.LongURL != repository.record.LongURL || cache.setTTL != 10*time.Minute {
		t.Fatalf("expected cache population, got code=%q entry=%+v ttl=%s", cache.setShortCode, cache.setEntry, cache.setTTL)
	}
}

func TestRedirectServiceCapsCacheTTLAtLinkExpiration(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 14, 10, 0, 0, 0, time.UTC)
	expiresAt := now.Add(90 * time.Second)
	repository := &fakeAccessRepository{record: urlmodel.URL{
		LongURL:   "https://example.com/database",
		ShortCode: "AbC123",
		ExpiresAt: &expiresAt,
	}}
	cache := &fakeRedirectCache{getErr: urlcache.ErrMiss}
	service := newCachedRedirectService(t, repository, cache, &fakeAccessEnqueuer{}, now, 10*time.Minute, nil)

	if _, err := service.Resolve(context.Background(), "AbC123"); err != nil {
		t.Fatalf("expected redirect: %v", err)
	}

	if cache.setTTL != 90*time.Second {
		t.Fatalf("expected TTL capped at expiration, got %s", cache.setTTL)
	}
}

func TestRedirectServiceDoesNotCacheAlreadyExpiredRecord(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 14, 10, 0, 0, 0, time.UTC)
	expiresAt := now
	repository := &fakeAccessRepository{record: urlmodel.URL{
		LongURL:   "https://example.com/database",
		ShortCode: "AbC123",
		ExpiresAt: &expiresAt,
	}}
	cache := &fakeRedirectCache{getErr: urlcache.ErrMiss}
	service := newCachedRedirectService(t, repository, cache, &fakeAccessEnqueuer{}, now, time.Minute, nil)

	if _, err := service.Resolve(context.Background(), "AbC123"); err != nil {
		t.Fatalf("expected repository result: %v", err)
	}

	if cache.setCalled {
		t.Fatal("expected expired record not to be cached")
	}
}

func TestRedirectServiceFailsOpenOnCacheReadError(t *testing.T) {
	t.Parallel()

	cacheErr := errors.New("Redis unavailable")
	reported := make(chan error, 1)
	repository := &fakeAccessRepository{record: urlmodel.URL{LongURL: "https://example.com/database", ShortCode: "AbC123"}}
	cache := &fakeRedirectCache{getErr: cacheErr}
	service := newCachedRedirectService(t, repository, cache, &fakeAccessEnqueuer{}, time.Now(), time.Minute, func(err error) {
		reported <- err
	})

	resolved, err := service.Resolve(context.Background(), "AbC123")
	if err != nil {
		t.Fatalf("expected database fallback: %v", err)
	}

	if resolved != repository.record {
		t.Fatalf("expected repository result, got %+v", resolved)
	}

	select {
	case err := <-reported:
		if !errors.Is(err, cacheErr) {
			t.Fatalf("expected reported cache read error, got %v", err)
		}
	default:
		t.Fatal("expected cache read error to be reported")
	}
}

func TestRedirectServiceFailsOpenOnCacheWriteError(t *testing.T) {
	t.Parallel()

	cacheErr := errors.New("Redis unavailable")
	reported := make(chan error, 1)
	repository := &fakeAccessRepository{record: urlmodel.URL{LongURL: "https://example.com/database", ShortCode: "AbC123"}}
	cache := &fakeRedirectCache{getErr: urlcache.ErrMiss, setErr: cacheErr}
	service := newCachedRedirectService(t, repository, cache, &fakeAccessEnqueuer{}, time.Now(), time.Minute, func(err error) {
		reported <- err
	})

	if _, err := service.Resolve(context.Background(), "AbC123"); err != nil {
		t.Fatalf("expected redirect despite cache write error: %v", err)
	}

	select {
	case err := <-reported:
		if !errors.Is(err, cacheErr) {
			t.Fatalf("expected reported cache write error, got %v", err)
		}
	default:
		t.Fatal("expected cache write error to be reported")
	}
}

func TestRedirectServiceRequiresConfiguredRepository(t *testing.T) {
	t.Parallel()

	_, err := (&RedirectService{}).Resolve(context.Background(), "AbC123")
	if !errors.Is(err, ErrRepositoryRequired) {
		t.Fatalf("expected repository error, got %v", err)
	}
}

type fakeAccessRepository struct {
	record urlmodel.URL
	err    error
	params urlmodel.RecordAccessParams
	called bool
}

func (r *fakeAccessRepository) RecordAccess(_ context.Context, params urlmodel.RecordAccessParams) (urlmodel.URL, error) {
	r.called = true
	r.params = params

	if r.err != nil {
		return urlmodel.URL{}, r.err
	}

	return r.record, nil
}

func newCachedRedirectService(
	t *testing.T,
	repository AccessRepository,
	cache urlcache.Store,
	recorder AccessEnqueuer,
	now time.Time,
	ttl time.Duration,
	onCacheError func(error),
) *RedirectService {
	t.Helper()

	service, err := NewRedirectService(repository, RedirectOptions{
		Cache:        cache,
		CacheTTL:     ttl,
		Recorder:     recorder,
		OnCacheError: onCacheError,
		Now:          func() time.Time { return now },
	})
	if err != nil {
		t.Fatalf("create cached redirect service: %v", err)
	}

	return service
}

type fakeAccessEnqueuer struct {
	shortCode  string
	accessedAt time.Time
	err        error
}

type fakeAnalyticsEnqueuer struct {
	calls     int
	shortCode string
	clickedAt time.Time
	err       error
}

func (r *fakeAnalyticsEnqueuer) Enqueue(shortCode string, clickedAt time.Time) error {
	r.calls++
	r.shortCode = shortCode
	r.clickedAt = clickedAt
	return r.err
}

func (r *fakeAccessEnqueuer) Enqueue(shortCode string, accessedAt time.Time) error {
	r.shortCode = shortCode
	r.accessedAt = accessedAt
	return r.err
}

type fakeRedirectCache struct {
	entry        urlcache.Entry
	getErr       error
	setErr       error
	deleteErr    error
	getShortCode string
	setShortCode string
	setEntry     urlcache.Entry
	setTTL       time.Duration
	deleteCode   string
	setCalled    bool
}

func (c *fakeRedirectCache) Get(_ context.Context, shortCode string) (urlcache.Entry, error) {
	c.getShortCode = shortCode
	return c.entry, c.getErr
}

func (c *fakeRedirectCache) Set(_ context.Context, shortCode string, entry urlcache.Entry, ttl time.Duration) error {
	c.setCalled = true
	c.setShortCode = shortCode
	c.setEntry = entry
	c.setTTL = ttl
	return c.setErr
}

func (c *fakeRedirectCache) Delete(_ context.Context, shortCode string) error {
	c.deleteCode = shortCode
	return c.deleteErr
}
