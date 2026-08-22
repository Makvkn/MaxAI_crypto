// Package portfolio defines the portfolio, snapshot and performance services.
// All three are deterministic: the AI explains their output but never produces
// it (§54, §168, §169).
package portfolio

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/maxaicrypto/backend/internal/domain/performance"
	"github.com/maxaicrypto/backend/internal/domain/portfolio"
)

// Service computes the current portfolio of a wallet (§54). It obtains prices
// through the pricing service and never learns which provider supplied them.
type Service interface {
	// Get returns the current portfolio: valuation, allocation, valuation
	// status, data quality and the positions that could not be valued (§54).
	Get(ctx context.Context, userID, walletID uuid.UUID) (portfolio.Portfolio, error)
}

// SnapshotService builds and persists immutable historical snapshots (§168).
type SnapshotService interface {
	// Create captures the current portfolio as a snapshot, together with its
	// positions and the price metadata that explains the valuation, in one
	// transaction (§117, §162). It is called only after a successful sync;
	// a failed sync produces no snapshot (§50).
	Create(ctx context.Context, walletID uuid.UUID, syncRunID *uuid.UUID, capturedAt time.Time) (portfolio.Snapshot, error)
}

// PerformanceService computes snapshot-based performance (§169). It never
// computes realized or unrealized PnL, cost basis or tax lots (§52).
type PerformanceService interface {
	// Get compares the current valid snapshot with the valid snapshot closest
	// to the start of the period. When no historical snapshot anchors the
	// period the status is UNAVAILABLE rather than a fabricated zero (§53).
	Get(ctx context.Context, userID, walletID uuid.UUID, period performance.Period) (performance.Performance, error)
}
