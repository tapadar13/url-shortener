package redis

import (
	"context"
	"errors"
	"fmt"
	"strings"

	redisdriver "github.com/redis/go-redis/v9"
	"github.com/tapadar13/url-shortener/apps/api/internal/config"
)

type Client struct {
	driver    *redisdriver.Client
	keyPrefix string
}

func Connect(ctx context.Context, cfg config.RedisConfig) (*Client, error) {
	if cfg.ConnectTimeout <= 0 {
		return nil, errors.New("Redis connection timeout must be greater than zero")
	}

	keyPrefix := strings.TrimSpace(cfg.KeyPrefix)
	if keyPrefix == "" {
		return nil, errors.New("Redis key prefix is required")
	}

	options, err := redisdriver.ParseURL(strings.TrimSpace(cfg.URL))
	if err != nil {
		return nil, fmt.Errorf("parse Redis URL: %w", err)
	}
	options.DialTimeout = cfg.ConnectTimeout
	options.MaxRetries = -1
	options.ContextTimeoutEnabled = true

	driver := redisdriver.NewClient(options)
	pingCtx, cancel := context.WithTimeout(contextOrBackground(ctx), cfg.ConnectTimeout)
	defer cancel()

	if err := driver.Ping(pingCtx).Err(); err != nil {
		_ = driver.Close()
		return nil, fmt.Errorf("ping Redis: %w", err)
	}

	return &Client{
		driver:    driver,
		keyPrefix: keyPrefix,
	}, nil
}

func (c *Client) Close() error {
	if c == nil || c.driver == nil {
		return nil
	}

	if err := c.driver.Close(); err != nil {
		return fmt.Errorf("close Redis: %w", err)
	}

	return nil
}

func (c *Client) Ping(ctx context.Context) error {
	if c == nil || c.driver == nil {
		return errors.New("Redis client is required")
	}

	if err := c.driver.Ping(contextOrBackground(ctx)).Err(); err != nil {
		return fmt.Errorf("ping Redis: %w", err)
	}

	return nil
}

func (c *Client) Driver() *redisdriver.Client {
	if c == nil {
		return nil
	}

	return c.driver
}

func (c *Client) KeyPrefix() string {
	if c == nil {
		return ""
	}

	return c.keyPrefix
}

func contextOrBackground(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}

	return ctx
}
