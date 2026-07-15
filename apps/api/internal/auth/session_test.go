package auth

import (
	"strings"
	"testing"
	"time"
)

func TestNewSessionCreatesOpaqueHashedToken(t *testing.T) {
	now := time.Date(2026, time.July, 15, 10, 0, 0, 0, time.UTC)
	session, token, err := NewSession("user-1", now, 30*24*time.Hour)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if token == "" || strings.Contains(session.TokenHash, token) || session.TokenHash != HashSessionToken(token) {
		t.Fatalf("expected hashed session token, session=%+v", session)
	}
	if !session.ExpiresAt.Equal(now.Add(30 * 24 * time.Hour)) {
		t.Fatalf("unexpected session expiry: %s", session.ExpiresAt)
	}
}

func TestNewSessionRejectsInvalidInput(t *testing.T) {
	if _, _, err := NewSession("", time.Now(), time.Hour); err != ErrSessionUserRequired {
		t.Fatalf("expected user validation error, got %v", err)
	}
	if _, _, err := NewSession("user-1", time.Now(), 0); err != ErrSessionExpiryInvalid {
		t.Fatalf("expected expiry validation error, got %v", err)
	}
}
