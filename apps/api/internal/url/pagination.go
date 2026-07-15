package url

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var (
	ErrCursorInvalid         = errors.New("pagination cursor is invalid")
	ErrCursorTimestampAbsent = errors.New("pagination cursor timestamp is required")
	ErrCursorIDRequired      = errors.New("pagination cursor ID is required")
)

const maxEncodedCursorLength = 512

type ListCursor struct {
	CreatedAt time.Time
	ID        string
}

type cursorPayload struct {
	Version   int       `json:"v"`
	CreatedAt time.Time `json:"createdAt"`
	ID        string    `json:"id"`
}

func EncodeListCursor(cursor ListCursor) (string, error) {
	normalized, err := normalizeListCursor(cursor)
	if err != nil {
		return "", err
	}
	payload, err := json.Marshal(cursorPayload{Version: 1, CreatedAt: normalized.CreatedAt, ID: normalized.ID})
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func DecodeListCursor(encoded string) (ListCursor, error) {
	if encoded == "" || len(encoded) > maxEncodedCursorLength {
		return ListCursor{}, ErrCursorInvalid
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return ListCursor{}, ErrCursorInvalid
	}
	var payload cursorPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil || payload.Version != 1 {
		return ListCursor{}, ErrCursorInvalid
	}
	return normalizeListCursor(ListCursor{CreatedAt: payload.CreatedAt, ID: payload.ID})
}

func normalizeListCursor(cursor ListCursor) (ListCursor, error) {
	if cursor.CreatedAt.IsZero() {
		return ListCursor{}, ErrCursorTimestampAbsent
	}
	id := strings.TrimSpace(cursor.ID)
	if id == "" {
		return ListCursor{}, ErrCursorIDRequired
	}
	return ListCursor{CreatedAt: cursor.CreatedAt.UTC(), ID: id}, nil
}
