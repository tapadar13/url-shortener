package url

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound           = errors.New("url not found")
	ErrDuplicateShortCode = errors.New("short code already exists")
)

type Repository interface {
	Create(ctx context.Context, record URL) (URL, error)
	FindByShortCode(ctx context.Context, shortCode string) (URL, error)
	FindByShortCodeForOwner(ctx context.Context, ownerID, shortCode string) (URL, error)
	UpdateLongURL(ctx context.Context, params UpdateLongURLParams) (URL, error)
	DeleteByShortCode(ctx context.Context, shortCode string) error
	RecordAccess(ctx context.Context, params RecordAccessParams) (URL, error)
}

type UpdateLongURLParams struct {
	ShortCode string
	LongURL   string
	UpdatedAt time.Time
}

type RecordAccessParams struct {
	ShortCode  string
	AccessedAt time.Time
}
