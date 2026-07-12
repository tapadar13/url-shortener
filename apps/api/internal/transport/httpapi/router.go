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

type URLUpdater interface {
	UpdateLongURL(ctx context.Context, params service.UpdateParams) (urlmodel.URL, error)
}

type URLDeleter interface {
	DeleteByShortCode(ctx context.Context, shortCode string) error
}

type URLRedirector interface {
	Resolve(ctx context.Context, shortCode string) (urlmodel.URL, error)
}

type Dependencies struct {
	URLCreator         URLCreator
	URLFinder          URLFinder
	URLUpdater         URLUpdater
	URLDeleter         URLDeleter
	URLRedirector      URLRedirector
	RedirectStatusCode int
}

func NewRouter(dependencies Dependencies) http.Handler {
	router := chi.NewRouter()

	router.Get("/healthz", handleHealth)
	router.Get("/readyz", handleReadiness)

	if dependencies.URLCreator != nil {
		router.Post("/shorten", newCreateURLHandler(dependencies.URLCreator))
	}

	if dependencies.URLFinder != nil {
		router.Get("/shorten/{shortCode}/stats", newGetURLStatsHandler(dependencies.URLFinder))
		router.Get("/shorten/{shortCode}", newGetURLHandler(dependencies.URLFinder))
	}

	if dependencies.URLUpdater != nil {
		router.Put("/shorten/{shortCode}", newUpdateURLHandler(dependencies.URLUpdater))
	}

	if dependencies.URLDeleter != nil {
		router.Delete("/shorten/{shortCode}", newDeleteURLHandler(dependencies.URLDeleter))
	}

	if dependencies.URLRedirector != nil {
		router.Get("/{shortCode}", newRedirectHandler(dependencies.URLRedirector, dependencies.RedirectStatusCode))
	}

	return router
}
