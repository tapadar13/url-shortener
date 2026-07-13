package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrCounterStoreRequired = errors.New("rate limit counter store is required")
	ErrRequestsInvalid      = errors.New("rate limit requests must be zero or greater")
	ErrWindowInvalid        = errors.New("rate limit window must be greater than zero")
	ErrClientKeyRequired    = errors.New("rate limit client key is required")
)

type CounterStore interface {
	Increment(ctx context.Context, params IncrementParams) (int, error)
}

type IncrementParams struct {
	ClientKey   string
	WindowStart time.Time
	ExpiresAt   time.Time
}

type Options struct {
	Requests int
	Window   time.Duration
	Now      func() time.Time
}

type Result struct {
	Allowed   bool
	Limit     int
	Remaining int
	ResetAt   time.Time
}

type Limiter struct {
	store    CounterStore
	requests int
	window   time.Duration
	now      func() time.Time
}

func New(store CounterStore, options Options) (*Limiter, error) {
	if store == nil {
		return nil, ErrCounterStoreRequired
	}

	if options.Requests < 0 {
		return nil, ErrRequestsInvalid
	}

	if options.Window <= 0 {
		return nil, ErrWindowInvalid
	}

	if options.Now == nil {
		options.Now = time.Now
	}

	return &Limiter{
		store:    store,
		requests: options.Requests,
		window:   options.Window,
		now:      options.Now,
	}, nil
}

func (l *Limiter) Allow(ctx context.Context, clientKey string) (Result, error) {
	if l == nil || l.store == nil {
		return Result{}, ErrCounterStoreRequired
	}

	if l.requests == 0 {
		return Result{Allowed: true}, nil
	}

	key := strings.TrimSpace(clientKey)
	if key == "" {
		return Result{}, ErrClientKeyRequired
	}

	if l.window <= 0 {
		return Result{}, ErrWindowInvalid
	}

	if l.now == nil {
		l.now = time.Now
	}

	now := l.now().UTC()
	windowStart := now.Truncate(l.window)
	resetAt := windowStart.Add(l.window)
	count, err := l.store.Increment(ctx, IncrementParams{
		ClientKey:   key,
		WindowStart: windowStart,
		ExpiresAt:   resetAt,
	})
	if err != nil {
		return Result{}, fmt.Errorf("increment rate limit counter: %w", err)
	}

	remaining := l.requests - count
	if remaining < 0 {
		remaining = 0
	}

	return Result{
		Allowed:   count <= l.requests,
		Limit:     l.requests,
		Remaining: remaining,
		ResetAt:   resetAt,
	}, nil
}
