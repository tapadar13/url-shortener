package mongodb

import (
	"context"
	"errors"
	"fmt"

	"github.com/tapadar13/url-shortener/apps/api/internal/auth"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type insertOneCollection interface {
	InsertOne(context.Context, any, ...options.Lister[options.InsertOneOptions]) (*mongo.InsertOneResult, error)
}

type findOneCollection interface {
	FindOne(context.Context, any, ...options.Lister[options.FindOneOptions]) *mongo.SingleResult
}

type collection interface {
	insertOneCollection
	findOneCollection
}

type Repository struct{ collection collection }

func New(collection *mongo.Collection) *Repository {
	return &Repository{collection: collection}
}

func newRepository(collection collection) *Repository {
	return &Repository{collection: collection}
}

func (r *Repository) CreateUser(ctx context.Context, user auth.User) (auth.User, error) {
	if r == nil || r.collection == nil {
		return auth.User{}, errors.New("MongoDB user collection is required")
	}
	document, err := newUserDocument(user)
	if err != nil {
		return auth.User{}, err
	}
	result, err := r.collection.InsertOne(ctx, document)
	if mongo.IsDuplicateKeyError(err) {
		return auth.User{}, fmt.Errorf("%w: %w", auth.ErrEmailTaken, err)
	}
	if err != nil {
		return auth.User{}, fmt.Errorf("insert user: %w", err)
	}
	if result == nil {
		return auth.User{}, errors.New("insert user: missing insert result")
	}
	id, ok := result.InsertedID.(bson.ObjectID)
	if !ok {
		return auth.User{}, fmt.Errorf("insert user: expected ObjectID, got %T", result.InsertedID)
	}
	document.ID = id
	return document.toDomain()
}

func (r *Repository) FindUserByEmail(ctx context.Context, email string) (auth.User, error) {
	if r == nil || r.collection == nil {
		return auth.User{}, errors.New("MongoDB user collection is required")
	}
	normalized, err := auth.NormalizeEmail(email)
	if err != nil {
		return auth.User{}, err
	}
	return r.find(ctx, bson.M{"email": normalized}, "find user by email")
}

func (r *Repository) FindUserByID(ctx context.Context, id string) (auth.User, error) {
	if r == nil || r.collection == nil {
		return auth.User{}, errors.New("MongoDB user collection is required")
	}
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return auth.User{}, fmt.Errorf("parse user ID: %w", err)
	}
	return r.find(ctx, bson.M{"_id": objectID}, "find user by ID")
}

func (r *Repository) find(ctx context.Context, filter bson.M, operation string) (auth.User, error) {
	result := r.collection.FindOne(ctx, filter)
	if result == nil {
		return auth.User{}, fmt.Errorf("%s: missing result", operation)
	}
	var document userDocument
	if err := result.Decode(&document); errors.Is(err, mongo.ErrNoDocuments) {
		return auth.User{}, fmt.Errorf("%w: %w", auth.ErrUserNotFound, err)
	} else if err != nil {
		return auth.User{}, fmt.Errorf("%s: %w", operation, err)
	}
	return document.toDomain()
}

var _ auth.Repository = (*Repository)(nil)
