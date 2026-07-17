//go:build integration

package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/tapadar13/url-shortener/apps/api/internal/ratelimit"
	ratelimitrepository "github.com/tapadar13/url-shortener/apps/api/internal/ratelimit/repository/mongodb"
)

func TestRateLimitRepositoryConcurrentIncrements(t *testing.T) {
	client := newIntegrationMongoClient(t)
	repository := ratelimitrepository.New(client.RateLimitsCollection())
	ctx, cancel := context.WithTimeout(context.Background(), integrationTimeout)
	defer cancel()

	const incrementCount = 32
	windowStart := time.Now().UTC().Truncate(time.Minute)
	params := ratelimit.IncrementParams{
		ClientKey:   "203.0.113.10",
		WindowStart: windowStart,
		ExpiresAt:   windowStart.Add(time.Minute),
	}

	type incrementResult struct {
		count int
		err   error
	}

	results := make(chan incrementResult, incrementCount)
	for range incrementCount {
		go func() {
			count, err := repository.Increment(ctx, params)
			results <- incrementResult{count: count, err: err}
		}()
	}

	seen := make(map[int]bool, incrementCount)
	for range incrementCount {
		result := <-results
		if result.err != nil {
			t.Fatalf("increment rate limit counter: %v", result.err)
		}

		seen[result.count] = true
	}

	for expected := 1; expected <= incrementCount; expected++ {
		if !seen[expected] {
			t.Fatalf("expected atomic counts 1 through %d, missing %d; got %#v", incrementCount, expected, seen)
		}
	}
}
