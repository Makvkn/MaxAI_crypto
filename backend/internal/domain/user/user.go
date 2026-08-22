// Package user models accounts and authentication identities. An anonymous
// user is a real user record, not an unauthenticated request (§10).
package user

import (
	"time"

	"github.com/google/uuid"
)

// Kind distinguishes an anonymous account from an upgraded one. The wire
// values match the frontend contract.
type Kind string

const (
	// KindGuest is a real, persisted anonymous account.
	KindGuest Kind = "GUEST"
	// KindRegistered is an account with at least one non-guest identity.
	KindRegistered Kind = "REGISTERED"
)

// User is an account. A guest that later signs in keeps the same ID so all of
// its wallets, snapshots and conversations survive the upgrade (§12).
type User struct {
	ID          uuid.UUID
	Kind        Kind
	Email       *string
	DisplayName *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// AuthProvider identifies how an identity authenticates.
type AuthProvider string

const (
	ProviderGuest  AuthProvider = "guest"
	ProviderGoogle AuthProvider = "google"
	ProviderEmail  AuthProvider = "email"
)

// Identity links a user to one authentication provider. A user record is never
// coupled directly to a single provider, which is what makes guest upgrade and
// account linking possible (§11).
type Identity struct {
	ID       uuid.UUID
	UserID   uuid.UUID
	Provider AuthProvider
	// Subject is the provider's stable identifier for this identity: the
	// Google `sub`, the normalized email, or the generated guest token.
	Subject string
	Email   *string
	// PasswordHash is set only for the email provider. Plaintext passwords are
	// never stored (§14).
	PasswordHash  *string
	EmailVerified bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// RefreshSession is a server-side refresh token record supporting rotation,
// revocation and reuse detection. Only a hash of the secret is persisted (§13).
type RefreshSession struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	TokenHash  string
	IssuedAt   time.Time
	ExpiresAt  time.Time
	RevokedAt  *time.Time
	RotatedTo  *uuid.UUID
	UserAgent  *string
	IPAddress  *string
	LastUsedAt *time.Time
}

// IsActive reports whether the session may still be exchanged for new tokens.
func (s RefreshSession) IsActive(now time.Time) bool {
	return s.RevokedAt == nil && now.Before(s.ExpiresAt)
}
