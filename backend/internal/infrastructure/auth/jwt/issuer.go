// Package jwt implements access-token issuance and verification (§13).
package jwt

import (
	"context"
	"fmt"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/maxaicrypto/backend/internal/app/config"
	appauth "github.com/maxaicrypto/backend/internal/application/auth"
	"github.com/maxaicrypto/backend/internal/domain/apperr"
)

// Issuer signs and verifies short-lived access tokens.
type Issuer struct {
	secret []byte
	issuer string
	ttl    time.Duration
}

// NewIssuer builds an issuer from configuration.
func NewIssuer(cfg config.AuthConfig) *Issuer {
	return &Issuer{
		secret: []byte(cfg.JWTSecret),
		issuer: cfg.JWTIssuer,
		ttl:    cfg.AccessTokenTTL,
	}
}

// Issue mints a signed access token for userID.
func (i *Issuer) Issue(_ context.Context, userID uuid.UUID) (string, time.Time, error) {
	expiresAt := time.Now().UTC().Add(i.ttl)
	claims := jwtlib.MapClaims{
		"sub": userID.String(),
		"iss": i.issuer,
		"exp": expiresAt.Unix(),
		"iat": time.Now().UTC().Unix(),
	}
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	signed, err := token.SignedString(i.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign access token: %w", err)
	}
	return signed, expiresAt, nil
}

// Verify validates a presented access token and returns the user identifier.
func (i *Issuer) Verify(_ context.Context, token string) (uuid.UUID, error) {
	parsed, err := jwtlib.Parse(token, func(t *jwtlib.Token) (any, error) {
		if t.Method != jwtlib.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return i.secret, nil
	}, jwtlib.WithIssuer(i.issuer))
	if err != nil || !parsed.Valid {
		return uuid.Nil, apperr.New(apperr.CodeAuthentication)
	}

	claims, ok := parsed.Claims.(jwtlib.MapClaims)
	if !ok {
		return uuid.Nil, apperr.New(apperr.CodeAuthentication)
	}

	sub, err := claims.GetSubject()
	if err != nil || sub == "" {
		return uuid.Nil, apperr.New(apperr.CodeAuthentication)
	}
	userID, err := uuid.Parse(sub)
	if err != nil {
		return uuid.Nil, apperr.New(apperr.CodeAuthentication)
	}
	return userID, nil
}

var _ appauth.TokenIssuer = (*Issuer)(nil)
