// Package subscription models plans and entitlements. Feature access is
// resolved through entitlements rather than `if user.IsPro` checks scattered
// across the codebase (§147).
package subscription

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Plan is a subscription tier. The MVP ships FREE only; PRO exists so the
// abstraction does not have to be retrofitted later (§147, §148).
type Plan string

const (
	PlanFree Plan = "FREE"
	PlanPro  Plan = "PRO"
)

// Status is the lifecycle state of a subscription.
type Status string

const (
	StatusActive   Status = "ACTIVE"
	StatusCanceled Status = "CANCELED"
	StatusExpired  Status = "EXPIRED"
)

// Feature is a gated capability (§173).
type Feature string

const (
	FeatureAI       Feature = "AI"
	FeatureScenario Feature = "SCENARIO"
	FeatureWallets  Feature = "WALLETS"
)

// Entitlements are the concrete limits a plan grants.
type Entitlements struct {
	MaxWallets         int
	AIOperationsPerDay int
	Features           []Feature
}

// Allows reports whether a feature is granted.
func (e Entitlements) Allows(feature Feature) bool {
	for _, granted := range e.Features {
		if granted == feature {
			return true
		}
	}
	return false
}

// Subscription is a user's current plan.
type Subscription struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	Plan             Plan
	Status           Status
	CurrentPeriodEnd *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// FreeEntitlements are the MVP free-plan limits (§148). The AI limit is
// overridden by configuration so it can be tuned without a deployment.
var FreeEntitlements = Entitlements{
	MaxWallets:         3,
	AIOperationsPerDay: 10,
	Features:           []Feature{FeatureAI, FeatureScenario, FeatureWallets},
}

// Repository persists subscriptions.
type Repository interface {
	// GetByUser returns a user's subscription, creating nothing. Users without
	// a stored record are treated as free by the entitlement service.
	GetByUser(ctx context.Context, userID uuid.UUID) (Subscription, bool, error)
	Upsert(ctx context.Context, s Subscription) (Subscription, error)
}
