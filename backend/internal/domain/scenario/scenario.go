// Package scenario models deterministic what-if calculations. The LLM never
// computes the financial outcome; it only explains the structured result
// produced here (§83).
package scenario

import (
	"time"

	"github.com/google/uuid"

	"github.com/maxaicrypto/backend/internal/domain/asset"
	"github.com/maxaicrypto/backend/internal/domain/shared"
)

// CalculationVersion identifies the scenario algorithm. It is returned with
// every result so an explanation can be traced back to the exact computation
// that produced it (§51, §85).
const CalculationVersion = 1

// Type is a supported scenario. The MVP supports a single-asset price change
// (§83); the enum exists so new scenario kinds do not change the API shape.
type Type string

const TypeAssetPriceChange Type = "ASSET_PRICE_CHANGE"

// Request is a validated scenario input (§84).
type Request struct {
	WalletID uuid.UUID
	Type     Type
	AssetID  uuid.UUID
	// ChangePct is a percentage change such as -20, held as an exact decimal.
	ChangePct shared.Decimal
}

// Baseline is the current state the scenario starts from.
type Baseline struct {
	PortfolioValueUSD  shared.NullDecimal
	AssetValueUSD      shared.NullDecimal
	AssetAllocationPct shared.NullDecimal
}

// Projection is the deterministic outcome of applying the scenario.
type Projection struct {
	PortfolioValueUSD  shared.NullDecimal
	AssetValueUSD      shared.NullDecimal
	AssetImpactUSD     shared.NullDecimal
	PortfolioChangeUSD shared.NullDecimal
	PortfolioChangePct shared.NullDecimal
}

// Result is the structured scenario outcome handed to the AI for explanation.
type Result struct {
	ID       uuid.UUID
	UserID   uuid.UUID
	WalletID uuid.UUID
	Type     Type
	Currency shared.Currency
	Asset    asset.Asset
	// AssetID is retained so a persisted result can be reloaded before the
	// asset aggregate is attached.
	AssetID   uuid.UUID
	ChangePct shared.Decimal

	Baseline   Baseline
	Projection Projection

	// DataQuality propagates the quality of the portfolio the scenario was
	// computed from, so the AI cannot describe a partial result as exact
	// (§144).
	DataQuality shared.DataQuality

	CalculationID      uuid.UUID
	CalculationVersion int
	CreatedAt          time.Time
}
