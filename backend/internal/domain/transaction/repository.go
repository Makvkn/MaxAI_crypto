package transaction

import (
	"context"

	"github.com/google/uuid"

	"github.com/maxaicrypto/backend/internal/domain/shared"
)

// Repository persists canonical transactions.
type Repository interface {
	// ListByWallet returns transactions newest first, using the stable
	// ordering `timestamp DESC, id DESC` that the cursor encodes (§109).
	ListByWallet(ctx context.Context, walletID uuid.UUID, filter Filter, page shared.Cursor, limit int) ([]Transaction, error)
	// GetByID returns one transaction. Wallet ownership is verified by the
	// application layer (§108).
	GetByID(ctx context.Context, id uuid.UUID) (Transaction, error)
	// UpsertBatch writes normalized transactions idempotently, keyed on
	// Identity, so retrying a sync cannot create duplicates (§48, §60).
	UpsertBatch(ctx context.Context, transactions []Transaction) (int, error)
}

// Classifier assigns a canonical type to a normalized transaction. It is
// deterministic and backend-owned; uncertainty resolves to TypeUnknown rather
// than a guess (§47).
type Classifier interface {
	Classify(ctx context.Context, tx Transaction) (Type, error)
}
