package auth

import (
	"context"
	"testing"
	"time"
)

type fakeSessionRepository struct {
	session Session
	created Session
	revoked string
}

func (r *fakeSessionRepository) CreateSession(_ context.Context, session Session) (Session, error) {
	session.ID = "session-1"
	r.created = session
	return session, nil
}

func (r *fakeSessionRepository) FindSessionByTokenHash(context.Context, string) (Session, error) {
	return r.session, nil
}

func (r *fakeSessionRepository) RevokeSession(_ context.Context, sessionID string, _ time.Time) error {
	r.revoked = sessionID
	return nil
}

func TestSessionServiceRotatesRefreshToken(t *testing.T) {
	now := time.Date(2026, time.July, 15, 10, 0, 0, 0, time.UTC)
	repository := &fakeSessionRepository{session: Session{ID: "old-session", UserID: "user-1", TokenHash: HashSessionToken("old-token"), CreatedAt: now, ExpiresAt: now.Add(time.Hour)}}
	service, err := newSessionService(repository, 30*24*time.Hour, func() time.Time { return now })
	if err != nil {
		t.Fatalf("create session service: %v", err)
	}
	rotated, token, err := service.Rotate(context.Background(), "old-token")
	if err != nil {
		t.Fatalf("rotate session: %v", err)
	}
	if repository.revoked != "old-session" || rotated.ID != "session-1" || token == "" || repository.created.UserID != "user-1" {
		t.Fatalf("unexpected rotation: revoked=%q rotated=%+v", repository.revoked, rotated)
	}
}
