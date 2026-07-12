package mongodb

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	ShortCodeIndexName           = "uniq_urls_short_code"
	CreatedAtIndexName           = "idx_urls_created_at"
	ExpirationIndexName          = "ttl_urls_expires_at"
	RateLimitExpirationIndexName = "ttl_rate_limits_expires_at"
)

func EnsureIndexes(ctx context.Context, client *Client) error {
	if client == nil {
		return errors.New("MongoDB client is required")
	}

	urlsCollection := client.URLsCollection()
	if urlsCollection == nil {
		return errors.New("URLs collection is required")
	}

	if _, err := urlsCollection.Indexes().CreateMany(ctx, URLIndexModels()); err != nil {
		return fmt.Errorf("create URL indexes: %w", err)
	}

	rateLimitsCollection := client.RateLimitsCollection()
	if rateLimitsCollection == nil {
		return errors.New("rate limits collection is required")
	}

	if _, err := rateLimitsCollection.Indexes().CreateMany(ctx, RateLimitIndexModels()); err != nil {
		return fmt.Errorf("create rate limit indexes: %w", err)
	}

	return nil
}

func RateLimitIndexModels() []mongo.IndexModel {
	return []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "expires_at", Value: 1}},
			Options: options.Index().
				SetName(RateLimitExpirationIndexName).
				SetExpireAfterSeconds(0),
		},
	}
}

func URLIndexModels() []mongo.IndexModel {
	return []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "short_code", Value: 1}},
			Options: options.Index().
				SetName(ShortCodeIndexName).
				SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "created_at", Value: -1}},
			Options: options.Index().
				SetName(CreatedAtIndexName),
		},
		{
			Keys: bson.D{{Key: "expires_at", Value: 1}},
			Options: options.Index().
				SetName(ExpirationIndexName).
				SetExpireAfterSeconds(0),
		},
	}
}
