package mongodb

import (
	"context"
	"errors"
	"fmt"

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

type deleteOneCollection interface {
	DeleteOne(ctx context.Context, filter any, opts ...options.Lister[options.DeleteOneOptions]) (*mongo.DeleteResult, error)
}

type collection interface {
	insertOneCollection
	findOneCollection
	findOneAndUpdateCollection
	deleteOneCollection
}

type Repository struct {
	collection collection
}

func New(collection *mongo.Collection) *Repository {
	if collection == nil {
		return &Repository{}
	}

	return newRepository(collection)
}

func newRepository(collection collection) *Repository {
	return &Repository{
		collection: collection,
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
	if r == nil || r.collection == nil {
		return urlmodel.URL{}, errors.New("MongoDB URL collection is required")
	}

	normalizedShortCode, err := shortcode.Normalize(shortCode)
	if err != nil {
		return urlmodel.URL{}, err
	}

	result := r.collection.FindOne(ctx, bson.D{{Key: "short_code", Value: normalizedShortCode}})
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
	result := r.collection.FindOneAndUpdate(
		ctx,
		bson.D{{Key: "short_code", Value: normalizedShortCode}},
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
	if r == nil || r.collection == nil {
		return errors.New("MongoDB URL collection is required")
	}

	normalizedShortCode, err := shortcode.Normalize(shortCode)
	if err != nil {
		return err
	}

	result, err := r.collection.DeleteOne(ctx, bson.D{{Key: "short_code", Value: normalizedShortCode}})
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
