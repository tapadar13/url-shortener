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

type collection interface {
	insertOneCollection
	findOneCollection
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
