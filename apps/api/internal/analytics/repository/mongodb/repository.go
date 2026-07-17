package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/tapadar13/url-shortener/apps/api/internal/analytics"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type updateOneCollection interface {
	UpdateOne(ctx context.Context, filter any, update any, opts ...options.Lister[options.UpdateOneOptions]) (*mongo.UpdateResult, error)
}

type findCollection interface {
	Find(ctx context.Context, filter any, opts ...options.Lister[options.FindOptions]) (*mongo.Cursor, error)
}

type collection interface {
	updateOneCollection
	findCollection
}

type Repository struct {
	collection collection
}

type dailyClicksDocument struct {
	DayStart   time.Time `bson:"day_start"`
	ClickCount int64     `bson:"click_count"`
}

func New(collection *mongo.Collection) *Repository {
	if collection == nil {
		return &Repository{}
	}

	return newRepository(collection)
}

func newRepository(collection collection) *Repository {
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

func (r *Repository) FindDailyClicks(ctx context.Context, rangeValue analytics.Range) ([]analytics.DailyClicks, error) {
	if r == nil || r.collection == nil {
		return nil, errors.New("MongoDB analytics collection is required")
	}

	normalized, err := analytics.NewRange(rangeValue.ShortCode, rangeValue.Start, rangeValue.EndExclusive)
	if err != nil {
		return nil, err
	}

	filter := bson.D{
		{Key: "short_code", Value: normalized.ShortCode},
		{Key: "day_start", Value: bson.D{
			{Key: "$gte", Value: normalized.Start},
			{Key: "$lt", Value: normalized.EndExclusive},
		}},
	}
	cursor, err := r.collection.Find(
		ctx,
		filter,
		options.Find().
			SetProjection(bson.D{
				{Key: "_id", Value: 0},
				{Key: "day_start", Value: 1},
				{Key: "click_count", Value: 1},
			}).
			SetSort(bson.D{{Key: "day_start", Value: 1}}),
	)
	if err != nil {
		return nil, fmt.Errorf("find daily click analytics: %w", err)
	}
	if cursor == nil {
		return nil, errors.New("find daily click analytics: missing cursor")
	}

	var documents []dailyClicksDocument
	if err := cursor.All(ctx, &documents); err != nil {
		return nil, fmt.Errorf("decode daily click analytics: %w", err)
	}

	dailyClicks := make([]analytics.DailyClicks, 0, len(documents))
	for _, document := range documents {
		daily, err := analytics.NewDailyClicks(document.DayStart, document.ClickCount)
		if err != nil {
			return nil, fmt.Errorf("validate daily click analytics: %w", err)
		}

		dailyClicks = append(dailyClicks, daily)
	}

	return dailyClicks, nil
}

var _ analytics.Recorder = (*Repository)(nil)
var _ analytics.Reader = (*Repository)(nil)
