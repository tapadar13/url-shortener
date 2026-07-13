//go:build integration

package integration_test

import (
	"context"
	"os"
	"testing"

	"github.com/tapadar13/url-shortener/apps/api/internal/config"
	redisplatform "github.com/tapadar13/url-shortener/apps/api/internal/platform/redis"
)

func newIntegrationRedisClient(t *testing.T, keyPrefix string) *redisplatform.Client {
	t.Helper()

	redisURL := os.Getenv("REDIS_INTEGRATION_URL")
	if redisURL == "" {
		t.Skip("set REDIS_INTEGRATION_URL to run Redis integration tests")
	}

	ctx, cancel := context.WithTimeout(context.Background(), integrationTimeout)
	defer cancel()

	client, err := redisplatform.Connect(ctx, config.RedisConfig{
		URL:            redisURL,
		KeyPrefix:      keyPrefix,
		ConnectTimeout: integrationTimeout,
	})
	if err != nil {
		t.Fatalf("connect Redis: %v", err)
	}

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), integrationTimeout)
		defer cleanupCancel()

		if err := deleteRedisKeysWithPrefix(cleanupCtx, client, keyPrefix); err != nil {
			t.Errorf("delete integration cache keys: %v", err)
		}

		if err := client.Close(); err != nil {
			t.Errorf("close Redis: %v", err)
		}
	})

	return client
}

func deleteRedisKeysWithPrefix(ctx context.Context, client *redisplatform.Client, keyPrefix string) error {
	if client == nil || client.Driver() == nil {
		return nil
	}

	iterator := client.Driver().Scan(ctx, 0, keyPrefix+":*", 0).Iterator()
	var keys []string
	for iterator.Next(ctx) {
		keys = append(keys, iterator.Val())
	}
	if err := iterator.Err(); err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	return client.Driver().Del(ctx, keys...).Err()
}
