package asset

import (
	"context"

	"github.com/google/uuid"
)

// Repository resolves and persists asset identities.
type Repository interface {
	GetByID(ctx context.Context, id uuid.UUID) (Asset, error)
	// GetManyByID loads several assets at once so that portfolio assembly does
	// not issue one query per position.
	GetManyByID(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]Asset, error)
	// FindByIdentity resolves an asset by chain and contract address.
	FindByIdentity(ctx context.Context, identity Identity) (Asset, bool, error)
	// Upsert resolves an asset discovered during synchronization, creating it
	// when unknown. It must be idempotent so that a retried sync cannot create
	// duplicate assets (§60).
	Upsert(ctx context.Context, a Asset) (Asset, error)
	// SetMarketDataMapping records or clears the price-source mapping.
	SetMarketDataMapping(ctx context.Context, id uuid.UUID, provider *MarketDataProvider, marketDataID *string) error
	// ListUnmapped returns assets that still lack a market-data mapping, so a
	// background job can attempt resolution without blocking a sync.
	ListUnmapped(ctx context.Context, limit int) ([]Asset, error)
}

// VisibilityClassifier applies the deterministic spam and dust rules (§43).
// It is a domain service, not an LLM call.
type VisibilityClassifier interface {
	Classify(ctx context.Context, input VisibilityInput) (Visibility, error)
}

// VisibilityInput carries the facts the deterministic rules operate on.
type VisibilityInput struct {
	Asset Asset
	// ValueUSD is nil when the asset has no known price, which on its own is
	// not evidence of spam.
	ValueUSD *string
	// HasMarketData mirrors Asset.HasMarketData and is passed explicitly so
	// rules stay easy to test in isolation.
	HasMarketData bool
}
