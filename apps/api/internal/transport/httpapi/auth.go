package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/tapadar13/url-shortener/apps/api/internal/auth"
)

type AuthService interface {
	Register(context.Context, string, string) (auth.User, error)
	Authenticate(context.Context, string, string) (auth.User, error)
}

type AccessTokenIssuer interface {
	Issue(string) (string, time.Time, error)
}

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	AccessToken string           `json:"accessToken"`
	TokenType   string           `json:"tokenType"`
	ExpiresAt   time.Time        `json:"expiresAt"`
	User        userAuthResponse `json:"user"`
}

type userAuthResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func newRegisterHandler(service AuthService, issuer AccessTokenIssuer) http.HandlerFunc {
	return newAuthHandler(service, issuer, false)
}

func newLoginHandler(service AuthService, issuer AccessTokenIssuer) http.HandlerFunc {
	return newAuthHandler(service, issuer, true)
}

func newAuthHandler(service AuthService, issuer AccessTokenIssuer, login bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request authRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_request", "request body must be a valid JSON object")
			return
		}
		var user auth.User
		var err error
		if login {
			user, err = service.Authenticate(r.Context(), request.Email, request.Password)
		} else {
			user, err = service.Register(r.Context(), request.Email, request.Password)
		}
		if err != nil {
			writeAuthError(w, err, login)
			return
		}
		accessToken, expiresAt, err := issuer.Issue(user.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "token_issue_failed", "could not issue access token")
			return
		}
		status := http.StatusCreated
		if login {
			status = http.StatusOK
		}
		writeJSON(w, status, authResponse{AccessToken: accessToken, TokenType: "Bearer", ExpiresAt: expiresAt, User: userAuthResponse{ID: user.ID, Email: user.Email}})
	}
}

func writeAuthError(w http.ResponseWriter, err error, login bool) {
	if login || errors.Is(err, auth.ErrInvalidCredentials) {
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "email or password is invalid")
		return
	}
	if errors.Is(err, auth.ErrEmailTaken) {
		writeError(w, http.StatusConflict, "email_taken", "email is already registered")
		return
	}
	writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
}
