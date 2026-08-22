package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/maxaicrypto/backend/internal/app/config"
	"github.com/maxaicrypto/backend/internal/transport/http/handlers"
	"github.com/maxaicrypto/backend/internal/transport/http/middleware"
)

type stubChecker struct{ err error }

func (s stubChecker) Check(context.Context) error { return s.err }

func newTestRouter(t *testing.T, dependencies map[string]handlers.Checker) http.Handler {
	t.Helper()
	return NewRouter(RouterDeps{
		Config: &config.Config{HTTP: config.HTTPConfig{
			MaxRequestBytes: 1 << 20,
			AllowedOrigins:  []string{"http://localhost:5173"},
		}},
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		Health: handlers.NewHealthHandler("test", dependencies),
	})
}

func TestLivenessIgnoresDependencyFailures(t *testing.T) {
	router := newTestRouter(t, map[string]handlers.Checker{
		"postgres": stubChecker{err: errors.New("down")},
	})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestReadinessFailsWhenADependencyIsDown(t *testing.T) {
	router := newTestRouter(t, map[string]handlers.Checker{
		"postgres": stubChecker{},
		"redis":    stubChecker{err: errors.New("down")},
	})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ready", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}

	var body struct {
		Status       string            `json:"status"`
		Dependencies map[string]string `json:"dependencies"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if body.Dependencies["redis"] != "unavailable" || body.Dependencies["postgres"] != "ok" {
		t.Errorf("dependencies = %v, want redis unavailable and postgres ok", body.Dependencies)
	}
}

func TestUnknownRouteReturnsTheStandardErrorEnvelope(t *testing.T) {
	router := newTestRouter(t, nil)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/does-not-exist", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	var body struct {
		Error struct {
			Code    string         `json:"code"`
			Message string         `json:"message"`
			Details map[string]any `json:"details"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if body.Error.Code != "NOT_FOUND" {
		t.Errorf("error.code = %q, want NOT_FOUND", body.Error.Code)
	}
	if body.Error.Message == "" {
		t.Error("error.message is empty")
	}
	if body.Error.Details == nil {
		t.Error("error.details is missing, want an object")
	}
}

func TestEveryResponseCarriesARequestID(t *testing.T) {
	router := newTestRouter(t, nil)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	if rec.Header().Get(middleware.HeaderRequestID) == "" {
		t.Error("response is missing the request id header")
	}
}

func TestInboundRequestIDIsReusedWhenSafe(t *testing.T) {
	router := newTestRouter(t, nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set(middleware.HeaderRequestID, "abc-123")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if got := rec.Header().Get(middleware.HeaderRequestID); got != "abc-123" {
		t.Errorf("request id = %q, want the inbound value", got)
	}
}

func TestMaliciousInboundRequestIDIsReplaced(t *testing.T) {
	router := newTestRouter(t, nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set(middleware.HeaderRequestID, "abc\r\nX-Injected: 1")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if got := rec.Header().Get(middleware.HeaderRequestID); got == "abc\r\nX-Injected: 1" {
		t.Error("unsafe inbound request id was echoed back")
	}
}

func TestContractRoutesAreRegistered(t *testing.T) {
	router := newTestRouter(t, nil)

	// Routes the frontend already calls must exist, even before their slice is
	// implemented, so an unimplemented feature never looks like a routing bug.
	routes := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/v1/auth/guest"},
		{http.MethodGet, "/api/v1/auth/session"},
		{http.MethodPost, "/api/v1/wallets"},
		{http.MethodGet, "/api/v1/wallets/w1/portfolio"},
		{http.MethodGet, "/api/v1/wallets/w1/performance"},
		{http.MethodGet, "/api/v1/wallets/w1/transactions"},
		{http.MethodGet, "/api/v1/wallets/w1/transactions/t1"},
		{http.MethodGet, "/api/v1/ai/usage"},
		{http.MethodPost, "/api/v1/ai/scenarios"},
		{http.MethodPost, "/api/v1/ai/conversations"},
		{http.MethodPost, "/api/v1/ai/conversations/c1/messages"},
	}
	for _, route := range routes {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(route.method, route.path, nil))

		if rec.Code == http.StatusNotFound {
			t.Errorf("%s %s is not registered", route.method, route.path)
		}
	}
}
