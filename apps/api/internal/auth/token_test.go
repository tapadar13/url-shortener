package auth

import (
	"errors"
	"testing"
	"time"
)

func TestTokenIssueAndVerify(t *testing.T) {
	now := time.Date(2026, time.July, 15, 10, 0, 0, 0, time.UTC)
	service, err := newTokenService(TokenOptions{Secret: "a-long-development-secret-value", Issuer: "url-shortener", Audience: "api", TTL: time.Hour}, func() time.Time { return now })
	if err != nil {
		t.Fatalf("create token service: %v", err)
	}
	token, expiresAt, err := service.Issue("507f1f77bcf86cd799439011")
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	if !expiresAt.Equal(now.Add(time.Hour)) {
		t.Fatalf("unexpected expiry: %s", expiresAt)
	}
	claims, err := service.Verify(token)
	if err != nil || claims.UserID != "507f1f77bcf86cd799439011" {
		t.Fatalf("verify token: claims=%+v err=%v", claims, err)
	}
}

func TestTokenRejectsTamperingAndInvalidOptions(t *testing.T) {
	if _, err := NewTokenService(TokenOptions{TTL: time.Hour}); !errors.Is(err, ErrTokenSecretRequired) {
		t.Fatalf("expected secret validation error, got %v", err)
	}
	service, err := NewTokenService(TokenOptions{Secret: "secret", Issuer: "issuer", Audience: "audience", TTL: time.Hour})
	if err != nil {
		t.Fatalf("create token service: %v", err)
	}
	token, _, err := service.Issue("user-id")
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	if _, err := service.Verify(token + "tampered"); !errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("expected invalid token error, got %v", err)
	}
}
