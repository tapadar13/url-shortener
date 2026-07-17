package mongodb

import (
	"fmt"
	"time"

	urlmodel "github.com/tapadar13/url-shortener/apps/api/internal/url"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type urlDocument struct {
	ID             bson.ObjectID `bson:"_id,omitempty"`
	OwnerID        string        `bson:"owner_id,omitempty"`
	LongURL        string        `bson:"url"`
	ShortCode      string        `bson:"short_code"`
	AccessCount    int64         `bson:"access_count"`
	CreatedAt      time.Time     `bson:"created_at"`
	UpdatedAt      time.Time     `bson:"updated_at"`
	LastAccessedAt *time.Time    `bson:"last_accessed_at,omitempty"`
	ExpiresAt      *time.Time    `bson:"expires_at,omitempty"`
}

func newURLDocument(record urlmodel.URL) (urlDocument, error) {
	var id bson.ObjectID
	if record.ID != "" {
		parsedID, err := bson.ObjectIDFromHex(record.ID)
		if err != nil {
			return urlDocument{}, fmt.Errorf("parse URL id: %w", err)
		}

		id = parsedID
	}

	return urlDocument{
		ID:             id,
		OwnerID:        record.OwnerID,
		LongURL:        record.LongURL,
		ShortCode:      record.ShortCode,
		AccessCount:    record.AccessCount,
		CreatedAt:      record.CreatedAt,
		UpdatedAt:      record.UpdatedAt,
		LastAccessedAt: record.LastAccessedAt,
		ExpiresAt:      record.ExpiresAt,
	}, nil
}

func (doc urlDocument) toDomain() (urlmodel.URL, error) {
	record := urlmodel.URL{
		ID:             doc.ID.Hex(),
		OwnerID:        doc.OwnerID,
		LongURL:        doc.LongURL,
		ShortCode:      doc.ShortCode,
		AccessCount:    doc.AccessCount,
		CreatedAt:      doc.CreatedAt,
		UpdatedAt:      doc.UpdatedAt,
		LastAccessedAt: doc.LastAccessedAt,
		ExpiresAt:      doc.ExpiresAt,
	}

	if doc.ID.IsZero() {
		record.ID = ""
	}

	if err := record.Validate(); err != nil {
		return urlmodel.URL{}, fmt.Errorf("validate URL document: %w", err)
	}

	return record, nil
}
