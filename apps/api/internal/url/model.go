package url

import (
	"errors"
	"strings"
	"time"

	"github.com/tapadar13/url-shortener/apps/api/internal/shortcode"
)

var (
	ErrLongURLRequired   = errors.New("long URL is required")
	ErrShortCodeRequired = errors.New("short code is required")
	ErrTimestampRequired = errors.New("timestamp is required")
	ErrNegativeAccesses  = errors.New("access count cannot be negative")
)

type URL struct {
	ID             string
	LongURL        string
	ShortCode      string
	AccessCount    int64
	CreatedAt      time.Time
	UpdatedAt      time.Time
	LastAccessedAt *time.Time
}

type NewParams struct {
	LongURL   string
	ShortCode string
	Now       time.Time
}

func New(params NewParams) (URL, error) {
	longURL := strings.TrimSpace(params.LongURL)
	if normalizedURL, err := NormalizeLongURL(params.LongURL); err == nil {
		longURL = normalizedURL
	}

	shortCode := strings.TrimSpace(params.ShortCode)
	if normalizedShortCode, err := shortcode.Normalize(params.ShortCode); err == nil {
		shortCode = normalizedShortCode
	}

	now := params.Now.UTC()

	record := URL{
		LongURL:     longURL,
		ShortCode:   shortCode,
		AccessCount: 0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := record.Validate(); err != nil {
		return URL{}, err
	}

	return record, nil
}

func (u URL) Validate() error {
	var errs []error

	if strings.TrimSpace(u.LongURL) == "" {
		errs = append(errs, ErrLongURLRequired)
	} else if err := ValidateLongURL(u.LongURL); err != nil {
		errs = append(errs, err)
	}

	if strings.TrimSpace(u.ShortCode) == "" {
		errs = append(errs, ErrShortCodeRequired)
	} else if err := shortcode.Validate(u.ShortCode); err != nil {
		errs = append(errs, err)
	}

	if u.CreatedAt.IsZero() || u.UpdatedAt.IsZero() {
		errs = append(errs, ErrTimestampRequired)
	}

	if u.AccessCount < 0 {
		errs = append(errs, ErrNegativeAccesses)
	}

	return errors.Join(errs...)
}

func (u URL) WithLongURL(longURL string, now time.Time) (URL, error) {
	normalizedURL, err := NormalizeLongURL(longURL)
	if err != nil {
		return URL{}, err
	}

	updated := u
	updated.LongURL = normalizedURL
	updated.UpdatedAt = now.UTC()

	if err := updated.Validate(); err != nil {
		return URL{}, err
	}

	return updated, nil
}

func (u URL) WithAccessRecorded(now time.Time) (URL, error) {
	accessedAt := now.UTC()
	updated := u
	updated.AccessCount++
	updated.LastAccessedAt = &accessedAt

	if err := updated.Validate(); err != nil {
		return URL{}, err
	}

	return updated, nil
}
