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
	UserEmailIndexName           = "uniq_users_email"
	CreatedAtIndexName           = "idx_urls_created_at"
	ExpirationIndexName          = "ttl_urls_expires_at"
	RateLimitExpirationIndexName = "ttl_rate_limits_expires_at"
	AnalyticsDailyIndexName      = "uniq_analytics_short_code_day_start"
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

	usersCollection := client.UsersCollection()
	if usersCollection == nil {
		return errors.New("users collection is required")
	}
	if _, err := usersCollection.Indexes().CreateMany(ctx, UserIndexModels()); err != nil {
		return fmt.Errorf("create user indexes: %w", err)
	}

	rateLimitsCollection := client.RateLimitsCollection()
	if rateLimitsCollection == nil {
		return errors.New("rate limits collection is required")
	}

	if _, err := rateLimitsCollection.Indexes().CreateMany(ctx, RateLimitIndexModels()); err != nil {
		return fmt.Errorf("create rate limit indexes: %w", err)
	}

	analyticsCollection := client.AnalyticsCollection()
	if analyticsCollection == nil {
		return errors.New("analytics collection is required")
	}

	if _, err := analyticsCollection.Indexes().CreateMany(ctx, AnalyticsIndexModels()); err != nil {
		return fmt.Errorf("create analytics indexes: %w", err)
	}

	return nil
}

func UserIndexModels() []mongo.IndexModel {
	return []mongo.IndexModel{{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetName(UserEmailIndexName).SetUnique(true),
	}}
}

func AnalyticsIndexModels() []mongo.IndexModel {
	return []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "short_code", Value: 1},
				{Key: "day_start", Value: 1},
			},
			Options: options.Index().
				SetName(AnalyticsDailyIndexName).
				SetUnique(true),
		},
	}
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
