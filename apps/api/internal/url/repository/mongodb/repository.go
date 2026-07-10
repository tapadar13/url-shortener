package mongodb

import (
	"context"
	"errors"
	"fmt"

	urlmodel "github.com/tapadar13/url-shortener/apps/api/internal/url"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type insertOneCollection interface {
	InsertOne(ctx context.Context, document any, opts ...options.Lister[options.InsertOneOptions]) (*mongo.InsertOneResult, error)
}

type Repository struct {
	collection insertOneCollection
}

func New(collection *mongo.Collection) *Repository {
	if collection == nil {
		return &Repository{}
	}

	return newRepository(collection)
}

func newRepository(collection insertOneCollection) *Repository {
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
