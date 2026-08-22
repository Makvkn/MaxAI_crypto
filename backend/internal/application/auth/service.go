// Package auth defines the authentication application service. Guest, Google
// and email authentication all resolve to the same user model, and a guest
// upgrade preserves the original user ID (§12, §165).
package auth

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/maxaicrypto/backend/internal/domain/subscription"
	"github.com/maxaicrypto/backend/internal/domain/user"
)

// Session is an issued token pair together with the authenticated user.
type Session struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	User         Profile
}

// Profile is the user representation the API exposes.
type Profile struct {
	ID           uuid.UUID
	Kind         user.Kind
	Email        *string
	DisplayName  *string
	AuthMethods  []user.AuthProvider
	Subscription subscription.Subscription
	Entitlements subscription.Entitlements
	CreatedAt    time.Time
}

// EmailCredentials carries an email/password pair.
type EmailCredentials struct {
	Email    string
	Password string
}

// GoogleCredentials carries a Google ID token. The backend verifies it against
// Google; a client-supplied user ID is never trusted (§15).
type GoogleCredentials struct {
	IDToken string
}

// UpgradeMethod selects how a guest account is upgraded.
type UpgradeMethod string

const (
	UpgradeEmail  UpgradeMethod = "EMAIL"
	UpgradeGoogle UpgradeMethod = "GOOGLE"
)

// UpgradeRequest promotes the current guest to a registered account.
type UpgradeRequest struct {
	Method UpgradeMethod
	Email  *EmailCredentials
	Google *GoogleCredentials
}

// ClientContext carries request metadata recorded on refresh sessions for
// reuse detection and audit.
type ClientContext struct {
	UserAgent *string
	IPAddress *string
}

// Service owns authentication and the user lifecycle (§165).
type Service interface {
	// CreateGuest provisions an anonymous account and issues a session.
	CreateGuest(ctx context.Context, client ClientContext) (Session, error)
	// RegisterEmail creates a registered account from email credentials.
	RegisterEmail(ctx context.Context, creds EmailCredentials, client ClientContext) (Session, error)
	// LoginEmail authenticates existing email credentials.
	LoginEmail(ctx context.Context, creds EmailCredentials, client ClientContext) (Session, error)
	// LoginGoogle authenticates a verified Google identity, linking it to an
	// existing account when one already owns that identity.
	LoginGoogle(ctx context.Context, creds GoogleCredentials, client ClientContext) (Session, error)
	// Upgrade attaches a permanent identity to the current guest. It runs in
	// one transaction and keeps the same user ID so wallets, snapshots and
	// conversations survive (§12).
	Upgrade(ctx context.Context, userID uuid.UUID, req UpgradeRequest, client ClientContext) (Session, error)
	// Refresh rotates a refresh session and issues a new token pair (§13).
	Refresh(ctx context.Context, refreshToken string, client ClientContext) (Session, error)
	// Logout revokes the presented refresh session.
	Logout(ctx context.Context, refreshToken string) error
	// Session returns the current user's profile.
	Session(ctx context.Context, userID uuid.UUID) (Profile, error)
}

// TokenIssuer mints and verifies access tokens. The concrete format is an
// infrastructure decision the frontend never needs to know (§13).
type TokenIssuer interface {
	Issue(ctx context.Context, userID uuid.UUID) (token string, expiresAt time.Time, err error)
	// Verify returns the user behind a presented access token.
	Verify(ctx context.Context, token string) (uuid.UUID, error)
}

// PasswordHasher hashes and verifies passwords with a modern algorithm.
// Plaintext passwords are never stored (§14).
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(hash, password string) error
}

// GoogleVerifier validates a Google ID token and returns the verified identity.
type GoogleVerifier interface {
	Verify(ctx context.Context, idToken string) (GoogleIdentity, error)
}

// GoogleIdentity is the verified subject of a Google ID token.
type GoogleIdentity struct {
	Subject       string
	Email         string
	EmailVerified bool
	DisplayName   *string
}
