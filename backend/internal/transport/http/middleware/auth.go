package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"

	appauth "github.com/maxaicrypto/backend/internal/application/auth"
	"github.com/maxaicrypto/backend/internal/domain/apperr"
	"github.com/maxaicrypto/backend/internal/infrastructure/observability"
	"github.com/maxaicrypto/backend/internal/transport/http/response"
)

type userIDContextKey struct{}

// WithUserID stores the authenticated user on the context.
func WithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDContextKey{}, userID)
}

// UserIDFrom returns the authenticated user identifier when present.
func UserIDFrom(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(userIDContextKey{}).(uuid.UUID)
	return userID, ok && userID != uuid.Nil
}

// Authenticate verifies bearer access tokens and attaches the user to the
// request context (§14).
func Authenticate(tokens appauth.TokenIssuer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := strings.TrimSpace(r.Header.Get("Authorization"))
			if header == "" || !strings.HasPrefix(header, "Bearer ") {
				response.Error(w, r, apperr.New(apperr.CodeAuthentication))
				return
			}
			token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
			userID, err := tokens.Verify(r.Context(), token)
			if err != nil {
				response.Error(w, r, apperr.From(err))
				return
			}

			ctx := WithUserID(r.Context(), userID)
			logger := observability.LoggerFrom(ctx).With(slog.String(observability.FieldUserID, userID.String()))
			next.ServeHTTP(w, r.WithContext(observability.WithLogger(ctx, logger)))
		})
	}
}

// OptionalAuthenticate attaches a user when a valid bearer token is present but
// does not reject unauthenticated requests.
func OptionalAuthenticate(tokens appauth.TokenIssuer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := strings.TrimSpace(r.Header.Get("Authorization"))
			if strings.HasPrefix(header, "Bearer ") {
				token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
				if userID, err := tokens.Verify(r.Context(), token); err == nil {
					ctx := WithUserID(r.Context(), userID)
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
