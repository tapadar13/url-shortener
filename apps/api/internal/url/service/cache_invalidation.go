package service

import (
	"context"
	"fmt"
)

type RedirectCacheInvalidator interface {
	Delete(ctx context.Context, shortCode string) error
}

func invalidateRedirectCache(
	ctx context.Context,
	cache RedirectCacheInvalidator,
	shortCode string,
	onError func(error),
) {
	if cache == nil {
		return
	}

	if err := cache.Delete(ctx, shortCode); err != nil && onError != nil {
		onError(fmt.Errorf("invalidate redirect cache entry: %w", err))
	}
}
