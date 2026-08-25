// Package portfolio models portfolio valuation and its historical snapshots.
// Valuation is deterministic backend logic; the AI only explains it (§39, §54).
package portfolio

import (
	"time"

	"github.com/google/uuid"

	"github.com/maxaicrypto/backend/internal/domain/asset"
	"github.com/maxaicrypto/backend/internal/domain/price"
	"github.com/maxaicrypto/backend/internal/domain/shared"
)

// CalculationVersion identifies the valuation algorithm that produced a
// result. It is persisted with every snapshot so results stay reproducible
// after the logic changes (§51).
const CalculationVersion = 1

// Position is one valued holding inside a portfolio.
type Position struct {
	Asset   asset.Asset
	Balance shared.Decimal
	// BalanceRaw is the unscaled on-chain integer amount (§38).
	BalanceRaw string
	// Price is nil when no reliable market price exists.
	Price *price.Price
	// ValueUSD is unknown, not zero, when the price is unavailable (§40).
	ValueUSD      shared.NullDecimal
	AllocationPct shared.NullDecimal
	Change24hPct  shared.NullDecimal
	Change24hUSD  shared.NullDecimal
	Visibility    asset.Visibility
	// ValuationStatus is COMPLETE when the position could be valued and
	// UNAVAILABLE when its price is missing.
	ValuationStatus shared.ValuationStatus
	UpdatedAt       time.Time
}

// IsValued reports whether this position contributed to the portfolio total.
func (p Position) IsValued() bool { return p.ValueUSD.Valid }

// Exclusions explains what was deliberately left out of the total, so the UI
// can state it rather than silently under-reporting (§44).
type Exclusions struct {
	UnpricedPositions     int
	NFTsExcluded          bool
	DeFiPositionsExcluded bool
}

// NoticeCode identifies a machine-readable data-quality notice attached to a
// portfolio response.
type NoticeCode string

const (
	NoticeUnpricedAssetsExcluded NoticeCode = "UNPRICED_ASSETS_EXCLUDED"
	NoticeNFTsExcluded           NoticeCode = "NFTS_EXCLUDED_FROM_VALUATION"
	NoticeDeFiExcluded           NoticeCode = "DEFI_POSITIONS_EXCLUDED"
	NoticeDataStale              NoticeCode = "DATA_STALE"
	NoticeHistoryIncomplete      NoticeCode = "HISTORY_INCOMPLETE"
	NoticeSyncPartiallyFailed    NoticeCode = "SYNC_PARTIALLY_FAILED"
)

// NoticeSeverity ranks a notice for presentation.
type NoticeSeverity string

const (
	SeverityInfo    NoticeSeverity = "INFO"
	SeverityWarning NoticeSeverity = "WARNING"
)

// Notice is a structured data-quality statement. The backend emits codes and
// parameters; the frontend owns the wording (§141).
type Notice struct {
	Code     NoticeCode
	Severity NoticeSeverity
	Params   map[string]string
}

// Portfolio is the computed current state of a wallet.
type Portfolio struct {
	WalletID uuid.UUID
	Currency shared.Currency
	// TotalValueUSD is the sum over positions with a valid price only (§39).
	// A known-empty wallet (synced, no holdings) is $0, not unknown.
	TotalValueUSD      shared.NullDecimal
	ValuationStatus    shared.ValuationStatus
	DataQuality        shared.DataQuality
	DataFreshness      shared.DataFreshness
	Change24hUSD       shared.NullDecimal
	Change24hPct       shared.NullDecimal
	AsOf               time.Time
	LastSyncedAt       *time.Time
	CalculationVersion int
	Positions          []Position
	Exclusions         Exclusions
	Notices            []Notice
}
