package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	redisdriver "github.com/redis/go-redis/v9"
	"github.com/tapadar13/url-shortener/apps/api/internal/shortcode"
	urlmodel "github.com/tapadar13/url-shortener/apps/api/internal/url"
	urlcache "github.com/tapadar13/url-shortener/apps/api/internal/url/cache"
)

const (
	entryVersion = 1
	keySegment   = "redirect"
)

type commandClient interface {
	Get(ctx context.Context, key string) *redisdriver.StringCmd
	Set(ctx context.Context, key string, value any, expiration time.Duration) *redisdriver.StatusCmd
	Del(ctx context.Context, keys ...string) *redisdriver.IntCmd
}

type Repository struct {
	client    commandClient
	keyPrefix string
}

type entryDocument struct {
	Version int    `json:"version"`
	LongURL string `json:"longUrl"`
}

func New(client *redisdriver.Client, keyPrefix string) *Repository {
	if client == nil {
		return newRepository(nil, keyPrefix)
	}

	return newRepository(client, keyPrefix)
}

func newRepository(client commandClient, keyPrefix string) *Repository {
	return &Repository{
		client:    client,
		keyPrefix: strings.Trim(strings.TrimSpace(keyPrefix), ":"),
	}
}

func (r *Repository) Get(ctx context.Context, shortCode string) (urlcache.Entry, error) {
	key, err := r.key(shortCode)
	if err != nil {
		return urlcache.Entry{}, err
	}

	payload, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redisdriver.Nil) {
		return urlcache.Entry{}, urlcache.ErrMiss
	}
	if err != nil {
		return urlcache.Entry{}, fmt.Errorf("get redirect cache entry: %w", err)
	}

	var document entryDocument
	if err := json.Unmarshal([]byte(payload), &document); err != nil {
		return urlcache.Entry{}, fmt.Errorf("decode redirect cache entry: %w", err)
	}

	if document.Version != entryVersion {
		return urlcache.Entry{}, fmt.Errorf("decode redirect cache entry: unsupported version %d", document.Version)
	}

	longURL, err := urlmodel.NormalizeLongURL(document.LongURL)
	if err != nil {
		return urlcache.Entry{}, fmt.Errorf("validate redirect cache entry: %w", err)
	}

	return urlcache.Entry{LongURL: longURL}, nil
}

func (r *Repository) Set(ctx context.Context, shortCode string, entry urlcache.Entry, ttl time.Duration) error {
	key, err := r.key(shortCode)
	if err != nil {
		return err
	}

	if ttl <= 0 {
		return urlcache.ErrTTLInvalid
	}

	longURL, err := urlmodel.NormalizeLongURL(entry.LongURL)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(entryDocument{
		Version: entryVersion,
		LongURL: longURL,
	})
	if err != nil {
		return fmt.Errorf("encode redirect cache entry: %w", err)
	}

	if err := r.client.Set(ctx, key, payload, ttl).Err(); err != nil {
		return fmt.Errorf("set redirect cache entry: %w", err)
	}

	return nil
}

func (r *Repository) Delete(ctx context.Context, shortCode string) error {
	key, err := r.key(shortCode)
	if err != nil {
		return err
	}

	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("delete redirect cache entry: %w", err)
	}

	return nil
}

func (r *Repository) key(shortCode string) (string, error) {
	if r == nil || r.client == nil {
		return "", errors.New("Redis redirect cache client is required")
	}

	if r.keyPrefix == "" {
		return "", errors.New("Redis redirect cache key prefix is required")
	}

	normalized, err := shortcode.Normalize(shortCode)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s:%s:%s", r.keyPrefix, keySegment, normalized), nil
}

var _ urlcache.Store = (*Repository)(nil)
