// Package sync defines the wallet synchronization service. It coordinates the
// provider resolver, normalization, persistence, pricing and snapshot creation;
// actual execution happens inside background jobs (§167).
package sync

import (
	"context"

	"github.com/google/uuid"

	"github.com/maxaicrypto/backend/internal/domain/wallet"
)

// Request describes one synchronization run.
type Request struct {
	WalletID uuid.UUID
	Trigger  wallet.SyncTrigger
	// JobID is deterministic per attempt so a retried job resumes the same run
	// instead of opening a second one (§60).
	JobID string
}

// Result reports the outcome of a synchronization run.
type Result struct {
	Status wallet.SyncStatus
	// SnapshotID is set only when the run produced a new snapshot. A failed
	// sync creates none and leaves the previous valid snapshot intact (§50).
	SnapshotID *uuid.UUID
	// BalancesFetched and TransactionsFetched feed observability rather than
	// the API contract.
	BalancesFetched     int
	TransactionsFetched int
}

// Service executes the synchronization pipeline from §57.
type Service interface {
	// Run performs a full synchronization. It acquires the wallet sync lock
	// first and returns a domain-level error when another run holds it (§61).
	// The whole operation is idempotent: retrying must not duplicate
	// transactions, positions or snapshots (§60).
	Run(ctx context.Context, req Request) (Result, error)
	// ListDue returns wallets whose data is older than the configured
	// interval, which the scheduler enqueues (§62).
	ListDue(ctx context.Context, limit int) ([]uuid.UUID, error)
}

// Enqueuer hands work to the background queue. The HTTP layer uses it so that
// wallet creation never performs a synchronous blockchain sync (§57, §204).
type Enqueuer interface {
	EnqueueInitialSync(ctx context.Context, walletID uuid.UUID) error
	EnqueueSync(ctx context.Context, walletID uuid.UUID, trigger wallet.SyncTrigger) error
}

// StageReporter records real pipeline progress. A stage is reported only once
// the backend has actually reached it (§19).
type StageReporter interface {
	Report(ctx context.Context, walletID uuid.UUID, stage wallet.SyncStage) error
}
