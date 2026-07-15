package httpapi

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/tapadar13/url-shortener/apps/api/internal/analytics"
	"github.com/tapadar13/url-shortener/apps/api/internal/metrics"
	urlmodel "github.com/tapadar13/url-shortener/apps/api/internal/url"
	"github.com/tapadar13/url-shortener/apps/api/internal/url/service"
)

type URLCreator interface {
	Create(ctx context.Context, params service.CreateParams) (urlmodel.URL, error)
}

type URLLister interface {
	ListByOwner(ctx context.Context, ownerID string, limit int64) ([]urlmodel.URL, error)
}

type URLFinder interface {
	GetByShortCode(ctx context.Context, shortCode string) (urlmodel.URL, error)
}

type OwnerURLFinder interface {
	GetByShortCodeForOwner(ctx context.Context, ownerID, shortCode string) (urlmodel.URL, error)
}

type URLUpdater interface {
	UpdateLongURL(ctx context.Context, params service.UpdateParams) (urlmodel.URL, error)
}

type OwnerURLUpdater interface {
	UpdateLongURLForOwner(ctx context.Context, ownerID string, params service.UpdateParams) (urlmodel.URL, error)
}

type URLDeleter interface {
	DeleteByShortCode(ctx context.Context, shortCode string) error
}

type OwnerURLDeleter interface {
	DeleteByShortCodeForOwner(ctx context.Context, ownerID, shortCode string) error
}

type URLRedirector interface {
	Resolve(ctx context.Context, shortCode string) (urlmodel.URL, error)
}

type URLAnalyticsReporter interface {
	Get(ctx context.Context, rangeValue analytics.Range) (analytics.Report, error)
}

type Dependencies struct {
	ReadinessChecker    ReadinessChecker
	URLCreator          URLCreator
	URLLister           URLLister
	URLFinder           URLFinder
	URLUpdater          URLUpdater
	URLDeleter          URLDeleter
	URLRedirector       URLRedirector
	AnalyticsReporter   URLAnalyticsReporter
	AnalyticsNow        func() time.Time
	Metrics             *metrics.Metrics
	AuthService         AuthService
	AccessTokenIssuer   AccessTokenIssuer
	AccessTokenVerifier AccessTokenVerifier
	RefreshSessions     RefreshSessionManager
	RedirectStatusCode  int
}

func NewRouter(dependencies Dependencies) http.Handler {
	router := chi.NewRouter()
	if dependencies.Metrics != nil {
		router.Use(metrics.Middleware(dependencies.Metrics))
	}
	router.NotFound(handleNotFound)
	router.MethodNotAllowed(handleMethodNotAllowed)

	router.Get("/healthz", handleHealth)
	router.Get("/readyz", newReadinessHandler(dependencies.ReadinessChecker))
	if dependencies.Metrics != nil {
		router.Get("/metrics", newMetricsHandler(dependencies.Metrics))
	}
	if dependencies.AuthService != nil && dependencies.AccessTokenIssuer != nil {
		router.Post("/auth/register", newRegisterHandler(dependencies.AuthService, dependencies.AccessTokenIssuer, dependencies.RefreshSessions))
		router.Post("/auth/login", newLoginHandler(dependencies.AuthService, dependencies.AccessTokenIssuer, dependencies.RefreshSessions))
		if dependencies.RefreshSessions != nil {
			router.Post("/auth/refresh", newRefreshHandler(dependencies.AccessTokenIssuer, dependencies.RefreshSessions))
		}
	}

	if dependencies.URLCreator != nil {
		createHandler := newCreateURLHandler(dependencies.URLCreator)
		if dependencies.AccessTokenVerifier != nil {
			router.With(RequireAuth(dependencies.AccessTokenVerifier)).Post("/shorten", createHandler)
		} else {
			router.Post("/shorten", createHandler)
		}
	}
	if dependencies.URLLister != nil {
		registerManagementRoute(router, dependencies.AccessTokenVerifier, http.MethodGet, "/shorten", newListURLHandler(dependencies.URLLister))
	}

	if dependencies.URLFinder != nil {
		if dependencies.AnalyticsReporter != nil {
			analyticsHandler := newGetURLAnalyticsHandler(
				dependencies.URLFinder,
				dependencies.AnalyticsReporter,
				dependencies.AnalyticsNow,
			)
			registerManagementRoute(router, dependencies.AccessTokenVerifier, http.MethodGet, "/shorten/{shortCode}/analytics", analyticsHandler)
		}
		registerManagementRoute(router, dependencies.AccessTokenVerifier, http.MethodGet, "/shorten/{shortCode}/stats", newGetURLStatsHandler(dependencies.URLFinder))
		registerManagementRoute(router, dependencies.AccessTokenVerifier, http.MethodGet, "/shorten/{shortCode}", newGetURLHandler(dependencies.URLFinder))
	}

	if dependencies.URLUpdater != nil {
		registerManagementRoute(router, dependencies.AccessTokenVerifier, http.MethodPut, "/shorten/{shortCode}", newUpdateURLHandler(dependencies.URLUpdater))
	}

	if dependencies.URLDeleter != nil {
		registerManagementRoute(router, dependencies.AccessTokenVerifier, http.MethodDelete, "/shorten/{shortCode}", newDeleteURLHandler(dependencies.URLDeleter))
	}

	if dependencies.URLRedirector != nil {
		router.Get("/{shortCode}", newRedirectHandler(dependencies.URLRedirector, dependencies.RedirectStatusCode))
	}

	return router
}

func registerManagementRoute(router *chi.Mux, verifier AccessTokenVerifier, method, path string, handler http.HandlerFunc) {
	if verifier != nil {
		router.With(RequireAuth(verifier)).MethodFunc(method, path, handler)
		return
	}
	router.MethodFunc(method, path, handler)
}

func handleNotFound(w http.ResponseWriter, _ *http.Request) {
	writeError(w, http.StatusNotFound, "not_found", "route was not found")
}

func handleMethodNotAllowed(w http.ResponseWriter, _ *http.Request) {
	writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed for this route")
}
