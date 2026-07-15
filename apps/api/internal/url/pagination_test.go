package url

import (
	"errors"
	"testing"
	"time"
)

func TestListCursorRoundTrip(t *testing.T) {
	cursor := ListCursor{CreatedAt: time.Date(2026, time.July, 15, 10, 0, 0, 123, time.FixedZone("IST", 19800)), ID: "507f1f77bcf86cd799439011"}
	encoded, err := EncodeListCursor(cursor)
	if err != nil {
		t.Fatalf("encode cursor: %v", err)
	}
	decoded, err := DecodeListCursor(encoded)
	if err != nil {
		t.Fatalf("decode cursor: %v", err)
	}
	if decoded.ID != cursor.ID || !decoded.CreatedAt.Equal(cursor.CreatedAt) || decoded.CreatedAt.Location() != time.UTC {
		t.Fatalf("unexpected decoded cursor: %+v", decoded)
	}
}

func TestDecodeListCursorRejectsMalformedValues(t *testing.T) {
	for _, encoded := range []string{"", "not-base64", "e30"} {
		if _, err := DecodeListCursor(encoded); err == nil {
			t.Fatalf("expected cursor %q to be rejected", encoded)
		}
	}
}

func TestEncodeListCursorValidatesFields(t *testing.T) {
	if _, err := EncodeListCursor(ListCursor{ID: "id"}); !errors.Is(err, ErrCursorTimestampAbsent) {
		t.Fatalf("expected timestamp error, got %v", err)
	}
	if _, err := EncodeListCursor(ListCursor{CreatedAt: time.Now()}); !errors.Is(err, ErrCursorIDRequired) {
		t.Fatalf("expected ID error, got %v", err)
	}
}
