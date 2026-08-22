// Package handlers holds HTTP handlers. Handlers translate between HTTP and
// application services only; they never call providers or the database
// directly (§7, §204).
package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/maxaicrypto/backend/internal/transport/http/response"
)

// Checker reports whether a dependency is usable. The PostgreSQL pool and the
// Redis client both implement it.
type Checker interface {
	Check(ctx context.Context) error
}

// HealthHandler serves liveness and readiness. These endpoints sit outside
// /api/v1 because they are operational, not part of the public API contract.
type HealthHandler struct {
	version    string
	dependency map[string]Checker
}

// NewHealthHandler builds the handler for the named dependencies.
func NewHealthHandler(version string, dependencies map[string]Checker) *HealthHandler {
	return &HealthHandler{version: version, dependency: dependencies}
}

type healthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

type readinessResponse struct {
	Status       string            `json:"status"`
	Version      string            `json:"version"`
	Dependencies map[string]string `json:"dependencies"`
}

// Live reports that the process is running. It intentionally touches no
// dependency, so a database outage never causes the orchestrator to restart
// otherwise healthy instances.
func (h *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	response.OK(w, r, healthResponse{Status: "ok", Version: h.version})
}

// Ready reports whether every dependency required to serve traffic is
// reachable.
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	statuses := make(map[string]string, len(h.dependency))
	ready := true
	for name, checker := range h.dependency {
		if err := checker.Check(ctx); err != nil {
			statuses[name] = "unavailable"
			ready = false
			continue
		}
		statuses[name] = "ok"
	}

	body := readinessResponse{Status: "ready", Version: h.version, Dependencies: statuses}
	status := http.StatusOK
	if !ready {
		body.Status = "not_ready"
		status = http.StatusServiceUnavailable
	}
	response.JSON(w, r, status, body)
}
