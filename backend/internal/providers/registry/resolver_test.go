package registry

import (
	"context"
	"testing"

	"github.com/maxaicrypto/backend/internal/domain/apperr"
	"github.com/maxaicrypto/backend/internal/domain/chain"
	"github.com/maxaicrypto/backend/internal/domain/provider"
)

type stubBlockchain struct {
	name   provider.Name
	chains map[chain.ID][]provider.Capability
}

func (s stubBlockchain) Name() provider.Name { return s.name }

func (s stubBlockchain) Capabilities(context.Context) provider.Capabilities {
	return provider.Capabilities{Provider: s.name, Chains: s.chains}
}

func (s stubBlockchain) GetBalances(context.Context, provider.BalanceRequest) ([]provider.NormalizedBalance, error) {
	return nil, nil
}

func (s stubBlockchain) GetTransactions(context.Context, provider.TransactionRequest) (provider.TransactionPage, error) {
	return provider.TransactionPage{}, nil
}

func newTestRegistry() *Registry {
	reg := New()
	reg.RegisterBlockchain(stubBlockchain{
		name: provider.Zerion,
		chains: map[chain.ID][]provider.Capability{
			chain.Ethereum: {provider.CapabilityBalances, provider.CapabilityTransactions},
		},
	})
	reg.RegisterBlockchain(stubBlockchain{
		name: provider.Tatum,
		chains: map[chain.ID][]provider.Capability{
			chain.Ethereum: {provider.CapabilityBalances},
			chain.Bitcoin:  {provider.CapabilityBalances, provider.CapabilityTransactions},
		},
	})
	return reg
}

func TestResolveBlockchainHonoursPriority(t *testing.T) {
	resolver := NewResolver(newTestRegistry(), ResolverConfig{
		Priorities: []Priority{
			{ChainID: chain.Ethereum, Providers: []provider.Name{provider.Tatum, provider.Zerion}},
		},
	})

	p, err := resolver.ResolveBlockchain(context.Background(), chain.Ethereum, provider.CapabilityBalances)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if p.Name() != provider.Tatum {
		t.Fatalf("want tatum as primary, got %s", p.Name())
	}
}

func TestResolveBlockchainSkipsProvidersWithoutCapability(t *testing.T) {
	resolver := NewResolver(newTestRegistry(), ResolverConfig{
		Priorities: []Priority{
			{ChainID: chain.Ethereum, Providers: []provider.Name{provider.Tatum, provider.Zerion}},
		},
	})

	chainProviders, err := resolver.ResolveBlockchainChain(context.Background(), chain.Ethereum, provider.CapabilityTransactions)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if len(chainProviders) != 1 || chainProviders[0].Name() != provider.Zerion {
		t.Fatalf("want only zerion for transactions, got %v", chainProviders)
	}
}

func TestResolveBlockchainUnsupportedChain(t *testing.T) {
	resolver := NewResolver(newTestRegistry(), ResolverConfig{})

	_, err := resolver.ResolveBlockchain(context.Background(), chain.Solana, provider.CapabilityBalances)
	if !apperr.Is(err, apperr.CodeUnsupportedChain) {
		t.Fatalf("want UNSUPPORTED_CHAIN, got %v", err)
	}
}

func TestResolveMarketMissingProvider(t *testing.T) {
	resolver := NewResolver(New(), ResolverConfig{Market: provider.CoinGecko})

	if _, err := resolver.ResolveMarket(context.Background()); !apperr.Is(err, apperr.CodePriceDataUnavailable) {
		t.Fatalf("want PRICE_DATA_UNAVAILABLE, got %v", err)
	}
}

func TestResolveLLMMissingProvider(t *testing.T) {
	resolver := NewResolver(New(), ResolverConfig{LLM: provider.OpenAI})

	if _, err := resolver.ResolveLLM(context.Background()); !apperr.Is(err, apperr.CodeAIUnavailable) {
		t.Fatalf("want AI_UNAVAILABLE, got %v", err)
	}
}
