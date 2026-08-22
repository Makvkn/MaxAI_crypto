package jwt

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/maxaicrypto/backend/internal/app/config"
)

func TestIssueAndVerifyRoundTrip(t *testing.T) {
	issuer := NewIssuer(config.AuthConfig{
		JWTSecret:      "change-me-development-secret-change-me-please",
		JWTIssuer:      "maxai-crypto",
		AccessTokenTTL: time.Minute,
	})

	userID := uuid.New()
	token, _, err := issuer.Issue(context.Background(), userID)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	got, err := issuer.Verify(context.Background(), token)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if got != userID {
		t.Fatalf("user id = %s, want %s", got, userID)
	}
}
