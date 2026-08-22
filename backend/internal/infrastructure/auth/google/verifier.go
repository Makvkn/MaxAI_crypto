// Package google verifies Google ID tokens server-side (§15).
package google

import (
	"context"

	"google.golang.org/api/idtoken"

	"github.com/maxaicrypto/backend/internal/app/config"
	appauth "github.com/maxaicrypto/backend/internal/application/auth"
	"github.com/maxaicrypto/backend/internal/domain/apperr"
)

// Verifier validates Google ID tokens against Google's public keys.
type Verifier struct {
	clientID string
}

// NewVerifier builds a verifier from configuration.
func NewVerifier(cfg config.GoogleConfig) *Verifier {
	return &Verifier{clientID: cfg.ClientID}
}

// Verify implements auth.GoogleVerifier.
func (v *Verifier) Verify(ctx context.Context, idToken string) (appauth.GoogleIdentity, error) {
	if v.clientID == "" {
		return appauth.GoogleIdentity{}, apperr.New(apperr.CodeAuthentication).
			WithMessage("Google sign-in is not configured.")
	}

	payload, err := idtoken.Validate(ctx, idToken, v.clientID)
	if err != nil {
		return appauth.GoogleIdentity{}, apperr.New(apperr.CodeAuthentication).
			WithMessage("The Google sign-in token is invalid.").
			WithCause(err)
	}

	identity := appauth.GoogleIdentity{
		Subject: payload.Subject,
	}
	if email, ok := payload.Claims["email"].(string); ok {
		identity.Email = email
	}
	if verified, ok := payload.Claims["email_verified"].(bool); ok {
		identity.EmailVerified = verified
	}
	if name, ok := payload.Claims["name"].(string); ok && name != "" {
		identity.DisplayName = &name
	}
	if identity.Email == "" {
		return appauth.GoogleIdentity{}, apperr.New(apperr.CodeAuthentication).
			WithMessage("The Google account does not expose an email address.")
	}
	return identity, nil
}

var _ appauth.GoogleVerifier = (*Verifier)(nil)
