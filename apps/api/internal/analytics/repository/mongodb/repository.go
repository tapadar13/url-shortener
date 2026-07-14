package mongodb

import (
	"context"
	"errors"
	"fmt"

	"github.com/tapadar13/url-shortener/apps/api/internal/analytics"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type updateOneCollection interface {
	UpdateOne(ctx context.Context, filter any, update any, opts ...options.Lister[options.UpdateOneOptions]) (*mongo.UpdateResult, error)
}

type Repository struct {
	collection updateOneCollection
}

func New(collection *mongo.Collection) *Repository {
	if collection == nil {
		return &Repository{}
	}

	return newRepository(collection)
}

func newRepository(collection updateOneCollection) *Repository {
	return &Repository{collection: collection}
}

func (r *Repository) RecordClick(ctx context.Context, click analytics.Click) error {
	if r == nil || r.collection == nil {
		return errors.New("MongoDB analytics collection is required")
	}

	normalized, err := analytics.NewClick(click.ShortCode, click.ClickedAt)
	if err != nil {
		return err
	}

	filter := bson.D{
		{Key: "short_code", Value: normalized.ShortCode},
		{Key: "day_start", Value: normalized.DayStart()},
	}
	update := bson.D{
		{Key: "$inc", Value: bson.D{{Key: "click_count", Value: 1}}},
		{Key: "$min", Value: bson.D{{Key: "first_clicked_at", Value: normalized.ClickedAt}}},
		{Key: "$max", Value: bson.D{{Key: "last_clicked_at", Value: normalized.ClickedAt}}},
	}

	err = r.update(ctx, filter, update, true)
	if mongo.IsDuplicateKeyError(err) {
		err = r.update(ctx, filter, update, false)
	}
	if err != nil {
		return fmt.Errorf("record daily click analytics: %w", err)
	}

	return nil
}

func (r *Repository) update(ctx context.Context, filter bson.D, update bson.D, upsert bool) error {
	result, err := r.collection.UpdateOne(
		ctx,
		filter,
		update,
		options.UpdateOne().SetUpsert(upsert),
	)
	if err != nil {
		return err
	}

	if result == nil {
		return errors.New("missing update result")
	}

	return nil
}

var _ analytics.Recorder = (*Repository)(nil)
