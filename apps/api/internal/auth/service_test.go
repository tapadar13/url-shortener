package auth

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeRepository struct {
	created User
	found   User
	err     error
}

func (r *fakeRepository) CreateUser(_ context.Context, user User) (User, error) {
	r.created = user
	if r.err != nil {
		return User{}, r.err
	}
	user.ID = "507f1f77bcf86cd799439011"
	return user, nil
}

func (r *fakeRepository) FindUserByEmail(context.Context, string) (User, error) {
	if r.err != nil {
		return User{}, r.err
	}
	return r.found, nil
}

func (r *fakeRepository) FindUserByID(context.Context, string) (User, error) {
	if r.err != nil {
		return User{}, r.err
	}
	return r.found, nil
}

func TestRegisterNormalizesEmailAndSetsTimestamps(t *testing.T) {
	repository := &fakeRepository{}
	now := time.Date(2026, time.July, 15, 10, 0, 0, 0, time.FixedZone("IST", 19800))
	service, err := newService(repository, func() time.Time { return now })
	if err != nil {
		t.Fatalf("create service: %v", err)
	}

	user, err := service.Register(context.Background(), " User@Example.COM ", "correct horse battery staple")
	if err != nil {
		t.Fatalf("register user: %v", err)
	}
	if user.Email != "user@example.com" || repository.created.CreatedAt != now.UTC() || repository.created.UpdatedAt != now.UTC() {
		t.Fatalf("unexpected registered user: %+v", user)
	}
	if repository.created.PasswordHash == "correct horse battery staple" {
		t.Fatal("expected password to be hashed")
	}
}

func TestAuthenticateHidesRepositoryAndPasswordErrors(t *testing.T) {
	passwordHash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	for _, repository := range []*fakeRepository{
		{err: ErrUserNotFound},
		{found: User{Email: "user@example.com", PasswordHash: passwordHash}},
	} {
		service, err := NewService(repository)
		if err != nil {
			t.Fatalf("create service: %v", err)
		}
		if _, err := service.Authenticate(context.Background(), "user@example.com", "wrong password"); !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("expected generic invalid credentials, got %v", err)
		}
	}
}

func TestAuthenticatePreservesUnexpectedRepositoryFailure(t *testing.T) {
	repositoryFailure := errors.New("database unavailable")
	service, err := NewService(&fakeRepository{err: repositoryFailure})
	if err != nil {
		t.Fatalf("create service: %v", err)
	}
	_, err = service.Authenticate(context.Background(), "user@example.com", "correct horse battery staple")
	if !errors.Is(err, repositoryFailure) || errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected infrastructure failure, got %v", err)
	}
}

func TestGetUserReturnsAuthenticatedIdentity(t *testing.T) {
	repository := &fakeRepository{found: User{ID: "user-1", Email: "user@example.com", PasswordHash: "hidden"}}
	service, err := NewService(repository)
	if err != nil {
		t.Fatalf("create service: %v", err)
	}
	user, err := service.GetUser(context.Background(), "user-1")
	if err != nil || user.ID != "user-1" {
		t.Fatalf("get user: user=%+v err=%v", user, err)
	}
}
