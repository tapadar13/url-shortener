package httpapi

import (
	"context"
	"net/http"
	"net/netip"
	"strconv"
	"time"

	"github.com/tapadar13/url-shortener/apps/api/internal/ratelimit"
)

const (
	rateLimitLimitHeader     = "X-RateLimit-Limit"
	rateLimitRemainingHeader = "X-RateLimit-Remaining"
	rateLimitResetHeader     = "X-RateLimit-Reset"
)

type RequestRateLimiter interface {
	Allow(ctx context.Context, clientKey string) (ratelimit.Result, error)
}

func RateLimit(limiter RequestRateLimiter, trustedProxies ...netip.Prefix) func(http.Handler) http.Handler {
	return rateLimit(limiter, newClientIPResolver(trustedProxies), time.Now)
}

func rateLimit(limiter RequestRateLimiter, resolver clientIPResolver, now func() time.Time) func(http.Handler) http.Handler {
	if now == nil {
		now = time.Now
	}

	return func(next http.Handler) http.Handler {
		if next == nil {
			next = http.NotFoundHandler()
		}

		if limiter == nil {
			return next
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if shouldBypassRateLimit(r) {
				next.ServeHTTP(w, r)
				return
			}

			result, err := limiter.Allow(r.Context(), resolver.resolve(r))
			if err != nil {
				writeError(w, http.StatusServiceUnavailable, "rate_limit_unavailable", "unable to evaluate request rate limit")
				return
			}

			setRateLimitHeaders(w, result)
			if !result.Allowed {
				w.Header().Set("Retry-After", strconv.Itoa(secondsUntil(result.ResetAt, now())))
				writeError(w, http.StatusTooManyRequests, "rate_limit_exceeded", "request rate limit exceeded")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func shouldBypassRateLimit(r *http.Request) bool {
	if r == nil {
		return false
	}

	return r.Method == http.MethodOptions || r.URL.Path == "/healthz" || r.URL.Path == "/readyz"
}

func setRateLimitHeaders(w http.ResponseWriter, result ratelimit.Result) {
	if result.Limit <= 0 {
		return
	}

	w.Header().Set(rateLimitLimitHeader, strconv.Itoa(result.Limit))
	w.Header().Set(rateLimitRemainingHeader, strconv.Itoa(result.Remaining))
	w.Header().Set(rateLimitResetHeader, strconv.FormatInt(result.ResetAt.Unix(), 10))
}

func secondsUntil(resetAt time.Time, now time.Time) int {
	remaining := resetAt.Sub(now)
	if remaining <= 0 {
		return 1
	}

	return int((remaining + time.Second - 1) / time.Second)
}
