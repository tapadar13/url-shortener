package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/tapadar13/url-shortener/apps/api/internal/shortcode"
	urlmodel "github.com/tapadar13/url-shortener/apps/api/internal/url"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type insertOneCollection interface {
	InsertOne(ctx context.Context, document any, opts ...options.Lister[options.InsertOneOptions]) (*mongo.InsertOneResult, error)
}

type findOneCollection interface {
	FindOne(ctx context.Context, filter any, opts ...options.Lister[options.FindOneOptions]) *mongo.SingleResult
}

type findOneAndUpdateCollection interface {
	FindOneAndUpdate(ctx context.Context, filter any, update any, opts ...options.Lister[options.FindOneAndUpdateOptions]) *mongo.SingleResult
}

type findCollection interface {
	Find(context.Context, any, ...options.Lister[options.FindOptions]) (*mongo.Cursor, error)
}

type deleteOneCollection interface {
	DeleteOne(ctx context.Context, filter any, opts ...options.Lister[options.DeleteOneOptions]) (*mongo.DeleteResult, error)
}

type collection interface {
	insertOneCollection
	findOneCollection
	findOneAndUpdateCollection
	deleteOneCollection
}

func (r *Repository) ListByOwner(ctx context.Context, ownerID string, limit int64) ([]urlmodel.URL, error) {
	if r == nil || r.collection == nil {
		return nil, errors.New("MongoDB URL collection is required")
	}
	if ownerID == "" {
		return nil, errors.New("URL owner is required")
	}
	if limit <= 0 || limit > 100 {
		return nil, errors.New("URL list limit must be between 1 and 100")
	}
	findCollection, ok := r.collection.(findCollection)
	if !ok {
		return nil, errors.New("MongoDB URL collection does not support listing")
	}
	cursor, err := findCollection.Find(ctx, bson.M{"owner_id": ownerID}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}, {Key: "_id", Value: -1}}).SetLimit(limit))
	if err != nil {
		return nil, fmt.Errorf("find URLs by owner: %w", err)
	}
	if cursor == nil {
		return nil, errors.New("find URLs by owner: missing cursor")
	}
	defer cursor.Close(ctx)
	var documents []urlDocument
	if err := cursor.All(ctx, &documents); err != nil {
		return nil, fmt.Errorf("decode URLs by owner: %w", err)
	}
	result := make([]urlmodel.URL, 0, len(documents))
	for _, document := range documents {
		record, err := document.toDomain()
		if err != nil {
			return nil, err
		}
		result = append(result, record)
	}
	return result, nil
}

type Repository struct {
	collection collection
	now        func() time.Time
}

func New(collection *mongo.Collection) *Repository {
	if collection == nil {
		return &Repository{}
	}

	return newRepository(collection)
}

func newRepository(collection collection) *Repository {
	return newRepositoryWithClock(collection, time.Now)
}

func newRepositoryWithClock(collection collection, now func() time.Time) *Repository {
	if now == nil {
		now = time.Now
	}

	return &Repository{
		collection: collection,
		now:        now,
	}
}

func (r *Repository) Create(ctx context.Context, record urlmodel.URL) (urlmodel.URL, error) {
	if r == nil || r.collection == nil {
		return urlmodel.URL{}, errors.New("MongoDB URL collection is required")
	}

	if err := record.Validate(); err != nil {
		return urlmodel.URL{}, err
	}

	doc, err := newURLDocument(record)
	if err != nil {
		return urlmodel.URL{}, err
	}

	result, err := r.collection.InsertOne(ctx, doc)
	if mongo.IsDuplicateKeyError(err) {
		return urlmodel.URL{}, fmt.Errorf("%w: %w", urlmodel.ErrDuplicateShortCode, err)
	}

	if err != nil {
		return urlmodel.URL{}, fmt.Errorf("insert URL: %w", err)
	}

	if result == nil {
		return urlmodel.URL{}, errors.New("insert URL: missing insert result")
	}

	insertedID, ok := result.InsertedID.(bson.ObjectID)
	if !ok {
		return urlmodel.URL{}, fmt.Errorf("insert URL: expected ObjectID, got %T", result.InsertedID)
	}

	doc.ID = insertedID

	created, err := doc.toDomain()
	if err != nil {
		return urlmodel.URL{}, err
	}

	return created, nil
}

func (r *Repository) FindByShortCode(ctx context.Context, shortCode string) (urlmodel.URL, error) {
	return r.findByShortCode(ctx, shortCode, "")
}

func (r *Repository) FindByShortCodeForOwner(ctx context.Context, ownerID, shortCode string) (urlmodel.URL, error) {
	return r.findByShortCode(ctx, shortCode, ownerID)
}

func (r *Repository) findByShortCode(ctx context.Context, shortCode, ownerID string) (urlmodel.URL, error) {
	if r == nil || r.collection == nil {
		return urlmodel.URL{}, errors.New("MongoDB URL collection is required")
	}

	normalizedShortCode, err := shortcode.Normalize(shortCode)
	if err != nil {
		return urlmodel.URL{}, err
	}

	filter := r.activeShortCodeFilter(normalizedShortCode)
	if ownerID != "" {
		filter = append(filter, bson.E{Key: "owner_id", Value: ownerID})
	}
	result := r.collection.FindOne(ctx, filter)
	if result == nil {
		return urlmodel.URL{}, errors.New("find URL by short code: missing result")
	}

	var doc urlDocument
	if err := result.Decode(&doc); errors.Is(err, mongo.ErrNoDocuments) {
		return urlmodel.URL{}, fmt.Errorf("%w: %w", urlmodel.ErrNotFound, err)
	} else if err != nil {
		return urlmodel.URL{}, fmt.Errorf("find URL by short code: %w", err)
	}

	record, err := doc.toDomain()
	if err != nil {
		return urlmodel.URL{}, err
	}

	return record, nil
}

func (r *Repository) UpdateLongURL(ctx context.Context, params urlmodel.UpdateLongURLParams) (urlmodel.URL, error) {
	return r.updateLongURL(ctx, params, "")
}

func (r *Repository) UpdateLongURLForOwner(ctx context.Context, params urlmodel.UpdateLongURLParams, ownerID string) (urlmodel.URL, error) {
	return r.updateLongURL(ctx, params, ownerID)
}

func (r *Repository) updateLongURL(ctx context.Context, params urlmodel.UpdateLongURLParams, ownerID string) (urlmodel.URL, error) {
	if r == nil || r.collection == nil {
		return urlmodel.URL{}, errors.New("MongoDB URL collection is required")
	}

	normalizedShortCode, err := shortcode.Normalize(params.ShortCode)
	if err != nil {
		return urlmodel.URL{}, err
	}

	normalizedLongURL, err := urlmodel.NormalizeLongURL(params.LongURL)
	if err != nil {
		return urlmodel.URL{}, err
	}

	if params.UpdatedAt.IsZero() {
		return urlmodel.URL{}, urlmodel.ErrTimestampRequired
	}

	updatedAt := params.UpdatedAt.UTC()
	filter := r.activeShortCodeFilter(normalizedShortCode)
	if ownerID != "" {
		filter = append(filter, bson.E{Key: "owner_id", Value: ownerID})
	}
	result := r.collection.FindOneAndUpdate(
		ctx,
		filter,
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "url", Value: normalizedLongURL},
			{Key: "updated_at", Value: updatedAt},
		}}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	if result == nil {
		return urlmodel.URL{}, errors.New("update URL by short code: missing result")
	}

	var doc urlDocument
	if err := result.Decode(&doc); errors.Is(err, mongo.ErrNoDocuments) {
		return urlmodel.URL{}, fmt.Errorf("%w: %w", urlmodel.ErrNotFound, err)
	} else if err != nil {
		return urlmodel.URL{}, fmt.Errorf("update URL by short code: %w", err)
	}

	updated, err := doc.toDomain()
	if err != nil {
		return urlmodel.URL{}, err
	}

	return updated, nil
}

func (r *Repository) DeleteByShortCode(ctx context.Context, shortCode string) error {
	return r.deleteByShortCode(ctx, shortCode, "")
}

func (r *Repository) DeleteByShortCodeForOwner(ctx context.Context, shortCode, ownerID string) error {
	return r.deleteByShortCode(ctx, shortCode, ownerID)
}

func (r *Repository) deleteByShortCode(ctx context.Context, shortCode, ownerID string) error {
	if r == nil || r.collection == nil {
		return errors.New("MongoDB URL collection is required")
	}

	normalizedShortCode, err := shortcode.Normalize(shortCode)
	if err != nil {
		return err
	}

	filter := r.activeShortCodeFilter(normalizedShortCode)
	if ownerID != "" {
		filter = append(filter, bson.E{Key: "owner_id", Value: ownerID})
	}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("delete URL by short code: %w", err)
	}

	if result == nil {
		return errors.New("delete URL by short code: missing result")
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("%w: %s", urlmodel.ErrNotFound, normalizedShortCode)
	}

	return nil
}

func (r *Repository) RecordAccess(ctx context.Context, params urlmodel.RecordAccessParams) (urlmodel.URL, error) {
	if r == nil || r.collection == nil {
		return urlmodel.URL{}, errors.New("MongoDB URL collection is required")
	}

	normalizedShortCode, err := shortcode.Normalize(params.ShortCode)
	if err != nil {
		return urlmodel.URL{}, err
	}

	if params.AccessedAt.IsZero() {
		return urlmodel.URL{}, urlmodel.ErrTimestampRequired
	}

	accessedAt := params.AccessedAt.UTC()
	result := r.collection.FindOneAndUpdate(
		ctx,
		r.activeShortCodeFilter(normalizedShortCode),
		bson.D{
			{Key: "$inc", Value: bson.D{{Key: "access_count", Value: 1}}},
			{Key: "$set", Value: bson.D{{Key: "last_accessed_at", Value: accessedAt}}},
		},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	if result == nil {
		return urlmodel.URL{}, errors.New("record URL access: missing result")
	}

	var doc urlDocument
	if err := result.Decode(&doc); errors.Is(err, mongo.ErrNoDocuments) {
		return urlmodel.URL{}, fmt.Errorf("%w: %w", urlmodel.ErrNotFound, err)
	} else if err != nil {
		return urlmodel.URL{}, fmt.Errorf("record URL access: %w", err)
	}

	recorded, err := doc.toDomain()
	if err != nil {
		return urlmodel.URL{}, err
	}

	return recorded, nil
}

func (r *Repository) activeShortCodeFilter(shortCode string) bson.D {
	now := time.Now
	if r != nil && r.now != nil {
		now = r.now
	}

	return bson.D{
		{Key: "short_code", Value: shortCode},
		{Key: "$or", Value: bson.A{
			bson.D{{Key: "expires_at", Value: nil}},
			bson.D{{Key: "expires_at", Value: bson.D{{Key: "$gt", Value: now().UTC()}}}},
		}},
	}
}
