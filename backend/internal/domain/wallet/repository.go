package wallet

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/maxaicrypto/backend/internal/domain/chain"
	"github.com/maxaicrypto/backend/internal/domain/shared"
)

// Repository persists wallets. Soft-deleted wallets are excluded from every
// listing method unless a method name says otherwise (§17).
type Repository interface {
	Create(ctx context.Context, w Wallet) (Wallet, error)
	// GetByID returns a wallet regardless of owner. Ownership is enforced by
	// the application layer, which knows the requesting user (§108).
	GetByID(ctx context.Context, id uuid.UUID) (Wallet, error)
	// ListByUser returns a user's active wallets, newest first.
	ListByUser(ctx context.Context, userID uuid.UUID, page shared.Cursor, limit int) ([]Wallet, error)
	// FindByAddress resolves an existing wallet for a user on a chain, which
	// prevents duplicate wallets for the same normalized address.
	FindByAddress(ctx context.Context, userID uuid.UUID, chainID chain.ID, address string) (Wallet, bool, error)
	// CountByUser counts active wallets for entitlement enforcement (§89).
	CountByUser(ctx context.Context, userID uuid.UUID) (int, error)
	// SoftDelete marks the wallet deleted without destroying historical data.
	SoftDelete(ctx context.Context, id uuid.UUID, at time.Time) error
	// UpdateStatus changes the lifecycle status.
	UpdateStatus(ctx context.Context, id uuid.UUID, status Status) error
	// ListDueForSync returns wallets whose last successful sync is older than
	// the configured interval, used by the scheduler (§62).
	ListDueForSync(ctx context.Context, olderThan time.Time, limit int) ([]Wallet, error)
}

// SyncStateRepository persists synchronization state and attempt history.
type SyncStateRepository interface {
	// Get returns the current sync state of a wallet.
	Get(ctx context.Context, walletID uuid.UUID) (SyncState, error)
	// GetMany returns sync state for several wallets, so listing endpoints do
	// not issue one query per wallet.
	GetMany(ctx context.Context, walletIDs []uuid.UUID) (map[uuid.UUID]SyncState, error)
	// Save writes the full state. Implementations must reject transitions the
	// state machine forbids (§18).
	Save(ctx context.Context, state SyncState) error
	// StartRun records the beginning of a synchronization attempt.
	StartRun(ctx context.Context, run SyncRun) (SyncRun, error)
	// FinishRun records the outcome of a synchronization attempt.
	FinishRun(ctx context.Context, run SyncRun) error
}
