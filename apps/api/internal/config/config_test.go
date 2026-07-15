package config

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestLoadFromMapUsesDefaults(t *testing.T) {
	t.Parallel()

	cfg, err := LoadFromMap(nil)
	if err != nil {
		t.Fatalf("expected defaults to be valid: %v", err)
	}

	if cfg.Environment != EnvironmentDevelopment {
		t.Fatalf("expected default environment %q, got %q", EnvironmentDevelopment, cfg.Environment)
	}

	if cfg.HTTP.Addr != ":8080" {
		t.Fatalf("expected default HTTP address, got %q", cfg.HTTP.Addr)
	}

	if cfg.HTTP.BaseURL != "http://localhost:8080" {
		t.Fatalf("expected default base URL, got %q", cfg.HTTP.BaseURL)
	}

	if len(cfg.HTTP.AllowedOrigins) != 0 {
		t.Fatalf("expected no default CORS origins, got %#v", cfg.HTTP.AllowedOrigins)
	}

	if len(cfg.HTTP.TrustedProxyCIDRs) != 0 {
		t.Fatalf("expected no trusted proxies by default, got %#v", cfg.HTTP.TrustedProxyCIDRs)
	}

	if cfg.MongoDB.Database != "url_shortener" {
		t.Fatalf("expected default MongoDB database, got %q", cfg.MongoDB.Database)
	}

	if cfg.MongoDB.RateLimitsCollection != "rate_limits" {
		t.Fatalf("expected default rate limit collection, got %q", cfg.MongoDB.RateLimitsCollection)
	}

	if cfg.MongoDB.AnalyticsCollection != "click_analytics" {
		t.Fatalf("expected default analytics collection, got %q", cfg.MongoDB.AnalyticsCollection)
	}

	if cfg.Redis.URL != "redis://localhost:6379/0" || cfg.Redis.KeyPrefix != "url-shortener" || cfg.Redis.ConnectTimeout != 5*time.Second {
		t.Fatalf("expected default Redis config, got %+v", cfg.Redis)
	}

	if cfg.ShortCode.Length != 7 {
		t.Fatalf("expected default short code length 7, got %d", cfg.ShortCode.Length)
	}

	if cfg.Redirect.StatusCode != 302 {
		t.Fatalf("expected default redirect status 302, got %d", cfg.Redirect.StatusCode)
	}

	if cfg.RedirectCache.Enabled ||
		cfg.RedirectCache.TTL != 10*time.Minute ||
		cfg.RedirectCache.AccessWorkers != 2 ||
		cfg.RedirectCache.AccessQueueSize != 1024 ||
		cfg.RedirectCache.AccessTimeout != 5*time.Second {
		t.Fatalf("expected disabled redirect cache with default TTL, got %+v", cfg.RedirectCache)
	}

	if cfg.RateLimit.Requests != 60 || cfg.RateLimit.Window != time.Minute {
		t.Fatalf("expected default rate limit config, got %+v", cfg.RateLimit)
	}

	if cfg.Analytics.Workers != 2 || cfg.Analytics.QueueSize != 4096 || cfg.Analytics.WriteTimeout != 5*time.Second {
		t.Fatalf("expected default analytics config, got %+v", cfg.Analytics)
	}

	if cfg.RequestTimeout != 10*time.Second {
		t.Fatalf("expected default request timeout 10s, got %s", cfg.RequestTimeout)
	}
	if cfg.MaxRequestBodyBytes != 1<<20 {
		t.Fatalf("expected default request body limit 1 MiB, got %d", cfg.MaxRequestBodyBytes)
	}
}

func TestLoadFromMapAppliesOverrides(t *testing.T) {
	t.Parallel()

	cfg, err := LoadFromMap(map[string]string{
		"APP_ENV":                          EnvironmentProduction,
		"HTTP_ADDR":                        ":9090",
		"BASE_URL":                         "https://sho.rt/",
		"CORS_ALLOWED_ORIGINS":             "http://localhost:3000, https://app.example.com",
		"TRUSTED_PROXY_CIDRS":              "10.0.0.0/8, 2001:db8::/32",
		"MONGODB_URI":                      "mongodb://mongo:27017",
		"MONGODB_DATABASE":                 "links",
		"MONGODB_URLS_COLLECTION":          "short_urls",
		"MONGODB_RATE_LIMITS_COLLECTION":   "limits",
		"MONGODB_ANALYTICS_COLLECTION":     "daily_clicks",
		"REDIS_URL":                        "rediss://cache-user:secret@redis.example.com:6380/2",
		"REDIS_KEY_PREFIX":                 "shortener-production",
		"REDIS_CONNECT_TIMEOUT":            "2s",
		"SHORT_CODE_LENGTH":                "9",
		"SHORT_CODE_MAX_RETRIES":           "12",
		"REDIRECT_STATUS":                  "307",
		"REDIRECT_CACHE_ENABLED":           "true",
		"REDIRECT_CACHE_TTL":               "15m",
		"REDIRECT_CACHE_ACCESS_WORKERS":    "4",
		"REDIRECT_CACHE_ACCESS_QUEUE_SIZE": "2048",
		"REDIRECT_CACHE_ACCESS_TIMEOUT":    "3s",
		"ANALYTICS_WORKERS":                "4",
		"ANALYTICS_QUEUE_SIZE":             "8192",
		"ANALYTICS_WRITE_TIMEOUT":          "2s",
		"RATE_LIMIT_REQUESTS":              "120",
		"RATE_LIMIT_WINDOW":                "5m",
		"LOG_LEVEL":                        LogLevelWarn,
		"LOG_FORMAT":                       LogFormatJSON,
		"REQUEST_TIMEOUT":                  "3s",
		"SHUTDOWN_TIMEOUT":                 "15s",
		"MAX_REQUEST_BODY_BYTES":           "2097152",
		"AUTH_TOKEN_SECRET":                "production-secret-value-0123456789",
		"AUTH_TOKEN_ISSUER":                "test-issuer",
		"AUTH_TOKEN_AUDIENCE":              "test-audience",
		"AUTH_TOKEN_TTL":                   "30m",
	})
	if err != nil {
		t.Fatalf("expected overrides to be valid: %v", err)
	}

	if cfg.Environment != EnvironmentProduction {
		t.Fatalf("expected production environment, got %q", cfg.Environment)
	}

	if cfg.HTTP.Addr != ":9090" {
		t.Fatalf("expected overridden HTTP address, got %q", cfg.HTTP.Addr)
	}

	if cfg.HTTP.BaseURL != "https://sho.rt" {
		t.Fatalf("expected trailing slash to be trimmed, got %q", cfg.HTTP.BaseURL)
	}

	if len(cfg.HTTP.AllowedOrigins) != 2 || cfg.HTTP.AllowedOrigins[0] != "http://localhost:3000" || cfg.HTTP.AllowedOrigins[1] != "https://app.example.com" {
		t.Fatalf("expected configured CORS origins, got %#v", cfg.HTTP.AllowedOrigins)
	}

	if len(cfg.HTTP.TrustedProxyCIDRs) != 2 || cfg.HTTP.TrustedProxyCIDRs[0].String() != "10.0.0.0/8" || cfg.HTTP.TrustedProxyCIDRs[1].String() != "2001:db8::/32" {
		t.Fatalf("expected configured trusted proxy CIDRs, got %#v", cfg.HTTP.TrustedProxyCIDRs)
	}

	if cfg.MongoDB.URI != "mongodb://mongo:27017" {
		t.Fatalf("expected overridden MongoDB URI, got %q", cfg.MongoDB.URI)
	}

	if cfg.MongoDB.URLsCollection != "short_urls" {
		t.Fatalf("expected overridden MongoDB collection, got %q", cfg.MongoDB.URLsCollection)
	}

	if cfg.MongoDB.RateLimitsCollection != "limits" {
		t.Fatalf("expected overridden rate limit collection, got %q", cfg.MongoDB.RateLimitsCollection)
	}

	if cfg.MongoDB.AnalyticsCollection != "daily_clicks" {
		t.Fatalf("expected overridden analytics collection, got %q", cfg.MongoDB.AnalyticsCollection)
	}

	if cfg.Redis.URL != "rediss://cache-user:secret@redis.example.com:6380/2" || cfg.Redis.KeyPrefix != "shortener-production" || cfg.Redis.ConnectTimeout != 2*time.Second {
		t.Fatalf("expected overridden Redis config, got %+v", cfg.Redis)
	}

	if cfg.ShortCode.Length != 9 || cfg.ShortCode.MaxRetries != 12 {
		t.Fatalf("expected overridden short code config, got %+v", cfg.ShortCode)
	}

	if cfg.Redirect.StatusCode != 307 {
		t.Fatalf("expected redirect status 307, got %d", cfg.Redirect.StatusCode)
	}

	if !cfg.RedirectCache.Enabled ||
		cfg.RedirectCache.TTL != 15*time.Minute ||
		cfg.RedirectCache.AccessWorkers != 4 ||
		cfg.RedirectCache.AccessQueueSize != 2048 ||
		cfg.RedirectCache.AccessTimeout != 3*time.Second {
		t.Fatalf("expected overridden redirect cache config, got %+v", cfg.RedirectCache)
	}

	if cfg.RateLimit.Requests != 120 || cfg.RateLimit.Window != 5*time.Minute {
		t.Fatalf("expected overridden rate limit config, got %+v", cfg.RateLimit)
	}

	if cfg.Analytics.Workers != 4 || cfg.Analytics.QueueSize != 8192 || cfg.Analytics.WriteTimeout != 2*time.Second {
		t.Fatalf("expected overridden analytics config, got %+v", cfg.Analytics)
	}

	if cfg.Log.Level != LogLevelWarn || cfg.Log.Format != LogFormatJSON {
		t.Fatalf("expected overridden log config, got %+v", cfg.Log)
	}

	if cfg.RequestTimeout != 3*time.Second || cfg.ShutdownTimeout != 15*time.Second {
		t.Fatalf("expected overridden timeouts, got request=%s shutdown=%s", cfg.RequestTimeout, cfg.ShutdownTimeout)
	}
	if cfg.MaxRequestBodyBytes != 2<<20 {
		t.Fatalf("expected overridden request body limit 2 MiB, got %d", cfg.MaxRequestBodyBytes)
	}
	if cfg.Auth.TokenSecret != "production-secret-value-0123456789" || cfg.Auth.TokenIssuer != "test-issuer" || cfg.Auth.TokenAudience != "test-audience" || cfg.Auth.TokenTTL != 30*time.Minute {
		t.Fatalf("expected overridden auth config, got %+v", cfg.Auth)
	}
}

func TestLoadFromMapRejectsInvalidRequestBodyLimit(t *testing.T) {
	t.Parallel()

	_, err := LoadFromMap(map[string]string{"MAX_REQUEST_BODY_BYTES": "0"})
	if err == nil || !strings.Contains(err.Error(), "MAX_REQUEST_BODY_BYTES") {
		t.Fatalf("expected invalid request body limit error, got %v", err)
	}
}

func TestLoadFromMapRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	cfg, err := LoadFromMap(map[string]string{
		"APP_ENV":                          "staging",
		"HTTP_ADDR":                        " ",
		"BASE_URL":                         "localhost:8080",
		"CORS_ALLOWED_ORIGINS":             "ftp://example.com",
		"MONGODB_URI":                      " ",
		"MONGODB_DATABASE":                 " ",
		"MONGODB_URLS_COLLECTION":          " ",
		"MONGODB_RATE_LIMITS_COLLECTION":   " ",
		"MONGODB_ANALYTICS_COLLECTION":     " ",
		"REDIS_URL":                        "https://redis.example.com",
		"REDIS_KEY_PREFIX":                 " ",
		"REDIS_CONNECT_TIMEOUT":            "0s",
		"SHORT_CODE_LENGTH":                "3",
		"SHORT_CODE_MAX_RETRIES":           "0",
		"REDIRECT_STATUS":                  "200",
		"REDIRECT_CACHE_TTL":               "0s",
		"REDIRECT_CACHE_ACCESS_WORKERS":    "65",
		"REDIRECT_CACHE_ACCESS_QUEUE_SIZE": "100001",
		"REDIRECT_CACHE_ACCESS_TIMEOUT":    "0s",
		"ANALYTICS_WORKERS":                "65",
		"ANALYTICS_QUEUE_SIZE":             "100001",
		"ANALYTICS_WRITE_TIMEOUT":          "0s",
		"RATE_LIMIT_REQUESTS":              "-1",
		"RATE_LIMIT_WINDOW":                "0s",
		"LOG_LEVEL":                        "trace",
		"LOG_FORMAT":                       "pretty",
		"REQUEST_TIMEOUT":                  "0s",
		"SHUTDOWN_TIMEOUT":                 "-1s",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}

	if !reflect.DeepEqual(cfg, Config{}) {
		t.Fatalf("expected empty config on validation failure, got %+v", cfg)
	}

	message := err.Error()
	for _, expected := range []string{
		"APP_ENV",
		"HTTP_ADDR",
		"BASE_URL",
		"CORS_ALLOWED_ORIGINS",
		"MONGODB_URI",
		"MONGODB_DATABASE",
		"MONGODB_URLS_COLLECTION",
		"MONGODB_RATE_LIMITS_COLLECTION",
		"MONGODB_ANALYTICS_COLLECTION",
		"REDIS_URL",
		"REDIS_KEY_PREFIX",
		"REDIS_CONNECT_TIMEOUT",
		"SHORT_CODE_LENGTH",
		"SHORT_CODE_MAX_RETRIES",
		"REDIRECT_STATUS",
		"REDIRECT_CACHE_TTL",
		"REDIRECT_CACHE_ACCESS_WORKERS",
		"REDIRECT_CACHE_ACCESS_QUEUE_SIZE",
		"REDIRECT_CACHE_ACCESS_TIMEOUT",
		"ANALYTICS_WORKERS",
		"ANALYTICS_QUEUE_SIZE",
		"ANALYTICS_WRITE_TIMEOUT",
		"RATE_LIMIT_REQUESTS",
		"RATE_LIMIT_WINDOW",
		"LOG_LEVEL",
		"LOG_FORMAT",
		"REQUEST_TIMEOUT",
		"SHUTDOWN_TIMEOUT",
	} {
		if !strings.Contains(message, expected) {
			t.Fatalf("expected validation error to include %s, got %q", expected, message)
		}
	}
}

func TestLoadFromMapRejectsUnparseableValues(t *testing.T) {
	t.Parallel()

	_, err := LoadFromMap(map[string]string{
		"SHORT_CODE_LENGTH":      "invalid",
		"SHORT_CODE_MAX_RETRIES": "invalid",
		"REDIRECT_STATUS":        "invalid",
		"RATE_LIMIT_REQUESTS":    "invalid",
		"RATE_LIMIT_WINDOW":      "invalid",
		"REQUEST_TIMEOUT":        "invalid",
		"SHUTDOWN_TIMEOUT":       "invalid",
	})
	if err == nil {
		t.Fatal("expected parsing error")
	}

	if !strings.Contains(err.Error(), "SHORT_CODE_LENGTH") {
		t.Fatalf("expected parsing error to include SHORT_CODE_LENGTH, got %q", err.Error())
	}
}

func TestLoadFromMapRejectsInvalidTrustedProxyCIDR(t *testing.T) {
	t.Parallel()

	_, err := LoadFromMap(map[string]string{
		"TRUSTED_PROXY_CIDRS": "10.0.0.0/8, not-a-cidr",
	})
	if err == nil {
		t.Fatal("expected parsing error")
	}

	if !strings.Contains(err.Error(), `TRUSTED_PROXY_CIDRS contains invalid CIDR "not-a-cidr"`) {
		t.Fatalf("expected invalid trusted proxy CIDR error, got %q", err.Error())
	}
}

func TestLoadFromMapRejectsInvalidRedirectCacheBoolean(t *testing.T) {
	t.Parallel()

	_, err := LoadFromMap(map[string]string{
		"REDIRECT_CACHE_ENABLED": "sometimes",
	})
	if err == nil {
		t.Fatal("expected parsing error")
	}

	if !strings.Contains(err.Error(), "REDIRECT_CACHE_ENABLED must be a boolean") {
		t.Fatalf("expected invalid redirect cache boolean error, got %q", err.Error())
	}
}

func TestLoadFromMapRejectsUnparseableRedirectCacheWorkers(t *testing.T) {
	t.Parallel()

	_, err := LoadFromMap(map[string]string{
		"REDIRECT_CACHE_ACCESS_WORKERS": "many",
	})
	if err == nil {
		t.Fatal("expected parsing error")
	}

	if !strings.Contains(err.Error(), "REDIRECT_CACHE_ACCESS_WORKERS must be an integer") {
		t.Fatalf("expected invalid access workers error, got %q", err.Error())
	}
}

func TestLoadFromMapRejectsUnparseableAnalyticsValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		key   string
		value string
	}{
		{name: "workers", key: "ANALYTICS_WORKERS", value: "many"},
		{name: "queue size", key: "ANALYTICS_QUEUE_SIZE", value: "large"},
		{name: "write timeout", key: "ANALYTICS_WRITE_TIMEOUT", value: "slow"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := LoadFromMap(map[string]string{tt.key: tt.value})
			if err == nil || !strings.Contains(err.Error(), tt.key) {
				t.Fatalf("expected parsing error to include %s, got %v", tt.key, err)
			}
		})
	}
}
