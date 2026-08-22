// Package transactions defines the transaction application service. It serves
// canonical transaction facts; classification lives in a dedicated
// deterministic classifier (§170).
package transactions

import (
	"context"

	"github.com/google/uuid"

	"github.com/maxaicrypto/backend/internal/domain/asset"
	"github.com/maxaicrypto/backend/internal/domain/shared"
	"github.com/maxaicrypto/backend/internal/domain/transaction"
)

// View is a transaction with its assets resolved, which is what both the API
// and the AI transaction explainer consume (§100).
type View struct {
	Transaction transaction.Transaction
	AssetIn     *asset.Asset
	AssetOut    *asset.Asset
	FeeAsset    *asset.Asset
	// ValueInUSD, ValueOutUSD and FeeValueUSD are backend-computed. The AI
	// never computes transaction amounts or fees itself (§39, §179).
	ValueInUSD  shared.NullDecimal
	ValueOutUSD shared.NullDecimal
	FeeValueUSD shared.NullDecimal
	ExplorerURL *string
}

// Service exposes canonical transactions (§170).
type Service interface {
	// List returns a wallet's transactions using cursor pagination over the
	// stable ordering `timestamp DESC, id DESC` (§99, §109).
	List(ctx context.Context, userID, walletID uuid.UUID, filter transaction.Filter, page shared.Cursor, limit int) (shared.Page[View], error)
	// Get returns one transaction after verifying wallet ownership (§108).
	Get(ctx context.Context, userID, walletID, transactionID uuid.UUID) (View, error)
}
