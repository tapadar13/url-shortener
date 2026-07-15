package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/tapadar13/url-shortener/apps/api/internal/auth"
)

type fakeAuthService struct {
	user UserAuthTestUser
	err  error
}

type UserAuthTestUser struct {
	ID    string
	Email string
}

func (s *fakeAuthService) Register(context.Context, string, string) (auth.User, error) {
	if s.err != nil {
		return auth.User{}, s.err
	}
	return auth.User{ID: s.user.ID, Email: s.user.Email}, nil
}

func (s *fakeAuthService) Authenticate(context.Context, string, string) (auth.User, error) {
	if s.err != nil {
		return auth.User{}, s.err
	}
	return auth.User{ID: s.user.ID, Email: s.user.Email}, nil
}

type fakeAccessTokenIssuer struct{ err error }

func (i *fakeAccessTokenIssuer) Issue(string) (string, time.Time, error) {
	if i.err != nil {
		return "", time.Time{}, i.err
	}
	return "signed-token", time.Date(2026, time.July, 15, 11, 0, 0, 0, time.UTC), nil
}

func TestRouterRegistersUserAndReturnsAccessToken(t *testing.T) {
	router := NewRouter(Dependencies{
		AuthService:       &fakeAuthService{user: UserAuthTestUser{ID: "user-1", Email: "user@example.com"}},
		AccessTokenIssuer: &fakeAccessTokenIssuer{},
	})
	response := executeRequestWithBody(t, router, http.MethodPost, "/auth/register", `{"email":"user@example.com","password":"correct horse battery staple"}`)
	assertStatus(t, response, http.StatusCreated)

	var body authResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode auth response: %v", err)
	}
	if body.AccessToken != "signed-token" || body.TokenType != "Bearer" || body.User.ID != "user-1" || body.User.Email != "user@example.com" {
		t.Fatalf("unexpected auth response: %+v", body)
	}
}

func TestRouterHidesLoginCredentialFailures(t *testing.T) {
	router := NewRouter(Dependencies{
		AuthService:       &fakeAuthService{err: auth.ErrInvalidCredentials},
		AccessTokenIssuer: &fakeAccessTokenIssuer{},
	})
	response := executeRequestWithBody(t, router, http.MethodPost, "/auth/login", `{"email":"user@example.com","password":"wrong password"}`)
	assertStatus(t, response, http.StatusUnauthorized)
	assertAuthErrorCode(t, response, "invalid_credentials")
}

func TestRouterMapsDuplicateRegistrationToConflict(t *testing.T) {
	router := NewRouter(Dependencies{
		AuthService:       &fakeAuthService{err: auth.ErrEmailTaken},
		AccessTokenIssuer: &fakeAccessTokenIssuer{},
	})
	response := executeRequestWithBody(t, router, http.MethodPost, "/auth/register", `{"email":"user@example.com","password":"correct horse battery staple"}`)
	assertStatus(t, response, http.StatusConflict)
	assertAuthErrorCode(t, response, "email_taken")
}

func TestRouterReturnsBadRequestForMalformedAuthJSON(t *testing.T) {
	router := NewRouter(Dependencies{
		AuthService:       &fakeAuthService{},
		AccessTokenIssuer: &fakeAccessTokenIssuer{},
	})
	response := executeRequestWithBody(t, router, http.MethodPost, "/auth/login", `{invalid`)
	assertStatus(t, response, http.StatusBadRequest)
}

func assertAuthErrorCode(t *testing.T, response *http.Response, expected string) {
	t.Helper()
	var body errorResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if body.Error.Code != expected {
		t.Fatalf("expected error code %q, got %q", expected, body.Error.Code)
	}
}
