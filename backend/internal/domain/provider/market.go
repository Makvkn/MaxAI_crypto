package provider

import (
	"context"
	"time"

	"github.com/maxaicrypto/backend/internal/domain/shared"
)

// PriceRequest asks for current prices. Assets are addressed by their
// market-data identifiers, never by symbol (§33).
type PriceRequest struct {
	MarketDataIDs []string
	Currency      shared.Currency
}

// HistoricalPriceRequest asks for past prices of a single asset.
type HistoricalPriceRequest struct {
	MarketDataID string
	Currency     shared.Currency
	From         time.Time
	To           time.Time
}

// MarketDataProvider is the port every price source implements (§34).
// Business services reach it through PriceService, never directly (§35).
type MarketDataProvider interface {
	// Name identifies the implementation for logging and metrics.
	Name() Name
	// GetPrices returns current quotes. Identifiers the provider does not know
	// are simply absent from the result: an unknown price is never zero (§40).
	GetPrices(ctx context.Context, req PriceRequest) ([]PriceQuote, error)
	// GetHistoricalPrices returns past quotes for one asset.
	GetHistoricalPrices(ctx context.Context, req HistoricalPriceRequest) ([]HistoricalPrice, error)
	// ResolveMarketDataID attempts to map a blockchain asset onto a provider
	// asset identifier. It returns false when no reliable mapping exists, in
	// which case the asset's price stays unknown (§33).
	ResolveMarketDataID(ctx context.Context, req MappingRequest) (string, bool, error)
}

// MappingRequest carries the facts used to resolve an asset's market-data
// identifier. Contract address plus chain is the reliable key; symbol alone is
// not (§33).
type MappingRequest struct {
	ChainID         string
	ContractAddress *string
	Symbol          string
	IsNative        bool
}
