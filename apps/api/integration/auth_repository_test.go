//go:build integration

package integration_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/tapadar13/url-shortener/apps/api/internal/auth"
	authrepository "github.com/tapadar13/url-shortener/apps/api/internal/auth/repository/mongodb"
)

func TestAuthenticationRepositoryLifecycle(t *testing.T) {
	client := newIntegrationMongoClient(t)
	users := authrepository.New(client.UsersCollection())
	sessions := authrepository.NewSessionRepository(client.SessionsCollection())
	ctx, cancel := context.WithTimeout(context.Background(), integrationTimeout)
	defer cancel()

	passwordHash, err := auth.HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	now := time.Now().UTC()
	created, err := users.CreateUser(ctx, auth.User{Email: "User@Example.COM", PasswordHash: passwordHash, CreatedAt: now, UpdatedAt: now})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	found, err := users.FindUserByEmail(ctx, " user@example.com ")
	if err != nil || found.ID != created.ID || found.Email != "user@example.com" {
		t.Fatalf("find normalized user: user=%+v err=%v", found, err)
	}
	if _, err := users.CreateUser(ctx, auth.User{Email: "user@example.com", PasswordHash: passwordHash, CreatedAt: now, UpdatedAt: now}); !errors.Is(err, auth.ErrEmailTaken) {
		t.Fatalf("expected duplicate email error, got %v", err)
	}

	session, plainToken, err := auth.NewSession(created.ID, now, time.Hour)
	if err != nil {
		t.Fatalf("create session model: %v", err)
	}
	createdSession, err := sessions.CreateSession(ctx, session)
	if err != nil {
		t.Fatalf("persist session: %v", err)
	}
	foundSession, err := sessions.FindSessionByTokenHash(ctx, auth.HashSessionToken(plainToken))
	if err != nil || foundSession.ID != createdSession.ID || foundSession.UserID != created.ID {
		t.Fatalf("find session: session=%+v err=%v", foundSession, err)
	}
	if err := sessions.RevokeSession(ctx, createdSession.ID, now.Add(time.Minute)); err != nil {
		t.Fatalf("revoke session: %v", err)
	}
	if _, err := sessions.FindSessionByTokenHash(ctx, auth.HashSessionToken(plainToken)); !errors.Is(err, auth.ErrSessionNotFound) {
		t.Fatalf("expected revoked session to be unavailable, got %v", err)
	}
}
