package httpserver

import (
	"net/http"
	"net/url"
	"strings"
)

const (
	accessControlAllowOrigin    = "Access-Control-Allow-Origin"
	accessControlAllowMethods   = "Access-Control-Allow-Methods"
	accessControlAllowHeaders   = "Access-Control-Allow-Headers"
	accessControlExposeHeaders  = "Access-Control-Expose-Headers"
	accessControlMaxAge         = "Access-Control-Max-Age"
	accessControlRequestMethod  = "Access-Control-Request-Method"
	accessControlRequestHeaders = "Access-Control-Request-Headers"
	allowMethods                = "GET, POST, PUT, DELETE, OPTIONS"
	allowHeaders                = "Content-Type, X-Request-ID"
	exposeHeaders               = "X-Request-ID, X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset, Retry-After"
	preflightMaxAge             = "600"
)

func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		if normalized, ok := normalizeOrigin(origin); ok {
			allowed[normalized] = struct{}{}
		}
	}

	return func(next http.Handler) http.Handler {
		if next == nil {
			next = http.NotFoundHandler()
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin, ok := normalizeOrigin(r.Header.Get("Origin"))
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			if _, ok := allowed[origin]; !ok {
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set(accessControlAllowOrigin, origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set(accessControlExposeHeaders, exposeHeaders)

			if r.Method == http.MethodOptions && r.Header.Get(accessControlRequestMethod) != "" {
				w.Header().Set(accessControlAllowMethods, allowMethods)
				w.Header().Set(accessControlAllowHeaders, allowHeaders)
				w.Header().Set(accessControlMaxAge, preflightMaxAge)
				w.Header().Add("Vary", accessControlRequestMethod)
				w.Header().Add("Vary", accessControlRequestHeaders)
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func normalizeOrigin(value string) (string, bool) {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil {
		return "", false
	}

	scheme := strings.ToLower(parsed.Scheme)
	if (scheme != "http" && scheme != "https") || parsed.Host == "" || parsed.User != nil || (parsed.Path != "" && parsed.Path != "/") || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", false
	}

	return scheme + "://" + strings.ToLower(parsed.Host), true
}
