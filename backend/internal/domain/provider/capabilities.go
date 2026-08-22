// Package provider defines the ports through which the domain reaches external
// data sources. The interfaces live in the domain and the implementations in
// infrastructure, so business logic never depends on a specific provider
// (§6 Principle 1, §22).
package provider

import (
	"context"

	"github.com/maxaicrypto/backend/internal/domain/chain"
)

// Name identifies a registered provider implementation.
type Name string

const (
	Zerion    Name = "zerion"
	Tatum     Name = "tatum"
	CoinGecko Name = "coingecko"
	OpenAI    Name = "openai"
	Fixture   Name = "fixture"
)

// Capability is a discrete operation a provider can perform for a chain.
// Capabilities are data, never provider-specific conditionals inside domain
// services (§23).
type Capability string

const (
	CapabilityBalances       Capability = "balances"
	CapabilityTransactions   Capability = "transactions"
	CapabilityTokenMetadata  Capability = "token_metadata"
	CapabilityNativeAsset    Capability = "native_asset"
	CapabilityHistoricalData Capability = "historical_data"
	CapabilityPagination     Capability = "pagination"
)

// Capabilities describes what one provider supports.
type Capabilities struct {
	Provider Name
	// Chains maps each supported chain to the capabilities available on it.
	// A provider that covers a chain only partially is expressed here rather
	// than by a branch in a service.
	Chains map[chain.ID][]Capability
}

// Supports reports whether the provider offers capability on chainID.
func (c Capabilities) Supports(chainID chain.ID, capability Capability) bool {
	for _, supported := range c.Chains[chainID] {
		if supported == capability {
			return true
		}
	}
	return false
}

// Registry holds the registered provider implementations (§24). It is
// infrastructure; the domain only consumes the interface.
type Registry interface {
	// Blockchain returns a registered blockchain provider by name.
	Blockchain(name Name) (BlockchainDataProvider, bool)
	// Market returns a registered market-data provider by name.
	Market(name Name) (MarketDataProvider, bool)
	// LLM returns a registered LLM provider by name.
	LLM(name Name) (LLMProvider, bool)
	// BlockchainProviders lists every registered blockchain provider, which
	// the resolver uses to build its routing table.
	BlockchainProviders() []BlockchainDataProvider
}

// Resolver selects the provider that should serve a given chain and capability
// (§25). Routing lives here so that no business service ever branches on a
// provider name.
type Resolver interface {
	// ResolveBlockchain returns the primary provider for a chain and
	// capability, or an error when no registered provider supports it.
	ResolveBlockchain(ctx context.Context, chainID chain.ID, capability Capability) (BlockchainDataProvider, error)
	// ResolveBlockchainChain returns the primary provider followed by any
	// configured fallbacks, in priority order. Fallback is capability-based
	// and is not mandatory for every operation in the MVP (§157).
	ResolveBlockchainChain(ctx context.Context, chainID chain.ID, capability Capability) ([]BlockchainDataProvider, error)
	// ResolveMarket returns the market-data provider.
	ResolveMarket(ctx context.Context) (MarketDataProvider, error)
	// ResolveLLM returns the configured LLM provider.
	ResolveLLM(ctx context.Context) (LLMProvider, error)
}
