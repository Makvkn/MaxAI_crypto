package registry

import (
	"context"
	"sort"

	"github.com/maxaicrypto/backend/internal/domain/apperr"
	"github.com/maxaicrypto/backend/internal/domain/chain"
	"github.com/maxaicrypto/backend/internal/domain/provider"
)

// Priority declares provider preference for one chain. Routing is
// configuration data, not a conditional inside a business service (§25).
type Priority struct {
	ChainID chain.ID
	// Providers is ordered: the first entry that declares the requested
	// capability is primary, the rest are fallbacks (§157).
	Providers []provider.Name
}

// ResolverConfig configures chain routing.
type ResolverConfig struct {
	// Priorities lists explicit per-chain ordering. Chains absent from it fall
	// back to registration order.
	Priorities []Priority
	// Market and LLM name the single providers used for those roles.
	Market provider.Name
	LLM    provider.Name
}

// Resolver implements provider.Resolver on top of a Registry.
type Resolver struct {
	registry   *Registry
	priorities map[chain.ID][]provider.Name
	market     provider.Name
	llm        provider.Name
}

// NewResolver builds a resolver over reg.
func NewResolver(reg *Registry, cfg ResolverConfig) *Resolver {
	priorities := make(map[chain.ID][]provider.Name, len(cfg.Priorities))
	for _, p := range cfg.Priorities {
		priorities[p.ChainID] = p.Providers
	}
	return &Resolver{
		registry:   reg,
		priorities: priorities,
		market:     cfg.Market,
		llm:        cfg.LLM,
	}
}

// ResolveBlockchain implements provider.Resolver.
func (r *Resolver) ResolveBlockchain(ctx context.Context, chainID chain.ID, capability provider.Capability) (provider.BlockchainDataProvider, error) {
	candidates, err := r.ResolveBlockchainChain(ctx, chainID, capability)
	if err != nil {
		return nil, err
	}
	return candidates[0], nil
}

// ResolveBlockchainChain implements provider.Resolver. It returns the primary
// provider followed by fallbacks, all of which declare the capability for the
// chain.
func (r *Resolver) ResolveBlockchainChain(ctx context.Context, chainID chain.ID, capability provider.Capability) ([]provider.BlockchainDataProvider, error) {
	var matches []provider.BlockchainDataProvider
	for _, p := range r.candidates(chainID) {
		if p.Capabilities(ctx).Supports(chainID, capability) {
			matches = append(matches, p)
		}
	}
	if len(matches) == 0 {
		return nil, apperr.New(apperr.CodeUnsupportedChain).
			WithDetail("chain", string(chainID)).
			WithDetail("capability", string(capability))
	}
	return matches, nil
}

// candidates returns the providers to try for a chain, in priority order.
func (r *Resolver) candidates(chainID chain.ID) []provider.BlockchainDataProvider {
	names, ok := r.priorities[chainID]
	if !ok {
		return r.registry.BlockchainProviders()
	}
	ranked := make(map[provider.Name]int, len(names))
	for i, name := range names {
		ranked[name] = i
	}
	all := r.registry.BlockchainProviders()
	// Providers outside the priority list stay usable but sort last, so a
	// misconfigured priority entry degrades instead of dropping coverage.
	sort.SliceStable(all, func(i, j int) bool {
		return rank(ranked, all[i].Name()) < rank(ranked, all[j].Name())
	})
	return all
}

func rank(ranked map[provider.Name]int, name provider.Name) int {
	if i, ok := ranked[name]; ok {
		return i
	}
	return len(ranked)
}

// ResolveMarket implements provider.Resolver.
func (r *Resolver) ResolveMarket(_ context.Context) (provider.MarketDataProvider, error) {
	p, ok := r.registry.Market(r.market)
	if !ok {
		return nil, apperr.New(apperr.CodePriceDataUnavailable).
			WithDetail("provider", string(r.market))
	}
	return p, nil
}

// ResolveLLM implements provider.Resolver.
func (r *Resolver) ResolveLLM(_ context.Context) (provider.LLMProvider, error) {
	p, ok := r.registry.LLM(r.llm)
	if !ok {
		return nil, apperr.New(apperr.CodeAIUnavailable).
			WithDetail("provider", string(r.llm))
	}
	return p, nil
}
