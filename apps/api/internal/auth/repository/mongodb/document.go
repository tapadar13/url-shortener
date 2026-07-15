package mongodb

import (
	"fmt"
	"time"

	"github.com/tapadar13/url-shortener/apps/api/internal/auth"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type userDocument struct {
	ID           bson.ObjectID `bson:"_id,omitempty"`
	Email        string        `bson:"email"`
	PasswordHash string        `bson:"password_hash"`
	CreatedAt    time.Time     `bson:"created_at"`
	UpdatedAt    time.Time     `bson:"updated_at"`
}

func newUserDocument(user auth.User) (userDocument, error) {
	email, err := auth.NormalizeEmail(user.Email)
	if err != nil {
		return userDocument{}, err
	}
	if user.PasswordHash == "" {
		return userDocument{}, auth.ErrPasswordHashInvalid
	}
	if user.CreatedAt.IsZero() || user.UpdatedAt.IsZero() {
		return userDocument{}, fmt.Errorf("user timestamps are required")
	}

	document := userDocument{
		Email:        email,
		PasswordHash: user.PasswordHash,
		CreatedAt:    user.CreatedAt.UTC(),
		UpdatedAt:    user.UpdatedAt.UTC(),
	}
	if user.ID != "" {
		id, err := bson.ObjectIDFromHex(user.ID)
		if err != nil {
			return userDocument{}, fmt.Errorf("parse user ID: %w", err)
		}
		document.ID = id
	}
	return document, nil
}

func (document userDocument) toDomain() (auth.User, error) {
	if document.ID.IsZero() {
		return auth.User{}, fmt.Errorf("user document ID is required")
	}
	email, err := auth.NormalizeEmail(document.Email)
	if err != nil {
		return auth.User{}, err
	}
	if document.PasswordHash == "" {
		return auth.User{}, auth.ErrPasswordHashInvalid
	}
	if document.CreatedAt.IsZero() || document.UpdatedAt.IsZero() {
		return auth.User{}, fmt.Errorf("user document timestamps are required")
	}
	return auth.User{
		ID:           document.ID.Hex(),
		Email:        email,
		PasswordHash: document.PasswordHash,
		CreatedAt:    document.CreatedAt.UTC(),
		UpdatedAt:    document.UpdatedAt.UTC(),
	}, nil
}
