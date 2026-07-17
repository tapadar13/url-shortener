package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tapadar13/url-shortener/apps/api/internal/auth"
)

type fakeTokenVerifier struct {
	claims auth.TokenClaims
	err    error
}

func (v fakeTokenVerifier) Verify(string) (auth.TokenClaims, error) { return v.claims, v.err }

func TestRequireAuthAddsCurrentUserID(t *testing.T) {
	handler := RequireAuth(fakeTokenVerifier{claims: auth.TokenClaims{UserID: "user-1"}})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := CurrentUserID(r.Context())
		if !ok || userID != "user-1" {
			t.Fatalf("expected authenticated user, got %q", userID)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.Header.Set("Authorization", "Bearer signed-token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("expected protected request to succeed, got %d", response.Code)
	}
}

func TestRequireAuthRejectsMissingOrInvalidBearerToken(t *testing.T) {
	for _, authorization := range []string{"", "Basic credentials", "Bearer"} {
		handler := RequireAuth(fakeTokenVerifier{})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			t.Fatal("expected unauthorized request to stop before handler")
		}))
		request := httptest.NewRequest(http.MethodGet, "/protected", nil)
		request.Header.Set("Authorization", authorization)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 for %q, got %d", authorization, response.Code)
		}
	}
}
