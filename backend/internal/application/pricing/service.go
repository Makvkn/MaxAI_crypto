// Package pricing defines the price service. Business services call it instead
// of a market-data provider, so no domain code knows that prices come from
// CoinGecko (§35).
package pricing

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/maxaicrypto/backend/internal/domain/asset"
	"github.com/maxaicrypto/backend/internal/domain/price"
)

// Service owns caching, freshness, normalization, source tracking, market-data
// mapping and future fallback behaviour (§35).
type Service interface {
	// GetCurrent returns the latest price for each asset. Assets without a
	// reliable price are absent from the result, and callers must treat that
	// as unknown rather than zero (§40).
	GetCurrent(ctx context.Context, assetIDs []uuid.UUID) (map[uuid.UUID]price.Price, error)
	// GetAt returns the price of an asset closest to a past instant, used to
	// reconstruct historical context.
	GetAt(ctx context.Context, assetID uuid.UUID, at time.Time) (price.Price, bool, error)
	// Refresh fetches and stores fresh quotes for the given assets, which the
	// price refresh job invokes.
	Refresh(ctx context.Context, assetIDs []uuid.UUID) (int, error)
	// ResolveMapping attempts to map an asset onto a market-data identifier.
	// When no reliable mapping exists the asset keeps a null mapping and its
	// price stays unknown (§33).
	ResolveMapping(ctx context.Context, a asset.Asset) (bool, error)
}
