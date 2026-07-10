package httpapi

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	urlmodel "github.com/tapadar13/url-shortener/apps/api/internal/url"
	"github.com/tapadar13/url-shortener/apps/api/internal/url/service"
)

type URLCreator interface {
	Create(ctx context.Context, params service.CreateParams) (urlmodel.URL, error)
}

type URLFinder interface {
	GetByShortCode(ctx context.Context, shortCode string) (urlmodel.URL, error)
}

type Dependencies struct {
	URLCreator URLCreator
	URLFinder  URLFinder
}

func NewRouter(dependencies Dependencies) http.Handler {
	router := chi.NewRouter()

	router.Get("/healthz", handleHealth)
	router.Get("/readyz", handleReadiness)

	if dependencies.URLCreator != nil {
		router.Post("/shorten", newCreateURLHandler(dependencies.URLCreator))
	}

	if dependencies.URLFinder != nil {
		router.Get("/shorten/{shortCode}", newGetURLHandler(dependencies.URLFinder))
	}

	return router
}
