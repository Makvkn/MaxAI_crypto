package portfolio

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/maxaicrypto/backend/internal/domain/price"
	"github.com/maxaicrypto/backend/internal/domain/shared"
)

// Snapshot is an immutable historical portfolio state. Snapshots are never
// mutated to reflect later prices (§49).
type Snapshot struct {
	ID            uuid.UUID
	WalletID      uuid.UUID
	CapturedAt    time.Time
	TotalValueUSD shared.NullDecimal
	Status        shared.ValuationStatus
	DataQuality   shared.DataQuality
	// CalculationVersion records which algorithm produced this snapshot (§51).
	CalculationVersion int
	SyncRunID          *uuid.UUID
	CreatedAt          time.Time
}

// IsValid reports whether the snapshot can serve as a performance endpoint. A
// snapshot without a total cannot anchor a percentage change (§53).
func (s Snapshot) IsValid() bool {
	return s.TotalValueUSD.Valid && s.Status != shared.ValuationUnavailable
}

// SnapshotPosition is one valued holding captured inside a snapshot. Price
// metadata is captured alongside the value so the valuation stays auditable
// after the fact (§161, §162).
type SnapshotPosition struct {
	SnapshotID     uuid.UUID
	AssetID        uuid.UUID
	Balance        shared.Decimal
	PriceUSD       shared.NullDecimal
	ValueUSD       shared.NullDecimal
	AllocationPct  shared.NullDecimal
	PriceTimestamp *time.Time
	PriceSource    *price.Source
}

// SnapshotRepository persists historical snapshots.
type SnapshotRepository interface {
	// Create writes a snapshot together with all of its positions in one
	// transaction; a partially written snapshot is not a valid history entry
	// (§117).
	Create(ctx context.Context, snapshot Snapshot, positions []SnapshotPosition) (Snapshot, error)
	// GetLatestValid returns the most recent snapshot usable for valuation.
	GetLatestValid(ctx context.Context, walletID uuid.UUID) (Snapshot, bool, error)
	// GetClosestValidBefore returns the valid snapshot nearest to a past
	// instant, which anchors the opening value of a performance period (§52).
	GetClosestValidBefore(ctx context.Context, walletID uuid.UUID, at time.Time) (Snapshot, bool, error)
	// GetFirstValid returns the earliest valid snapshot, used by ALL_TIME.
	GetFirstValid(ctx context.Context, walletID uuid.UUID) (Snapshot, bool, error)
	// ListBetween returns snapshots in a time range, ordered oldest first, to
	// build the historical chart series.
	ListBetween(ctx context.Context, walletID uuid.UUID, from, to time.Time, limit int) ([]Snapshot, error)
	// ListPositions returns the captured positions of a snapshot.
	ListPositions(ctx context.Context, snapshotID uuid.UUID) ([]SnapshotPosition, error)
	// GetBySyncRunID returns a snapshot created for a sync run, which makes
	// snapshot creation idempotent across job retries (§60).
	GetBySyncRunID(ctx context.Context, syncRunID uuid.UUID) (Snapshot, error)
}
