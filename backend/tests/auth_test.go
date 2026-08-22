//go:build integration

package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	appauth "github.com/maxaicrypto/backend/internal/application/auth"
	"github.com/maxaicrypto/backend/internal/app/config"
	googleauth "github.com/maxaicrypto/backend/internal/infrastructure/auth/google"
	jwtauth "github.com/maxaicrypto/backend/internal/infrastructure/auth/jwt"
	"github.com/maxaicrypto/backend/internal/infrastructure/auth/password"
	"github.com/maxaicrypto/backend/internal/infrastructure/postgres"
	subscriptionrepo "github.com/maxaicrypto/backend/internal/infrastructure/postgres/subscription"
	userrepo "github.com/maxaicrypto/backend/internal/infrastructure/postgres/user"
	transport "github.com/maxaicrypto/backend/internal/transport/http"
	"github.com/maxaicrypto/backend/internal/transport/http/handlers"
	"github.com/maxaicrypto/backend/tests/testsupport"
)

func TestGuestAuthFlow(t *testing.T) {
	pool := testsupport.Postgres(t)
	router := newAuthTestRouter(t, pool)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/auth/guest", nil))
	if rec.Code != http.StatusCreated {
		t.Fatalf("create guest status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var session struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		User         struct {
			Kind string `json:"kind"`
		} `json:"user"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &session); err != nil {
		t.Fatalf("decode session: %v", err)
	}
	if session.AccessToken == "" || session.RefreshToken == "" {
		t.Fatal("expected token pair")
	}
	if session.User.Kind != "GUEST" {
		t.Fatalf("kind = %q, want GUEST", session.User.Kind)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/session", nil)
	req.Header.Set("Authorization", "Bearer "+session.AccessToken)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("session status = %d, body = %s", rec.Code, rec.Body.String())
	}

	refreshBody, _ := json.Marshal(map[string]string{"refresh_token": session.RefreshToken})
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(refreshBody)))
	if rec.Code != http.StatusOK {
		t.Fatalf("refresh status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func newAuthTestRouter(t *testing.T, pool *postgres.Pool) http.Handler {
	t.Helper()

	cfg := &config.Config{
		HTTP: config.HTTPConfig{
			MaxRequestBytes: 1 << 20,
			AllowedOrigins:  []string{"http://localhost:5173"},
		},
		Auth: config.AuthConfig{
			JWTSecret:       "change-me-development-secret-change-me-please",
			JWTIssuer:       "maxai-crypto",
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 720 * time.Hour,
		},
		AI: config.AIConfig{DailyLimit: 10},
	}

	tx := postgres.NewTxRunner(pool)
	tokens := jwtauth.NewIssuer(cfg.Auth)
	authService := appauth.NewApp(
		userrepo.NewRepository(pool, tx),
		userrepo.NewSessionRepository(pool, tx),
		subscriptionrepo.NewRepository(pool),
		tokens,
		password.NewHasher(),
		googleauth.NewVerifier(cfg.Google),
		cfg,
	)

	return transport.NewRouter(transport.RouterDeps{
		Config: cfg,
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		Health: handlers.NewHealthHandler("test", nil),
		Auth:   handlers.NewAuthHandler(authService),
		Tokens: tokens,
	})
}
