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

	if cfg.MongoDB.Database != "url_shortener" {
		t.Fatalf("expected default MongoDB database, got %q", cfg.MongoDB.Database)
	}

	if cfg.MongoDB.RateLimitsCollection != "rate_limits" {
		t.Fatalf("expected default rate limit collection, got %q", cfg.MongoDB.RateLimitsCollection)
	}

	if cfg.ShortCode.Length != 7 {
		t.Fatalf("expected default short code length 7, got %d", cfg.ShortCode.Length)
	}

	if cfg.Redirect.StatusCode != 302 {
		t.Fatalf("expected default redirect status 302, got %d", cfg.Redirect.StatusCode)
	}

	if cfg.RateLimit.Requests != 60 || cfg.RateLimit.Window != time.Minute {
		t.Fatalf("expected default rate limit config, got %+v", cfg.RateLimit)
	}

	if cfg.RequestTimeout != 10*time.Second {
		t.Fatalf("expected default request timeout 10s, got %s", cfg.RequestTimeout)
	}
}

func TestLoadFromMapAppliesOverrides(t *testing.T) {
	t.Parallel()

	cfg, err := LoadFromMap(map[string]string{
		"APP_ENV":                        EnvironmentProduction,
		"HTTP_ADDR":                      ":9090",
		"BASE_URL":                       "https://sho.rt/",
		"CORS_ALLOWED_ORIGINS":           "http://localhost:3000, https://app.example.com",
		"MONGODB_URI":                    "mongodb://mongo:27017",
		"MONGODB_DATABASE":               "links",
		"MONGODB_URLS_COLLECTION":        "short_urls",
		"MONGODB_RATE_LIMITS_COLLECTION": "limits",
		"SHORT_CODE_LENGTH":              "9",
		"SHORT_CODE_MAX_RETRIES":         "12",
		"REDIRECT_STATUS":                "307",
		"RATE_LIMIT_REQUESTS":            "120",
		"RATE_LIMIT_WINDOW":              "5m",
		"LOG_LEVEL":                      LogLevelWarn,
		"LOG_FORMAT":                     LogFormatJSON,
		"REQUEST_TIMEOUT":                "3s",
		"SHUTDOWN_TIMEOUT":               "15s",
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

	if cfg.MongoDB.URI != "mongodb://mongo:27017" {
		t.Fatalf("expected overridden MongoDB URI, got %q", cfg.MongoDB.URI)
	}

	if cfg.MongoDB.URLsCollection != "short_urls" {
		t.Fatalf("expected overridden MongoDB collection, got %q", cfg.MongoDB.URLsCollection)
	}

	if cfg.MongoDB.RateLimitsCollection != "limits" {
		t.Fatalf("expected overridden rate limit collection, got %q", cfg.MongoDB.RateLimitsCollection)
	}

	if cfg.ShortCode.Length != 9 || cfg.ShortCode.MaxRetries != 12 {
		t.Fatalf("expected overridden short code config, got %+v", cfg.ShortCode)
	}

	if cfg.Redirect.StatusCode != 307 {
		t.Fatalf("expected redirect status 307, got %d", cfg.Redirect.StatusCode)
	}

	if cfg.RateLimit.Requests != 120 || cfg.RateLimit.Window != 5*time.Minute {
		t.Fatalf("expected overridden rate limit config, got %+v", cfg.RateLimit)
	}

	if cfg.Log.Level != LogLevelWarn || cfg.Log.Format != LogFormatJSON {
		t.Fatalf("expected overridden log config, got %+v", cfg.Log)
	}

	if cfg.RequestTimeout != 3*time.Second || cfg.ShutdownTimeout != 15*time.Second {
		t.Fatalf("expected overridden timeouts, got request=%s shutdown=%s", cfg.RequestTimeout, cfg.ShutdownTimeout)
	}
}

func TestLoadFromMapRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	cfg, err := LoadFromMap(map[string]string{
		"APP_ENV":                        "staging",
		"HTTP_ADDR":                      " ",
		"BASE_URL":                       "localhost:8080",
		"CORS_ALLOWED_ORIGINS":           "ftp://example.com",
		"MONGODB_URI":                    " ",
		"MONGODB_DATABASE":               " ",
		"MONGODB_URLS_COLLECTION":        " ",
		"MONGODB_RATE_LIMITS_COLLECTION": " ",
		"SHORT_CODE_LENGTH":              "3",
		"SHORT_CODE_MAX_RETRIES":         "0",
		"REDIRECT_STATUS":                "200",
		"RATE_LIMIT_REQUESTS":            "-1",
		"RATE_LIMIT_WINDOW":              "0s",
		"LOG_LEVEL":                      "trace",
		"LOG_FORMAT":                     "pretty",
		"REQUEST_TIMEOUT":                "0s",
		"SHUTDOWN_TIMEOUT":               "-1s",
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
		"SHORT_CODE_LENGTH",
		"SHORT_CODE_MAX_RETRIES",
		"REDIRECT_STATUS",
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
