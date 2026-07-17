package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/tapadar13/url-shortener/apps/api/internal/auth"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type sessionDocument struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	UserID    string        `bson:"user_id"`
	TokenHash string        `bson:"token_hash"`
	CreatedAt time.Time     `bson:"created_at"`
	ExpiresAt time.Time     `bson:"expires_at"`
	RevokedAt *time.Time    `bson:"revoked_at,omitempty"`
}

type updateOneCollection interface {
	UpdateOne(context.Context, any, any, ...options.Lister[options.UpdateOneOptions]) (*mongo.UpdateResult, error)
}

func (r *Repository) CreateSession(ctx context.Context, session auth.Session) (auth.Session, error) {
	if r == nil || r.collection == nil {
		return auth.Session{}, errors.New("MongoDB user collection is required")
	}
	if session.UserID == "" || session.TokenHash == "" || session.CreatedAt.IsZero() || session.ExpiresAt.IsZero() {
		return auth.Session{}, errors.New("session fields are required")
	}
	document := sessionDocument{UserID: session.UserID, TokenHash: session.TokenHash, CreatedAt: session.CreatedAt.UTC(), ExpiresAt: session.ExpiresAt.UTC(), RevokedAt: session.RevokedAt}
	result, err := r.collection.InsertOne(ctx, document)
	if err != nil {
		return auth.Session{}, fmt.Errorf("insert session: %w", err)
	}
	if result == nil {
		return auth.Session{}, errors.New("insert session: missing insert result")
	}
	id, ok := result.InsertedID.(bson.ObjectID)
	if !ok {
		return auth.Session{}, fmt.Errorf("insert session: expected ObjectID, got %T", result.InsertedID)
	}
	document.ID = id
	return document.toDomain()
}

func (r *Repository) FindSessionByTokenHash(ctx context.Context, tokenHash string) (auth.Session, error) {
	if r == nil || r.collection == nil {
		return auth.Session{}, errors.New("MongoDB user collection is required")
	}
	result := r.collection.FindOne(ctx, bson.M{"token_hash": tokenHash, "revoked_at": nil, "expires_at": bson.M{"$gt": time.Now().UTC()}})
	if result == nil {
		return auth.Session{}, errors.New("find session: missing result")
	}
	var document sessionDocument
	if err := result.Decode(&document); errors.Is(err, mongo.ErrNoDocuments) {
		return auth.Session{}, fmt.Errorf("%w: %w", auth.ErrSessionNotFound, err)
	} else if err != nil {
		return auth.Session{}, fmt.Errorf("find session: %w", err)
	}
	return document.toDomain()
}

func (r *Repository) RevokeSession(ctx context.Context, sessionID string, revokedAt time.Time) error {
	if r == nil || r.collection == nil {
		return errors.New("MongoDB user collection is required")
	}
	id, err := bson.ObjectIDFromHex(sessionID)
	if err != nil {
		return fmt.Errorf("parse session ID: %w", err)
	}
	collection, ok := r.collection.(updateOneCollection)
	if !ok {
		return errors.New("MongoDB user collection does not support session updates")
	}
	result, err := collection.UpdateOne(ctx, bson.M{"_id": id, "revoked_at": nil}, bson.M{"$set": bson.M{"revoked_at": revokedAt.UTC()}})
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	if result == nil || result.MatchedCount == 0 {
		return auth.ErrSessionNotFound
	}
	return nil
}

func (document sessionDocument) toDomain() (auth.Session, error) {
	if document.ID.IsZero() || document.UserID == "" || document.TokenHash == "" || document.CreatedAt.IsZero() || document.ExpiresAt.IsZero() {
		return auth.Session{}, errors.New("invalid session document")
	}
	return auth.Session{ID: document.ID.Hex(), UserID: document.UserID, TokenHash: document.TokenHash, CreatedAt: document.CreatedAt.UTC(), ExpiresAt: document.ExpiresAt.UTC(), RevokedAt: document.RevokedAt}, nil
}

var _ auth.SessionRepository = (*Repository)(nil)
