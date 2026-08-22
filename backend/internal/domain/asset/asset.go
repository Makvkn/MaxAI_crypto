// Package asset models blockchain assets and their market-data mapping. An
// asset is never identified by symbol alone (§31).
package asset

import (
	"time"

	"github.com/google/uuid"

	"github.com/maxaicrypto/backend/internal/domain/chain"
)

// Type classifies an asset. NFTs and complex DeFi positions are outside MVP
// valuation (§32, §44).
type Type string

const (
	// TypeNative is the chain's own asset, which has no contract address.
	TypeNative Type = "NATIVE"
	// TypeToken is a contract-issued token.
	TypeToken Type = "TOKEN"
	// TypeUnknown covers assets the backend could not classify confidently.
	TypeUnknown Type = "UNKNOWN"
)

// Visibility is the deterministic spam/dust classification from §43. The LLM is
// never the authoritative classifier.
type Visibility string

const (
	VisibilityVisible    Visibility = "VISIBLE"
	VisibilityHiddenDust Visibility = "HIDDEN_DUST"
	VisibilityHiddenSpam Visibility = "HIDDEN_SPAM"
	VisibilityUnknown    Visibility = "UNKNOWN"
)

// MarketDataProvider names the price source an asset is mapped to.
type MarketDataProvider string

const MarketDataCoinGecko MarketDataProvider = "coingecko"

// Asset is a canonical asset identity. Identity is chain plus contract
// address, with a dedicated representation for native assets (§31).
type Asset struct {
	ID      uuid.UUID
	ChainID chain.ID
	// ContractAddress is nil for native assets and a normalized contract
	// address for tokens.
	ContractAddress *string
	Symbol          string
	Name            string
	Decimals        int
	Type            Type
	IconURL         *string

	// MarketDataProvider and MarketDataID form the mapping to a price source.
	// When no reliable mapping exists both stay nil and the price is unknown,
	// never zero (§33, §40).
	MarketDataProvider *MarketDataProvider
	MarketDataID       *string

	CreatedAt time.Time
	UpdatedAt time.Time
}

// HasMarketData reports whether a reliable price lookup is possible for this
// asset.
func (a Asset) HasMarketData() bool {
	return a.MarketDataProvider != nil && a.MarketDataID != nil && *a.MarketDataID != ""
}

// IsNative reports whether this is the chain's own asset.
func (a Asset) IsNative() bool { return a.Type == TypeNative && a.ContractAddress == nil }

// Identity uniquely addresses an asset before it has been persisted, which is
// what the normalization layer produces from provider data (§30).
type Identity struct {
	ChainID chain.ID
	// ContractAddress is nil for the chain's native asset.
	ContractAddress *string
}
