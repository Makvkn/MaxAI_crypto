// Package middleware holds the explicit HTTP middleware chain described in
// §154. Order matters: request identity and panic recovery wrap everything, and
// authentication runs before any handler touches user-owned resources.
package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// HeaderRequestID is the inbound and outbound correlation header.
const HeaderRequestID = "X-Request-Id"

type requestIDKey struct{}

// RequestID assigns every request a correlation identifier, reusing a
// caller-supplied one when present so traces survive across services.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(HeaderRequestID)
		if !isValidRequestID(id) {
			id = uuid.NewString()
		}

		w.Header().Set(HeaderRequestID, id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDKey{}, id)))
	})
}

// RequestIDFrom returns the request identifier stored in ctx, if any.
func RequestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

// isValidRequestID rejects inbound values that could pollute logs or response
// headers, since the header is attacker-controlled.
func isValidRequestID(id string) bool {
	const maxLen = 64
	if id == "" || len(id) > maxLen {
		return false
	}
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
		default:
			return false
		}
	}
	return true
}
