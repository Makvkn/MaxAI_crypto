// Package providers wires the concrete adapters into the registry and resolver.
// This file is the single place a new provider is introduced (§26).
package providers

import (
	"github.com/maxaicrypto/backend/internal/app/config"
	"github.com/maxaicrypto/backend/internal/domain/chain"
	"github.com/maxaicrypto/backend/internal/domain/provider"
	"github.com/maxaicrypto/backend/internal/providers/coingecko"
	"github.com/maxaicrypto/backend/internal/providers/fixture"
	"github.com/maxaicrypto/backend/internal/providers/openai"
	"github.com/maxaicrypto/backend/internal/providers/registry"
	"github.com/maxaicrypto/backend/internal/providers/tatum"
	"github.com/maxaicrypto/backend/internal/providers/zerion"
)

// Providers bundles the registry and the resolver built over it.
type Providers struct {
	Registry *registry.Registry
	Resolver *registry.Resolver
}

// Build registers every adapter and returns the resolver business services use.
// Routing preference is declared here as data: Zerion is primary on EVM chains,
// Tatum serves the remaining MVP chains and backs up EVM (§25, §157).
func Build(cfg *config.Config) *Providers {
	reg := registry.New()
	reg.RegisterBlockchain(zerion.New(cfg.Provider))
	reg.RegisterBlockchain(tatum.New(cfg.Provider))
	if !cfg.Provider.HasBlockchainCredentials() {
		reg.RegisterBlockchain(fixture.New())
	}
	reg.RegisterMarket(coingecko.New(cfg.Provider))
	reg.RegisterLLM(openai.New(cfg.AI, cfg.Provider))

	priorities := defaultPriorities()
	if !cfg.Provider.HasBlockchainCredentials() {
		priorities = withFixtureFirst(priorities)
	}

	resolver := registry.NewResolver(reg, registry.ResolverConfig{
		Priorities: priorities,
		Market:     provider.CoinGecko,
		LLM:        provider.OpenAI,
	})

	return &Providers{Registry: reg, Resolver: resolver}
}

func defaultPriorities() []registry.Priority {
	return []registry.Priority{
		{ChainID: chain.Ethereum, Providers: []provider.Name{provider.Zerion, provider.Tatum}},
		{ChainID: chain.BNBChain, Providers: []provider.Name{provider.Zerion, provider.Tatum}},
		{ChainID: chain.Bitcoin, Providers: []provider.Name{provider.Tatum}},
		{ChainID: chain.Litecoin, Providers: []provider.Name{provider.Tatum}},
		{ChainID: chain.Dogecoin, Providers: []provider.Name{provider.Tatum}},
		{ChainID: chain.Solana, Providers: []provider.Name{provider.Tatum}},
		{ChainID: chain.Tron, Providers: []provider.Name{provider.Tatum}},
		{ChainID: chain.XRPLedger, Providers: []provider.Name{provider.Tatum}},
	}
}

func withFixtureFirst(priorities []registry.Priority) []registry.Priority {
	updated := make([]registry.Priority, len(priorities))
	for i, priority := range priorities {
		names := append([]provider.Name{provider.Fixture}, priority.Providers...)
		updated[i] = registry.Priority{ChainID: priority.ChainID, Providers: names}
	}
	return updated
}
