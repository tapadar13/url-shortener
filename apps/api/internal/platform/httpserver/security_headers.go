package httpserver

import "net/http"

const (
	contentSecurityPolicy = "default-src 'none'; base-uri 'none'; frame-ancestors 'none'"
	permissionsPolicy     = "camera=(), geolocation=(), microphone=()"
)

func SecurityHeaders(next http.Handler) http.Handler {
	if next == nil {
		next = http.NotFoundHandler()
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", contentSecurityPolicy)
		w.Header().Set("Permissions-Policy", permissionsPolicy)
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")

		next.ServeHTTP(w, r)
	})
}
