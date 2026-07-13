package ratelimit

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNewRequiresCounterStore(t *testing.T) {
	t.Parallel()

	_, err := New(nil, Options{Window: time.Minute})
	if !errors.Is(err, ErrCounterStoreRequired) {
		t.Fatalf("expected counter store error, got %v", err)
	}
}

func TestNewRejectsInvalidOptions(t *testing.T) {
	t.Parallel()

	store := &fakeCounterStore{}

	_, err := New(store, Options{Requests: -1, Window: time.Minute})
	if !errors.Is(err, ErrRequestsInvalid) {
		t.Fatalf("expected request limit error, got %v", err)
	}

	_, err = New(store, Options{Window: 0})
	if !errors.Is(err, ErrWindowInvalid) {
		t.Fatalf("expected window error, got %v", err)
	}
}

func TestLimiterAllowsRequestWithinWindow(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 12, 10, 34, 45, 0, time.FixedZone("IST", 5*60*60+30*60))
	store := &fakeCounterStore{count: 2}
	limiter := newTestLimiter(t, store, Options{
		Requests: 3,
		Window:   time.Minute,
		Now:      func() time.Time { return now },
	})

	result, err := limiter.Allow(context.Background(), " 203.0.113.10 ")
	if err != nil {
		t.Fatalf("expected request to be allowed: %v", err)
	}

	if !result.Allowed || result.Limit != 3 || result.Remaining != 1 {
		t.Fatalf("expected allowed result with one request remaining, got %+v", result)
	}

	expectedStart := time.Date(2026, 7, 12, 5, 4, 0, 0, time.UTC)
	if store.params.ClientKey != "203.0.113.10" || !store.params.WindowStart.Equal(expectedStart) || !store.params.ExpiresAt.Equal(expectedStart.Add(time.Minute)) {
		t.Fatalf("expected fixed window params, got %+v", store.params)
	}

	if !result.ResetAt.Equal(expectedStart.Add(time.Minute)) {
		t.Fatalf("expected reset at %s, got %s", expectedStart.Add(time.Minute), result.ResetAt)
	}
}

func TestLimiterRejectsRequestOverLimit(t *testing.T) {
	t.Parallel()

	limiter := newTestLimiter(t, &fakeCounterStore{count: 4}, Options{
		Requests: 3,
		Window:   time.Minute,
		Now:      fixedTime,
	})

	result, err := limiter.Allow(context.Background(), "203.0.113.10")
	if err != nil {
		t.Fatalf("expected request result: %v", err)
	}

	if result.Allowed || result.Remaining != 0 {
		t.Fatalf("expected rejected result with no remaining requests, got %+v", result)
	}
}

func TestLimiterSkipsCounterWhenDisabled(t *testing.T) {
	t.Parallel()

	store := &fakeCounterStore{}
	limiter := newTestLimiter(t, store, Options{
		Window: time.Minute,
	})

	result, err := limiter.Allow(context.Background(), "")
	if err != nil {
		t.Fatalf("expected disabled limiter to allow request: %v", err)
	}

	if !result.Allowed || store.called {
		t.Fatalf("expected disabled limiter to skip the counter, got result=%+v called=%t", result, store.called)
	}
}

func TestLimiterRequiresClientKey(t *testing.T) {
	t.Parallel()

	limiter := newTestLimiter(t, &fakeCounterStore{}, Options{
		Requests: 1,
		Window:   time.Minute,
	})

	_, err := limiter.Allow(context.Background(), " ")
	if !errors.Is(err, ErrClientKeyRequired) {
		t.Fatalf("expected client key error, got %v", err)
	}
}

func TestLimiterReturnsCounterStoreError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("database unavailable")
	limiter := newTestLimiter(t, &fakeCounterStore{err: expectedErr}, Options{
		Requests: 1,
		Window:   time.Minute,
	})

	_, err := limiter.Allow(context.Background(), "203.0.113.10")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected counter store error, got %v", err)
	}
}

func newTestLimiter(t *testing.T, store CounterStore, options Options) *Limiter {
	t.Helper()

	limiter, err := New(store, options)
	if err != nil {
		t.Fatalf("expected limiter to be created: %v", err)
	}

	return limiter
}

func fixedTime() time.Time {
	return time.Date(2026, 7, 12, 5, 4, 45, 0, time.UTC)
}

type fakeCounterStore struct {
	count  int
	err    error
	called bool
	params IncrementParams
}

func (s *fakeCounterStore) Increment(_ context.Context, params IncrementParams) (int, error) {
	s.called = true
	s.params = params

	if s.err != nil {
		return 0, s.err
	}

	return s.count, nil
}
