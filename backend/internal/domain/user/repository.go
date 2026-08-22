package user

import (
	"context"

	"github.com/google/uuid"
)

// Repository exposes user and identity persistence in terms of the flows the
// application actually performs (§115).
type Repository interface {
	// CreateGuest persists a new anonymous user together with its guest
	// identity in one transaction.
	CreateGuest(ctx context.Context, subject string) (User, error)
	// CreateRegistered persists a registered account with its first identity and
	// default subscription in one transaction.
	CreateRegistered(ctx context.Context, u User, identity Identity) (User, error)
	// GetByID returns a user by identifier.
	GetByID(ctx context.Context, id uuid.UUID) (User, error)
	// FindByIdentity resolves the user behind a provider subject, returning
	// false when no identity matches.
	FindByIdentity(ctx context.Context, provider AuthProvider, subject string) (User, bool, error)
	// FindIdentity returns the identity record behind a provider subject.
	FindIdentity(ctx context.Context, provider AuthProvider, subject string) (Identity, bool, error)

	// Upgrade attaches a new identity to an existing user and promotes it to
	// KindRegistered. It must be transactional and must never create a second
	// user, which would silently abandon the guest's data (§12).
	Upgrade(ctx context.Context, userID uuid.UUID, identity Identity, displayName *string) (User, error)

	// ListAuthProviders returns the providers a user can authenticate with,
	// which the session response exposes as auth_methods.
	ListAuthProviders(ctx context.Context, userID uuid.UUID) ([]AuthProvider, error)
}

// SessionRepository persists refresh sessions server-side (§13).
type SessionRepository interface {
	Create(ctx context.Context, session RefreshSession) (RefreshSession, error)
	// FindByTokenHash looks a session up by the hash of the presented secret.
	FindByTokenHash(ctx context.Context, tokenHash string) (RefreshSession, bool, error)
	// Rotate revokes the current session and records the successor, which is
	// what makes refresh-token reuse detectable.
	Rotate(ctx context.Context, currentID uuid.UUID, next RefreshSession) (RefreshSession, error)
	// Revoke invalidates a single session, used by logout.
	Revoke(ctx context.Context, id uuid.UUID) error
	// RevokeAllForUser invalidates every session of a user, used when token
	// reuse is detected.
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error
}
