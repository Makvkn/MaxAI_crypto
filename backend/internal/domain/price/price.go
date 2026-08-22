// Package price models market prices. A price is never an eternal value: it
// always carries a timestamp, a source and a freshness classification (§36).
package price

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/maxaicrypto/backend/internal/domain/shared"
)

// Status reports whether a usable price exists. An unavailable price is not
// zero (§40).
type Status string

const (
	StatusAvailable   Status = "AVAILABLE"
	StatusUnavailable Status = "UNAVAILABLE"
)

// Source names where a price came from, so a valuation stays auditable (§162).
type Source string

const SourceCoinGecko Source = "coingecko"

// Price is a market price observation for one asset.
type Price struct {
	AssetID  uuid.UUID
	Currency shared.Currency
	// ValueUSD is unknown rather than zero when Status is StatusUnavailable.
	ValueUSD  shared.NullDecimal
	Status    Status
	Source    Source
	AsOf      time.Time
	Change24h shared.NullDecimal
	CreatedAt time.Time
}

// IsUsable reports whether this price may participate in valuation (§39).
func (p Price) IsUsable() bool { return p.Status == StatusAvailable && p.ValueUSD.Valid }

// Freshness classifies the price age using configured thresholds (§37).
func (p Price) Freshness(thresholds shared.FreshnessThresholds, now time.Time) shared.DataFreshness {
	return thresholds.ClassifyAt(p.AsOf, now)
}

// HistoricalPoint is a price observation at a past moment, used to reconstruct
// historical context.
type HistoricalPoint struct {
	AssetID  uuid.UUID
	ValueUSD shared.Decimal
	AsOf     time.Time
	Source   Source
}

// Repository persists price observations. PostgreSQL holds the durable record;
// the Redis price cache in front of it is an optimization only (§118, §120).
type Repository interface {
	// UpsertLatest records the most recent observation for an asset.
	UpsertLatest(ctx context.Context, p Price) error
	// GetLatest returns the most recent stored price for each requested asset.
	// Assets without a price are simply absent from the result.
	GetLatest(ctx context.Context, assetIDs []uuid.UUID) (map[uuid.UUID]Price, error)
	// GetClosest returns the observation nearest to a past instant, used when
	// reconstructing historical context.
	GetClosest(ctx context.Context, assetID uuid.UUID, at time.Time) (Price, bool, error)
}
