package auth

import (
	"context"
	"errors"
	"time"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrEmailTaken      = errors.New("email is already registered")
	ErrSessionNotFound = errors.New("session not found")
)

type Repository interface {
	CreateUser(ctx context.Context, user User) (User, error)
	FindUserByEmail(ctx context.Context, email string) (User, error)
	FindUserByID(ctx context.Context, id string) (User, error)
}

type SessionRepository interface {
	CreateSession(ctx context.Context, session Session) (Session, error)
	FindSessionByTokenHash(ctx context.Context, tokenHash string) (Session, error)
	RevokeSession(ctx context.Context, sessionID string, revokedAt time.Time) error
}
