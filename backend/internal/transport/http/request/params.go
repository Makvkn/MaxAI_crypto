package request

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/maxaicrypto/backend/internal/domain/apperr"
	"github.com/maxaicrypto/backend/internal/domain/shared"
)

// Pagination limits. The API is cursor-based only; page numbers are never
// accepted (§109).
const (
	DefaultPageLimit = 50
	MaxPageLimit     = 100
)

// Page holds validated cursor pagination parameters.
type Page struct {
	Limit  int
	Cursor shared.Cursor
}

// ParsePage reads limit and cursor from the query string.
func ParsePage(r *http.Request) (Page, error) {
	page := Page{Limit: DefaultPageLimit}

	if raw := r.URL.Query().Get("limit"); raw != "" {
		limit, err := strconv.Atoi(raw)
		if err != nil || limit < 1 {
			return Page{}, apperr.New(apperr.CodeValidation).
				WithMessage("The limit parameter must be a positive integer.").
				WithDetail("fields", map[string]string{"limit": "must be a positive integer"})
		}
		page.Limit = min(limit, MaxPageLimit)
	}

	if raw := r.URL.Query().Get("cursor"); raw != "" {
		cursor, err := shared.ParseCursor(raw)
		if err != nil {
			return Page{}, apperr.New(apperr.CodeValidation).
				WithMessage("The cursor parameter is not valid.").
				WithDetail("fields", map[string]string{"cursor": "is not a valid cursor"})
		}
		page.Cursor = cursor
	}

	return page, nil
}

// UUIDParam parses a path parameter that must be a UUID. Identifiers are opaque
// to clients, so a malformed value is a validation error rather than a 404.
func UUIDParam(raw, field string) (uuid.UUID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, apperr.New(apperr.CodeValidation).
			WithMessage("The "+field+" parameter is not a valid identifier.").
			WithDetail("fields", map[string]string{field: "must be a valid identifier"})
	}
	return id, nil
}

// QueryString returns a trimmed optional query parameter.
func QueryString(r *http.Request, key string) (string, bool) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return "", false
	}
	return value, true
}
