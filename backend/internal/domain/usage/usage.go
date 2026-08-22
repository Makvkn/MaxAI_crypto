// Package usage models AI operation accounting. The limit is enforced by the
// backend; the frontend is never a security boundary (§86, §88).
package usage

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/maxaicrypto/backend/internal/domain/subscription"
)

// Operation is a billable AI action. Each one consumes exactly one unit (§86).
type Operation string

const (
	OperationInsight            Operation = "AI_INSIGHT"
	OperationAsk                Operation = "ASK_AI"
	OperationTransactionExplain Operation = "TRANSACTION_EXPLANATION"
	OperationScenario           Operation = "SCENARIO_SIMULATION"
)

// Daily is a user's AI consumption for one UTC day.
type Daily struct {
	UserID uuid.UUID
	// Date is the UTC day boundary the counter belongs to (§86).
	Date      time.Time
	Used      int
	Limit     int
	Plan      subscription.Plan
	ResetsAt  time.Time
	UpdatedAt time.Time
}

// Remaining reports how many operations are still available.
func (d Daily) Remaining() int {
	if d.Used >= d.Limit {
		return 0
	}
	return d.Limit - d.Used
}

// Reservation is a held usage unit. Charging is a reserve-then-settle flow
// rather than check-call-increment, which would be racy (§87).
type Reservation struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Operation Operation
	// IdempotencyKey makes a retried request reuse the same reservation
	// instead of consuming a second unit (§60).
	IdempotencyKey string
	ReservedAt     time.Time
}

// Repository persists durable usage records. Redis holds the fast counter, but
// PostgreSQL remains the source of truth (§6 Principle 6, §64).
type Repository interface {
	// GetDaily returns a user's counter for a UTC day.
	GetDaily(ctx context.Context, userID uuid.UUID, day time.Time) (Daily, error)
	// RecordConsumption durably records a settled operation.
	RecordConsumption(ctx context.Context, userID uuid.UUID, day time.Time, op Operation, idempotencyKey string) error
}

// Counter is the atomic quota primitive backed by Redis. Reserve, Commit and
// Release together implement the atomic charge flow from §87.
type Counter interface {
	// Reserve atomically claims one unit if the limit allows it, returning
	// false when the quota is exhausted.
	Reserve(ctx context.Context, userID uuid.UUID, day time.Time, limit int) (Reservation, bool, error)
	// Commit finalizes a reservation once the operation produced a result.
	Commit(ctx context.Context, reservation Reservation) error
	// Release returns an unused unit when the operation failed before doing
	// meaningful work.
	Release(ctx context.Context, reservation Reservation) error
	// Used reports the current count for a UTC day.
	Used(ctx context.Context, userID uuid.UUID, day time.Time) (int, error)
}
