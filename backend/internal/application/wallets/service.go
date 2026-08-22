// Package wallets defines the wallet application service. It owns wallet
// lifecycle and never fetches provider data itself (§166).
package wallets

import (
	"context"

	"github.com/google/uuid"

	"github.com/maxaicrypto/backend/internal/domain/chain"
	"github.com/maxaicrypto/backend/internal/domain/shared"
	"github.com/maxaicrypto/backend/internal/domain/wallet"
)

// CreateRequest is a validated wallet creation request (§93). Private keys and
// signing material are never accepted.
type CreateRequest struct {
	UserID  uuid.UUID
	ChainID chain.ID
	Address string
	Label   *string
}

// View is a wallet together with its synchronization state, which is what the
// frontend needs to drive the sync UX (§95).
type View struct {
	Wallet wallet.Wallet
	Sync   wallet.SyncState
}

// Service owns wallet creation, retrieval, soft deletion and sync triggering
// (§166).
type Service interface {
	// Create validates the chain and address, normalizes the address, enforces
	// the plan's wallet limit, persists the wallet and enqueues the initial
	// sync. It returns immediately without waiting for the sync (§57, §94).
	Create(ctx context.Context, req CreateRequest) (View, error)
	// Get returns one wallet after verifying the requesting user owns it
	// (§108).
	Get(ctx context.Context, userID, walletID uuid.UUID) (View, error)
	// List returns the user's active wallets.
	List(ctx context.Context, userID uuid.UUID, page shared.Cursor, limit int) (shared.Page[View], error)
	// Delete soft-deletes the wallet, preserving historical data (§17).
	Delete(ctx context.Context, userID, walletID uuid.UUID) error
	// RequestSync queues a manual synchronization. When one is already running
	// it returns a domain-level state rather than duplicating work (§158).
	RequestSync(ctx context.Context, userID, walletID uuid.UUID) (View, error)
}
