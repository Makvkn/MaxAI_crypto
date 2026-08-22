// Package position models the current balance of one asset in one wallet.
// Valuation is derived from balance and price at read time; the stored balance
// is the only fact (§38).
package position

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/maxaicrypto/backend/internal/domain/shared"
)

// WalletPosition is the current holding of an asset in a wallet.
type WalletPosition struct {
	WalletID uuid.UUID
	AssetID  uuid.UUID
	// BalanceRaw is the on-chain integer amount in the asset's smallest unit,
	// kept as a string so no precision is lost for large-decimal tokens.
	BalanceRaw string
	// BalanceNormalized is BalanceRaw scaled by the asset's decimals, stored
	// as an exact decimal rather than a float (§38, §111).
	BalanceNormalized shared.Decimal
	UpdatedAt         time.Time
}

// Repository persists wallet positions.
type Repository interface {
	// ListByWallet returns every current position of a wallet.
	ListByWallet(ctx context.Context, walletID uuid.UUID) ([]WalletPosition, error)
	// GetByAsset returns one position, used by scenario calculations.
	GetByAsset(ctx context.Context, walletID, assetID uuid.UUID) (WalletPosition, bool, error)
	// ReplaceForWallet atomically replaces the wallet's positions with the
	// result of a successful synchronization. Replacing rather than merging is
	// what makes a retried sync idempotent: a position that disappeared
	// on-chain must disappear here too (§60).
	ReplaceForWallet(ctx context.Context, walletID uuid.UUID, positions []WalletPosition, observedAt time.Time) error
}
