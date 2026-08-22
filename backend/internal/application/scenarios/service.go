// Package scenarios defines the scenario service. It performs the deterministic
// calculation and does not call the LLM (§171).
package scenarios

import (
	"context"

	"github.com/google/uuid"

	"github.com/maxaicrypto/backend/internal/domain/scenario"
)

// Service validates and computes what-if scenarios (§171).
type Service interface {
	// Simulate validates the request, reads current portfolio facts, performs
	// the deterministic calculation and returns a structured result carrying
	// its calculation ID and version. The LLM only explains this result; it
	// never produces the numbers (§46, §83, §85).
	Simulate(ctx context.Context, userID uuid.UUID, req scenario.Request) (View, error)
	// Compute performs the same deterministic calculation without consuming AI
	// quota. The AI tool loop uses it after a conversation turn already
	// reserved usage.
	Compute(ctx context.Context, userID uuid.UUID, req scenario.Request) (scenario.Result, error)
	// Get returns a previously computed scenario, which lets an AI claim cite
	// the exact calculation behind it (§73).
	Get(ctx context.Context, userID, calculationID uuid.UUID) (scenario.Result, error)
}
