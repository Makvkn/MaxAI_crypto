package shared

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// ErrInvalidCursor reports a cursor that cannot be decoded.
var ErrInvalidCursor = errors.New("invalid cursor")

// Direction is the traversal direction of a paginated query.
type Direction string

const (
	// DirectionForward walks the stable ordering from newest to oldest.
	DirectionForward Direction = "f"
)

// Cursor is an opaque pagination position. It encodes the sort key, a unique
// tie-breaker and the direction, which together reproduce the stable ordering
// `ORDER BY sort_key DESC, id DESC`. Database offsets are never exposed (§109).
type Cursor struct {
	// SortKey is the primary ordering value, typically a timestamp.
	SortKey time.Time `json:"k"`
	// TieBreaker is the unique identifier that disambiguates equal sort keys.
	TieBreaker string `json:"t"`
	// Direction records how the cursor is meant to be traversed.
	Direction Direction `json:"d"`
}

// IsZero reports whether this is the empty cursor that starts a listing.
func (c Cursor) IsZero() bool { return c.TieBreaker == "" && c.SortKey.IsZero() }

// NewCursor builds a forward cursor from a sort key and tie-breaker.
func NewCursor(sortKey time.Time, tieBreaker string) Cursor {
	return Cursor{SortKey: sortKey.UTC(), TieBreaker: tieBreaker, Direction: DirectionForward}
}

// Encode renders the cursor as an opaque URL-safe string.
func (c Cursor) Encode() string {
	if c.IsZero() {
		return ""
	}
	payload, err := json.Marshal(c)
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(payload)
}

// ParseCursor decodes an opaque cursor produced by Encode.
func ParseCursor(raw string) (Cursor, error) {
	if raw == "" {
		return Cursor{}, nil
	}
	payload, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return Cursor{}, fmt.Errorf("%w: %v", ErrInvalidCursor, err)
	}
	var c Cursor
	if err := json.Unmarshal(payload, &c); err != nil {
		return Cursor{}, fmt.Errorf("%w: %v", ErrInvalidCursor, err)
	}
	if c.TieBreaker == "" || c.Direction != DirectionForward {
		return Cursor{}, ErrInvalidCursor
	}
	c.SortKey = c.SortKey.UTC()
	return c, nil
}

// Page is a cursor-paginated result set. The JSON field names match the
// contract the frontend already consumes (§99).
type Page[T any] struct {
	Items      []T     `json:"items"`
	NextCursor *string `json:"next_cursor"`
	HasMore    bool    `json:"has_more"`
}

// NewPage builds a response page. Callers fetch limit+1 rows and pass the
// extra row's cursor so that HasMore is exact rather than guessed.
func NewPage[T any](items []T, next Cursor, hasMore bool) Page[T] {
	if items == nil {
		items = []T{}
	}
	page := Page[T]{Items: items, HasMore: hasMore}
	if hasMore {
		encoded := next.Encode()
		if encoded != "" {
			page.NextCursor = &encoded
		}
	}
	return page
}
