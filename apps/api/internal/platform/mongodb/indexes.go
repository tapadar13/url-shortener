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
	ShortCodeIndexName = "uniq_urls_short_code"
	CreatedAtIndexName = "idx_urls_created_at"
)

func EnsureIndexes(ctx context.Context, client *Client) error {
	if client == nil {
		return errors.New("MongoDB client is required")
	}

	collection := client.URLsCollection()
	if collection == nil {
		return errors.New("URLs collection is required")
	}

	if _, err := collection.Indexes().CreateMany(ctx, URLIndexModels()); err != nil {
		return fmt.Errorf("create URL indexes: %w", err)
	}

	return nil
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
	}
}
