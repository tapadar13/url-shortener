package httpapi

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
	urlmodel "github.com/tapadar13/url-shortener/apps/api/internal/url"
)

type fakeURLLister struct {
	ownerID string
	limit   int64
	urls    []urlmodel.URL
}

func (l *fakeURLLister) ListByOwner(_ context.Context, ownerID string, limit int64) ([]urlmodel.URL, error) {
	l.ownerID = ownerID
	l.limit = limit
	return l.urls, nil
}

func TestRouterListsAuthenticatedUserURLs(t *testing.T) {
	lister := &fakeURLLister{urls: []urlmodel.URL{{ID: "url-1", ShortCode: "AbC123", LongURL: "https://example.com"}}}
	router := chi.NewRouter()
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), authContextKey{}, "owner-1")))
		})
	})
	router.Get("/shorten", newListURLHandler(lister))

	response := executeRequestWithBody(t, router, http.MethodGet, "/shorten?limit=10", "")
	assertStatus(t, response, http.StatusOK)
	if lister.ownerID != "owner-1" || lister.limit != 10 {
		t.Fatalf("unexpected list request: owner=%q limit=%d", lister.ownerID, lister.limit)
	}
}

func TestListURLHandlerRejectsInvalidLimit(t *testing.T) {
	lister := &fakeURLLister{}
	router := chi.NewRouter()
	router.Get("/shorten", func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), authContextKey{}, "owner-1")
		newListURLHandler(lister)(w, r.WithContext(ctx))
	})

	response := executeRequestWithBody(t, router, http.MethodGet, "/shorten?limit=invalid", "")
	assertStatus(t, response, http.StatusBadRequest)
	if lister.limit != 0 {
		t.Fatal("expected invalid limit to stop before repository call")
	}
}
