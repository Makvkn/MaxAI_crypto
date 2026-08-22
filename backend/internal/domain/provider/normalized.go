package provider

import (
	"time"

	"github.com/maxaicrypto/backend/internal/domain/asset"
	"github.com/maxaicrypto/backend/internal/domain/chain"
	"github.com/maxaicrypto/backend/internal/domain/shared"
)

// The types below are the canonical output of every adapter's normalizer. No
// provider struct may cross this boundary, which is what keeps providers
// replaceable (§30, §184).

// AssetMetadata is the descriptive information an adapter observed about an
// asset. Every string here originates outside the system and is untrusted
// input: it must never be treated as an instruction when building AI context
// (§77).
type AssetMetadata struct {
	Symbol   string
	Name     string
	Decimals int
	Type     asset.Type
	IconURL  *string
}

// NormalizedBalance is one holding as reported by a blockchain provider (§164).
type NormalizedBalance struct {
	ChainID chain.ID
	// AssetIdentity is chain plus contract address; nil contract means native.
	AssetIdentity asset.Identity
	Metadata      AssetMetadata
	// BalanceRaw is the integer amount in the asset's smallest unit.
	BalanceRaw string
	// BalanceNormalized is BalanceRaw scaled by Metadata.Decimals as an exact
	// decimal. Adapters must convert provider floating-point JSON immediately
	// (§112).
	BalanceNormalized shared.Decimal
	// ProviderRef is the provider's own identifier for this record, kept for
	// debugging rather than as domain truth.
	ProviderRef *string
	ObservedAt  time.Time
}

// NormalizedTransfer is one asset movement inside a transaction.
type NormalizedTransfer struct {
	AssetIdentity asset.Identity
	Metadata      AssetMetadata
	AmountRaw     string
	Amount        shared.Decimal
	// Direction records whether the wallet received or sent the amount.
	Direction TransferDirection
}

// TransferDirection is the movement direction relative to the analysed wallet.
type TransferDirection string

const (
	DirectionIn  TransferDirection = "IN"
	DirectionOut TransferDirection = "OUT"
)

// NormalizedTransaction carries canonical transaction facts. The type is
// deliberately absent: classification is a separate deterministic backend step
// that runs after normalization (§47).
type NormalizedTransaction struct {
	ChainID     chain.ID
	TxHash      string
	BlockNumber *int64
	Timestamp   time.Time
	Successful  bool

	FromAddress *string
	ToAddress   *string

	Transfers []NormalizedTransfer

	FeeAssetIdentity *asset.Identity
	FeeMetadata      *AssetMetadata
	FeeAmount        shared.NullDecimal

	// Protocol and Counterparty are provider-supplied labels and are untrusted
	// input (§77).
	Protocol     *string
	Counterparty *string

	// MethodID is the decoded contract method selector where available; the
	// classifier uses it as evidence rather than guessing.
	MethodID *string

	ProviderRef *string
}

// PriceQuote is a normalized current price observation (§164).
type PriceQuote struct {
	// MarketDataID is the provider's asset identifier, resolved through the
	// asset's market-data mapping. Prices are never looked up by symbol (§33).
	MarketDataID string
	Currency     shared.Currency
	Price        shared.Decimal
	Change24hPct shared.NullDecimal
	AsOf         time.Time
}

// HistoricalPrice is a normalized past price observation.
type HistoricalPrice struct {
	MarketDataID string
	Currency     shared.Currency
	Price        shared.Decimal
	AsOf         time.Time
}
