// Package wallet models analysed wallets and their synchronization state. The
// UI analyses one wallet at a time, but the model is 1 user to N wallets from
// the start (§16).
package wallet

import (
	"time"

	"github.com/google/uuid"

	"github.com/maxaicrypto/backend/internal/domain/chain"
)

// Status is the wallet lifecycle state (§17). It is deliberately separate from
// synchronization state.
type Status string

const (
	StatusActive  Status = "ACTIVE"
	StatusSyncing Status = "SYNCING"
	StatusError   Status = "ERROR"
	StatusPaused  Status = "PAUSED"
	// StatusDeleted is a soft delete. Deleted wallets must not appear in
	// normal active-wallet queries.
	StatusDeleted Status = "DELETED"
)

// Wallet is a public blockchain address the user wants analysed. The backend
// never holds keys, seed phrases or any signing material (§2).
type Wallet struct {
	ID      uuid.UUID
	UserID  uuid.UUID
	ChainID chain.ID
	// Address is the canonical, chain-normalized representation (§188).
	Address   string
	Label     *string
	Status    Status
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// IsDeleted reports whether the wallet has been soft deleted.
func (w Wallet) IsDeleted() bool { return w.DeletedAt != nil || w.Status == StatusDeleted }
