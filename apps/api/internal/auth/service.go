package auth

import (
	"context"
	"errors"
	"strings"
	"time"
)

var ErrInvalidCredentials = errors.New("invalid email or password")

type Service struct {
	repository Repository
	now        func() time.Time
}

func NewService(repository Repository) (*Service, error) {
	if repository == nil {
		return nil, errors.New("authentication repository is required")
	}
	return &Service{repository: repository, now: time.Now}, nil
}

func newService(repository Repository, now func() time.Time) (*Service, error) {
	service, err := NewService(repository)
	if err != nil {
		return nil, err
	}
	if now != nil {
		service.now = now
	}
	return service, nil
}

func (s *Service) Register(ctx context.Context, email, password string) (User, error) {
	if s == nil || s.repository == nil {
		return User{}, errors.New("authentication service is required")
	}
	normalizedEmail, err := NormalizeEmail(email)
	if err != nil {
		return User{}, err
	}
	passwordHash, err := HashPassword(password)
	if err != nil {
		return User{}, err
	}
	now := s.now().UTC()
	created, err := s.repository.CreateUser(ctx, User{
		Email:        normalizedEmail,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		return User{}, err
	}
	return created, nil
}

func (s *Service) Authenticate(ctx context.Context, email, password string) (User, error) {
	if s == nil || s.repository == nil {
		return User{}, errors.New("authentication service is required")
	}
	normalizedEmail, err := NormalizeEmail(email)
	if err != nil || strings.TrimSpace(password) == "" {
		return User{}, ErrInvalidCredentials
	}
	user, err := s.repository.FindUserByEmail(ctx, normalizedEmail)
	if err != nil {
		return User{}, ErrInvalidCredentials
	}
	if err := ComparePassword(password, user.PasswordHash); err != nil {
		return User{}, ErrInvalidCredentials
	}
	return user, nil
}
