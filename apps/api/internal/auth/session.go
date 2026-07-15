package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
	"time"
)

var (
	ErrSessionUserRequired  = errors.New("session user ID is required")
	ErrSessionTokenRequired = errors.New("session token is required")
	ErrSessionExpiryInvalid = errors.New("session expiry must be in the future")
)

type Session struct {
	ID        string
	UserID    string
	TokenHash string
	CreatedAt time.Time
	ExpiresAt time.Time
	RevokedAt *time.Time
}

const refreshTokenBytes = 32

func NewSession(userID string, now time.Time, ttl time.Duration) (Session, string, error) {
	if strings.TrimSpace(userID) == "" {
		return Session{}, "", ErrSessionUserRequired
	}
	if ttl <= 0 {
		return Session{}, "", ErrSessionExpiryInvalid
	}
	tokenBytes := make([]byte, refreshTokenBytes)
	if _, err := rand.Read(tokenBytes); err != nil {
		return Session{}, "", err
	}
	plainToken := base64.RawURLEncoding.EncodeToString(tokenBytes)
	createdAt := now.UTC()
	return Session{
		UserID:    userID,
		TokenHash: HashSessionToken(plainToken),
		CreatedAt: createdAt,
		ExpiresAt: createdAt.Add(ttl),
	}, plainToken, nil
}

func HashSessionToken(token string) string {
	digest := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(digest[:])
}
