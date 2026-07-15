package httpapi

import (
	"context"
	"net/http"
	"strings"

	"github.com/tapadar13/url-shortener/apps/api/internal/auth"
)

type AccessTokenVerifier interface {
	Verify(string) (auth.TokenClaims, error)
}

type authContextKey struct{}

func RequireAuth(verifier AccessTokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if verifier == nil {
				writeError(w, http.StatusUnauthorized, "unauthorized", "authentication is required")
				return
			}
			parts := strings.Fields(r.Header.Get("Authorization"))
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				writeError(w, http.StatusUnauthorized, "unauthorized", "authentication is required")
				return
			}
			claims, err := verifier.Verify(parts[1])
			if err != nil || strings.TrimSpace(claims.UserID) == "" {
				writeError(w, http.StatusUnauthorized, "unauthorized", "authentication is required")
				return
			}
			ctx := context.WithValue(r.Context(), authContextKey{}, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func CurrentUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(authContextKey{}).(string)
	return userID, ok && strings.TrimSpace(userID) != ""
}
