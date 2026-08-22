// Package usage defines AI quota accounting and plan entitlements (§172, §173).
package usage

import (
	"context"

	"github.com/google/uuid"

	"github.com/maxaicrypto/backend/internal/domain/subscription"
	"github.com/maxaicrypto/backend/internal/domain/usage"
)

// Service enforces the daily AI operation limit (§172). Charging is atomic:
// a unit is reserved before the operation runs and settled afterwards, so
// concurrent requests cannot exceed the quota (§87).
type Service interface {
	// Today returns the user's current quota state for the UTC day.
	Today(ctx context.Context, userID uuid.UUID) (usage.Daily, error)
	// Reserve atomically claims one unit, returning a domain-level
	// limit-reached error when the quota is exhausted.
	Reserve(ctx context.Context, userID uuid.UUID, op usage.Operation, idempotencyKey string) (usage.Reservation, error)
	// Commit settles a reservation once the operation produced a result.
	Commit(ctx context.Context, reservation usage.Reservation) error
	// Release returns a unit when the operation failed before doing meaningful
	// work. The charging policy must stay explicit and consistent (§87).
	Release(ctx context.Context, reservation usage.Reservation) error
}

// EntitlementService resolves what a user's plan allows. Feature checks go
// through it instead of `if user.IsPro` scattered across the codebase (§173).
type EntitlementService interface {
	// Entitlements returns the effective limits for a user.
	Entitlements(ctx context.Context, userID uuid.UUID) (subscription.Entitlements, error)
	// Can reports whether a feature is available to the user.
	Can(ctx context.Context, userID uuid.UUID, feature subscription.Feature) (bool, error)
	// CanCreateWallet reports whether another wallet fits within the plan's
	// wallet limit (§89).
	CanCreateWallet(ctx context.Context, userID uuid.UUID) (bool, error)
}
