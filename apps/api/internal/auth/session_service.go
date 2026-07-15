package auth

import (
	"context"
	"errors"
	"time"
)

type SessionService struct {
	repository SessionRepository
	now        func() time.Time
	ttl        time.Duration
}

func NewSessionService(repository SessionRepository, ttl time.Duration) (*SessionService, error) {
	if repository == nil {
		return nil, errors.New("session repository is required")
	}
	if ttl <= 0 {
		return nil, ErrSessionExpiryInvalid
	}
	return &SessionService{repository: repository, now: time.Now, ttl: ttl}, nil
}

func newSessionService(repository SessionRepository, ttl time.Duration, now func() time.Time) (*SessionService, error) {
	service, err := NewSessionService(repository, ttl)
	if err != nil {
		return nil, err
	}
	if now != nil {
		service.now = now
	}
	return service, nil
}

func (s *SessionService) Create(ctx context.Context, userID string) (Session, string, error) {
	if s == nil || s.repository == nil {
		return Session{}, "", errors.New("session service is required")
	}
	session, token, err := NewSession(userID, s.now(), s.ttl)
	if err != nil {
		return Session{}, "", err
	}
	created, err := s.repository.CreateSession(ctx, session)
	if err != nil {
		return Session{}, "", err
	}
	return created, token, nil
}

func (s *SessionService) Rotate(ctx context.Context, token string) (Session, string, error) {
	if s == nil || s.repository == nil {
		return Session{}, "", errors.New("session service is required")
	}
	if token == "" {
		return Session{}, "", ErrSessionTokenRequired
	}
	session, err := s.repository.FindSessionByTokenHash(ctx, HashSessionToken(token))
	if err != nil {
		return Session{}, "", err
	}
	revokedAt := s.now().UTC()
	if err := s.repository.RevokeSession(ctx, session.ID, revokedAt); err != nil {
		return Session{}, "", err
	}
	return s.Create(ctx, session.UserID)
}
