// Package fixture provides deterministic blockchain data for development and
// integration tests when real provider credentials are not configured.
package fixture

import (
	"context"
	"time"

	"github.com/maxaicrypto/backend/internal/domain/asset"
	"github.com/maxaicrypto/backend/internal/domain/chain"
	"github.com/maxaicrypto/backend/internal/domain/provider"
	"github.com/maxaicrypto/backend/internal/domain/shared"
)

// Provider implements provider.BlockchainDataProvider with static sample data.
type Provider struct{}

// New builds the fixture adapter.
func New() *Provider { return &Provider{} }

// Name implements provider.BlockchainDataProvider.
func (p *Provider) Name() provider.Name { return provider.Fixture }

// Capabilities implements provider.BlockchainDataProvider.
func (p *Provider) Capabilities(context.Context) provider.Capabilities {
	all := []provider.Capability{
		provider.CapabilityBalances,
		provider.CapabilityTransactions,
		provider.CapabilityNativeAsset,
	}
	chains := make(map[chain.ID][]provider.Capability, len(chain.Supported))
	for _, id := range chain.Supported {
		chains[id] = all
	}
	return provider.Capabilities{Provider: provider.Fixture, Chains: chains}
}

// GetBalances implements provider.BlockchainDataProvider.
func (p *Provider) GetBalances(_ context.Context, req provider.BalanceRequest) ([]provider.NormalizedBalance, error) {
	now := time.Now().UTC()
	balance, _ := shared.ParseDecimal("1.5")
	return []provider.NormalizedBalance{{
		ChainID:       req.ChainID,
		AssetIdentity: asset.Identity{ChainID: req.ChainID},
		Metadata: provider.AssetMetadata{
			Symbol:   nativeSymbol(req.ChainID),
			Name:     nativeName(req.ChainID),
			Decimals: nativeDecimals(req.ChainID),
			Type:     asset.TypeNative,
		},
		BalanceRaw:        "1500000000000000000",
		BalanceNormalized: balance,
		ObservedAt:        now,
	}}, nil
}

// GetTransactions implements provider.BlockchainDataProvider.
func (p *Provider) GetTransactions(_ context.Context, req provider.TransactionRequest) (provider.TransactionPage, error) {
	now := time.Now().UTC().Add(-24 * time.Hour)
	amount, _ := shared.ParseDecimal("0.1")
	wallet := req.Address
	other := "0x0000000000000000000000000000000000000001"
	if req.ChainID != chain.Ethereum && req.ChainID != chain.BNBChain {
		other = "fixture-counterparty"
	}
	return provider.TransactionPage{
		Transactions: []provider.NormalizedTransaction{{
			ChainID:     req.ChainID,
			TxHash:      "0xfixture0000000000000000000000000000000000000000000000000000000001",
			BlockNumber: ptrInt64(1),
			Timestamp:   now,
			Successful:  true,
			FromAddress: &other,
			ToAddress:   &wallet,
			Transfers: []provider.NormalizedTransfer{{
				AssetIdentity: asset.Identity{ChainID: req.ChainID},
				Metadata: provider.AssetMetadata{
					Symbol:   nativeSymbol(req.ChainID),
					Name:     nativeName(req.ChainID),
					Decimals: nativeDecimals(req.ChainID),
					Type:     asset.TypeNative,
				},
				AmountRaw: "100000000000000000",
				Amount:    amount,
				Direction: provider.DirectionIn,
			}},
		}},
	}, nil
}

func nativeSymbol(id chain.ID) string {
	switch id {
	case chain.Bitcoin:
		return "BTC"
	case chain.BNBChain:
		return "BNB"
	case chain.Solana:
		return "SOL"
	case chain.Litecoin:
		return "LTC"
	case chain.XRPLedger:
		return "XRP"
	case chain.Tron:
		return "TRX"
	case chain.Dogecoin:
		return "DOGE"
	default:
		return "ETH"
	}
}

func nativeName(id chain.ID) string {
	switch id {
	case chain.Bitcoin:
		return "Bitcoin"
	case chain.BNBChain:
		return "BNB"
	case chain.Solana:
		return "Solana"
	case chain.Litecoin:
		return "Litecoin"
	case chain.XRPLedger:
		return "XRP"
	case chain.Tron:
		return "TRON"
	case chain.Dogecoin:
		return "Dogecoin"
	default:
		return "Ether"
	}
}

func nativeDecimals(id chain.ID) int {
	switch id {
	case chain.Bitcoin, chain.Litecoin, chain.Dogecoin:
		return 8
	case chain.Solana:
		return 9
	case chain.XRPLedger, chain.Tron:
		return 6
	default:
		return 18
	}
}

func ptrInt64(v int64) *int64 { return &v }
