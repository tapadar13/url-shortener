package mongodb

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/tapadar13/url-shortener/apps/api/internal/ratelimit"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type findOneAndUpdateCollection interface {
	FindOneAndUpdate(ctx context.Context, filter any, update any, opts ...options.Lister[options.FindOneAndUpdateOptions]) *mongo.SingleResult
}

type Repository struct {
	collection findOneAndUpdateCollection
}

type counterID struct {
	ClientKey   string        `bson:"client_key"`
	WindowStart bson.DateTime `bson:"window_start"`
}

type counterDocument struct {
	Count int `bson:"count"`
}

func New(collection *mongo.Collection) *Repository {
	if collection == nil {
		return &Repository{}
	}

	return newRepository(collection)
}

func newRepository(collection findOneAndUpdateCollection) *Repository {
	return &Repository{collection: collection}
}

func (r *Repository) Increment(ctx context.Context, params ratelimit.IncrementParams) (int, error) {
	if r == nil || r.collection == nil {
		return 0, errors.New("MongoDB rate limit collection is required")
	}

	clientKey := strings.TrimSpace(params.ClientKey)
	if clientKey == "" {
		return 0, ratelimit.ErrClientKeyRequired
	}

	if params.WindowStart.IsZero() {
		return 0, ratelimit.ErrWindowStartRequired
	}

	if params.ExpiresAt.IsZero() {
		return 0, ratelimit.ErrExpirationRequired
	}

	windowStart := params.WindowStart.UTC()
	expiresAt := params.ExpiresAt.UTC()
	if !expiresAt.After(windowStart) {
		return 0, ratelimit.ErrExpirationInvalid
	}

	result := r.collection.FindOneAndUpdate(
		ctx,
		bson.D{{Key: "_id", Value: counterID{
			ClientKey:   clientKey,
			WindowStart: bson.DateTime(windowStart.UnixMilli()),
		}}},
		bson.D{
			{Key: "$inc", Value: bson.D{{Key: "count", Value: 1}}},
			{Key: "$setOnInsert", Value: bson.D{{Key: "expires_at", Value: expiresAt}}},
		},
		options.FindOneAndUpdate().
			SetUpsert(true).
			SetReturnDocument(options.After),
	)
	if result == nil {
		return 0, errors.New("increment rate limit counter: missing result")
	}

	var document counterDocument
	if err := result.Decode(&document); err != nil {
		return 0, fmt.Errorf("increment rate limit counter: %w", err)
	}

	return document.Count, nil
}
