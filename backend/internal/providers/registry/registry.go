// Package registry implements the provider registry and resolver. Adding a
// provider means registering it here; no business service changes (§24, §26).
package registry

import (
	"sync"

	"github.com/maxaicrypto/backend/internal/domain/provider"
)

// Registry is the in-memory implementation of provider.Registry. It is
// populated once at startup and read concurrently afterwards.
type Registry struct {
	mu         sync.RWMutex
	blockchain map[provider.Name]provider.BlockchainDataProvider
	market     map[provider.Name]provider.MarketDataProvider
	llm        map[provider.Name]provider.LLMProvider
	// order preserves registration order so the resolver's routing table is
	// deterministic rather than dependent on map iteration.
	order []provider.Name
}

// New returns an empty registry.
func New() *Registry {
	return &Registry{
		blockchain: make(map[provider.Name]provider.BlockchainDataProvider),
		market:     make(map[provider.Name]provider.MarketDataProvider),
		llm:        make(map[provider.Name]provider.LLMProvider),
	}
}

// RegisterBlockchain adds a blockchain data provider.
func (r *Registry) RegisterBlockchain(p provider.BlockchainDataProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.blockchain[p.Name()]; !exists {
		r.order = append(r.order, p.Name())
	}
	r.blockchain[p.Name()] = p
}

// RegisterMarket adds a market data provider.
func (r *Registry) RegisterMarket(p provider.MarketDataProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.market[p.Name()] = p
}

// RegisterLLM adds an LLM provider.
func (r *Registry) RegisterLLM(p provider.LLMProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.llm[p.Name()] = p
}

// Blockchain implements provider.Registry.
func (r *Registry) Blockchain(name provider.Name) (provider.BlockchainDataProvider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.blockchain[name]
	return p, ok
}

// Market implements provider.Registry.
func (r *Registry) Market(name provider.Name) (provider.MarketDataProvider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.market[name]
	return p, ok
}

// LLM implements provider.Registry.
func (r *Registry) LLM(name provider.Name) (provider.LLMProvider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.llm[name]
	return p, ok
}

// BlockchainProviders implements provider.Registry, returning providers in
// registration order.
func (r *Registry) BlockchainProviders() []provider.BlockchainDataProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]provider.BlockchainDataProvider, 0, len(r.order))
	for _, name := range r.order {
		out = append(out, r.blockchain[name])
	}
	return out
}
