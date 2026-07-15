package httpapi

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/tapadar13/url-shortener/apps/api/internal/auth"
)

type AuthService interface {
	Register(context.Context, string, string) (auth.User, error)
	Authenticate(context.Context, string, string) (auth.User, error)
	GetUser(context.Context, string) (auth.User, error)
}

type AccessTokenIssuer interface {
	Issue(string) (string, time.Time, error)
}

type RefreshSessionManager interface {
	Create(context.Context, string) (auth.Session, string, error)
	Rotate(context.Context, string) (auth.Session, string, error)
	Revoke(context.Context, string) error
}

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	AccessToken  string           `json:"accessToken"`
	RefreshToken string           `json:"refreshToken,omitempty"`
	TokenType    string           `json:"tokenType"`
	ExpiresAt    time.Time        `json:"expiresAt"`
	User         userAuthResponse `json:"user"`
}

type userAuthResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type refreshResponse struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	TokenType    string    `json:"tokenType"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

func newRegisterHandler(service AuthService, issuer AccessTokenIssuer, sessions RefreshSessionManager) http.HandlerFunc {
	return newAuthHandler(service, issuer, sessions, false)
}

func newLoginHandler(service AuthService, issuer AccessTokenIssuer, sessions RefreshSessionManager) http.HandlerFunc {
	return newAuthHandler(service, issuer, sessions, true)
}

func newAuthHandler(service AuthService, issuer AccessTokenIssuer, sessions RefreshSessionManager, login bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request authRequest
		if err := decodeJSONRequest(r, &request); err != nil {
			writeAuthDecodeError(w, err)
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
		refreshToken := ""
		if sessions != nil {
			if _, refreshToken, err = sessions.Create(r.Context(), user.ID); err != nil {
				writeError(w, http.StatusInternalServerError, "session_issue_failed", "could not issue refresh token")
				return
			}
		}
		status := http.StatusCreated
		if login {
			status = http.StatusOK
		}
		writeJSON(w, status, authResponse{AccessToken: accessToken, RefreshToken: refreshToken, TokenType: "Bearer", ExpiresAt: expiresAt, User: userAuthResponse{ID: user.ID, Email: user.Email}})
	}
}

func newRefreshHandler(issuer AccessTokenIssuer, sessions RefreshSessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request refreshRequest
		if err := decodeJSONRequest(r, &request); err != nil {
			writeAuthDecodeError(w, err)
			return
		}
		if request.RefreshToken == "" {
			writeError(w, http.StatusBadRequest, "invalid_request", "refresh token is required")
			return
		}
		session, refreshToken, err := sessions.Rotate(r.Context(), request.RefreshToken)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid_refresh_token", "refresh token is invalid or expired")
			return
		}
		accessToken, expiresAt, err := issuer.Issue(session.UserID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "token_issue_failed", "could not issue access token")
			return
		}
		writeJSON(w, http.StatusOK, refreshResponse{AccessToken: accessToken, RefreshToken: refreshToken, TokenType: "Bearer", ExpiresAt: expiresAt})
	}
}

func newLogoutHandler(sessions RefreshSessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request refreshRequest
		if err := decodeJSONRequest(r, &request); err != nil {
			writeAuthDecodeError(w, err)
			return
		}
		if request.RefreshToken == "" {
			writeError(w, http.StatusBadRequest, "invalid_request", "refresh token is required")
			return
		}
		if err := sessions.Revoke(r.Context(), request.RefreshToken); err != nil {
			writeError(w, http.StatusUnauthorized, "invalid_refresh_token", "refresh token is invalid or expired")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func newCurrentUserHandler(service AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := CurrentUserID(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized", "authentication is required")
			return
		}
		user, err := service.GetUser(r.Context(), userID)
		if err != nil {
			if errors.Is(err, auth.ErrUserNotFound) {
				writeError(w, http.StatusUnauthorized, "unauthorized", "authentication is required")
				return
			}
			writeError(w, http.StatusInternalServerError, "internal_error", "an unexpected error occurred")
			return
		}
		writeJSON(w, http.StatusOK, userAuthResponse{ID: user.ID, Email: user.Email})
	}
}

func writeAuthDecodeError(w http.ResponseWriter, err error) {
	if isRequestBodyTooLarge(err) {
		writeError(w, http.StatusRequestEntityTooLarge, "request_entity_too_large", "request body exceeds the configured size limit")
		return
	}
	writeError(w, http.StatusBadRequest, "invalid_request", "request body must be a valid JSON object")
}

func writeAuthError(w http.ResponseWriter, err error, login bool) {
	if errors.Is(err, auth.ErrInvalidCredentials) {
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "email or password is invalid")
		return
	}
	if errors.Is(err, auth.ErrEmailTaken) {
		writeError(w, http.StatusConflict, "email_taken", "email is already registered")
		return
	}
	if !login && isAuthValidationError(err) {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	writeError(w, http.StatusInternalServerError, "internal_error", "an unexpected error occurred")
}

func isAuthValidationError(err error) bool {
	return errors.Is(err, auth.ErrEmailRequired) ||
		errors.Is(err, auth.ErrEmailInvalid) ||
		errors.Is(err, auth.ErrPasswordRequired) ||
		errors.Is(err, auth.ErrPasswordTooShort) ||
		errors.Is(err, auth.ErrPasswordTooLong)
}
