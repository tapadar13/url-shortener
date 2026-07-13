package cache

import (
	"context"
	"errors"
	"time"
)

var (
	ErrMiss       = errors.New("redirect cache miss")
	ErrTTLInvalid = errors.New("redirect cache TTL must be greater than zero")
)

type Entry struct {
	LongURL string
}

type Store interface {
	Get(ctx context.Context, shortCode string) (Entry, error)
	Set(ctx context.Context, shortCode string, entry Entry, ttl time.Duration) error
	Delete(ctx context.Context, shortCode string) error
}
