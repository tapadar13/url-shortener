package config

import (
	"errors"
	"fmt"
	"net/netip"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	EnvironmentDevelopment = "development"
	EnvironmentTest        = "test"
	EnvironmentProduction  = "production"

	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"

	LogFormatText = "text"
	LogFormatJSON = "json"
)

type Config struct {
	Environment     string
	HTTP            HTTPConfig
	MongoDB         MongoDBConfig
	ShortCode       ShortCodeConfig
	Redirect        RedirectConfig
	RateLimit       RateLimitConfig
	Log             LogConfig
	RequestTimeout  time.Duration
	ShutdownTimeout time.Duration
}

type HTTPConfig struct {
	Addr              string
	BaseURL           string
	AllowedOrigins    []string
	TrustedProxyCIDRs []netip.Prefix
}

type MongoDBConfig struct {
	URI                  string
	Database             string
	URLsCollection       string
	RateLimitsCollection string
}

type ShortCodeConfig struct {
	Length     int
	MaxRetries int
}

type RedirectConfig struct {
	StatusCode int
}

type RateLimitConfig struct {
	Requests int
	Window   time.Duration
}

type LogConfig struct {
	Level  string
	Format string
}

func Load() (Config, error) {
	return load(os.LookupEnv)
}

func LoadFromMap(values map[string]string) (Config, error) {
	return load(func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	})
}

func load(lookup func(string) (string, bool)) (Config, error) {
	shortCodeLength, err := intValue(lookup, "SHORT_CODE_LENGTH", 7)
	if err != nil {
		return Config{}, err
	}

	shortCodeMaxRetries, err := intValue(lookup, "SHORT_CODE_MAX_RETRIES", 5)
	if err != nil {
		return Config{}, err
	}

	redirectStatus, err := intValue(lookup, "REDIRECT_STATUS", 302)
	if err != nil {
		return Config{}, err
	}

	requestTimeout, err := durationValue(lookup, "REQUEST_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}

	shutdownTimeout, err := durationValue(lookup, "SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}

	rateLimitRequests, err := intValue(lookup, "RATE_LIMIT_REQUESTS", 60)
	if err != nil {
		return Config{}, err
	}

	rateLimitWindow, err := durationValue(lookup, "RATE_LIMIT_WINDOW", time.Minute)
	if err != nil {
		return Config{}, err
	}

	trustedProxyCIDRs, err := prefixListValue(lookup, "TRUSTED_PROXY_CIDRS")
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		Environment: value(lookup, "APP_ENV", EnvironmentDevelopment),
		HTTP: HTTPConfig{
			Addr:              value(lookup, "HTTP_ADDR", ":8080"),
			BaseURL:           trimTrailingSlash(value(lookup, "BASE_URL", "http://localhost:8080")),
			AllowedOrigins:    listValue(lookup, "CORS_ALLOWED_ORIGINS"),
			TrustedProxyCIDRs: trustedProxyCIDRs,
		},
		MongoDB: MongoDBConfig{
			URI:                  value(lookup, "MONGODB_URI", "mongodb://localhost:27017"),
			Database:             value(lookup, "MONGODB_DATABASE", "url_shortener"),
			URLsCollection:       value(lookup, "MONGODB_URLS_COLLECTION", "urls"),
			RateLimitsCollection: value(lookup, "MONGODB_RATE_LIMITS_COLLECTION", "rate_limits"),
		},
		ShortCode: ShortCodeConfig{
			Length:     shortCodeLength,
			MaxRetries: shortCodeMaxRetries,
		},
		Redirect: RedirectConfig{
			StatusCode: redirectStatus,
		},
		RateLimit: RateLimitConfig{
			Requests: rateLimitRequests,
			Window:   rateLimitWindow,
		},
		Log: LogConfig{
			Level:  value(lookup, "LOG_LEVEL", LogLevelInfo),
			Format: value(lookup, "LOG_FORMAT", LogFormatText),
		},
		RequestTimeout:  requestTimeout,
		ShutdownTimeout: shutdownTimeout,
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (cfg Config) Validate() error {
	var errs []error

	if !oneOf(cfg.Environment, EnvironmentDevelopment, EnvironmentTest, EnvironmentProduction) {
		errs = append(errs, fmt.Errorf("APP_ENV must be one of %s, %s, %s", EnvironmentDevelopment, EnvironmentTest, EnvironmentProduction))
	}

	if strings.TrimSpace(cfg.HTTP.Addr) == "" {
		errs = append(errs, errors.New("HTTP_ADDR is required"))
	}

	if !isHTTPURL(cfg.HTTP.BaseURL) {
		errs = append(errs, errors.New("BASE_URL must be a valid http or https URL with a host"))
	}

	for _, origin := range cfg.HTTP.AllowedOrigins {
		if !isHTTPOrigin(origin) {
			errs = append(errs, fmt.Errorf("CORS_ALLOWED_ORIGINS contains invalid origin %q", origin))
		}
	}

	if strings.TrimSpace(cfg.MongoDB.URI) == "" {
		errs = append(errs, errors.New("MONGODB_URI is required"))
	}

	if strings.TrimSpace(cfg.MongoDB.Database) == "" {
		errs = append(errs, errors.New("MONGODB_DATABASE is required"))
	}

	if strings.TrimSpace(cfg.MongoDB.URLsCollection) == "" {
		errs = append(errs, errors.New("MONGODB_URLS_COLLECTION is required"))
	}

	if strings.TrimSpace(cfg.MongoDB.RateLimitsCollection) == "" {
		errs = append(errs, errors.New("MONGODB_RATE_LIMITS_COLLECTION is required"))
	}

	if cfg.ShortCode.Length < 4 || cfg.ShortCode.Length > 32 {
		errs = append(errs, errors.New("SHORT_CODE_LENGTH must be between 4 and 32"))
	}

	if cfg.ShortCode.MaxRetries < 1 || cfg.ShortCode.MaxRetries > 100 {
		errs = append(errs, errors.New("SHORT_CODE_MAX_RETRIES must be between 1 and 100"))
	}

	if !oneOf(strconv.Itoa(cfg.Redirect.StatusCode), "301", "302", "307", "308") {
		errs = append(errs, errors.New("REDIRECT_STATUS must be one of 301, 302, 307, 308"))
	}

	if cfg.RateLimit.Requests < 0 {
		errs = append(errs, errors.New("RATE_LIMIT_REQUESTS must be zero or greater"))
	}

	if cfg.RateLimit.Window <= 0 {
		errs = append(errs, errors.New("RATE_LIMIT_WINDOW must be greater than zero"))
	}

	if !oneOf(cfg.Log.Level, LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelError) {
		errs = append(errs, fmt.Errorf("LOG_LEVEL must be one of %s, %s, %s, %s", LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelError))
	}

	if !oneOf(cfg.Log.Format, LogFormatText, LogFormatJSON) {
		errs = append(errs, fmt.Errorf("LOG_FORMAT must be one of %s, %s", LogFormatText, LogFormatJSON))
	}

	if cfg.RequestTimeout <= 0 {
		errs = append(errs, errors.New("REQUEST_TIMEOUT must be greater than zero"))
	}

	if cfg.ShutdownTimeout <= 0 {
		errs = append(errs, errors.New("SHUTDOWN_TIMEOUT must be greater than zero"))
	}

	return errors.Join(errs...)
}

func value(lookup func(string) (string, bool), key string, fallback string) string {
	raw, ok := lookup(key)
	if !ok {
		return fallback
	}

	return strings.TrimSpace(raw)
}

func listValue(lookup func(string) (string, bool), key string) []string {
	raw, ok := lookup(key)
	if !ok {
		return nil
	}

	var values []string
	for _, item := range strings.Split(raw, ",") {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			values = append(values, trimmed)
		}
	}

	return values
}

func prefixListValue(lookup func(string) (string, bool), key string) ([]netip.Prefix, error) {
	values := listValue(lookup, key)
	prefixes := make([]netip.Prefix, 0, len(values))

	for _, value := range values {
		prefix, err := netip.ParsePrefix(value)
		if err != nil {
			return nil, fmt.Errorf("%s contains invalid CIDR %q", key, value)
		}

		prefixes = append(prefixes, prefix.Masked())
	}

	return prefixes, nil
}

func intValue(lookup func(string) (string, bool), key string, fallback int) (int, error) {
	raw, ok := lookup(key)
	if !ok {
		return fallback, nil
	}

	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return 0, nil
	}

	parsed, err := strconv.Atoi(trimmed)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer", key)
	}

	return parsed, nil
}

func durationValue(lookup func(string) (string, bool), key string, fallback time.Duration) (time.Duration, error) {
	raw, ok := lookup(key)
	if !ok {
		return fallback, nil
	}

	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return 0, nil
	}

	parsed, err := time.ParseDuration(trimmed)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid duration", key)
	}

	return parsed, nil
}

func trimTrailingSlash(value string) string {
	if value == "/" {
		return value
	}

	return strings.TrimRight(value, "/")
}

func isHTTPURL(value string) bool {
	parsed, err := url.Parse(value)
	if err != nil {
		return false
	}

	return (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != ""
}

func isHTTPOrigin(value string) bool {
	parsed, err := url.Parse(value)
	if err != nil {
		return false
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return false
	}

	return parsed.Host != "" &&
		parsed.User == nil &&
		(parsed.Path == "" || parsed.Path == "/") &&
		parsed.RawQuery == "" &&
		parsed.Fragment == ""
}

func oneOf(value string, allowed ...string) bool {
	for _, item := range allowed {
		if value == item {
			return true
		}
	}

	return false
}
