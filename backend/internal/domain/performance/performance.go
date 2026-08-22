// Package performance models snapshot-based portfolio performance. The MVP
// deliberately does not implement an accounting engine: no realized or
// unrealized PnL, no cost basis, no tax lots (§52).
package performance

import (
	"time"

	"github.com/google/uuid"

	"github.com/maxaicrypto/backend/internal/domain/asset"
	"github.com/maxaicrypto/backend/internal/domain/shared"
)

// CalculationVersion identifies the performance algorithm, persisted with
// results used as AI evidence (§51, §56).
const CalculationVersion = 1

// Period is a supported comparison window (§53). The wire values match the
// frontend contract; see openapi/DECISIONS.md for the divergence from the
// §98 spelling of all_time.
type Period string

const (
	Period24h     Period = "24h"
	Period7d      Period = "7d"
	Period30d     Period = "30d"
	PeriodAllTime Period = "all"
)

// SupportedPeriods lists every period the API accepts.
var SupportedPeriods = []Period{Period24h, Period7d, Period30d, PeriodAllTime}

// ParsePeriod validates a period value from the query string.
func ParsePeriod(raw string) (Period, bool) {
	for _, period := range SupportedPeriods {
		if Period(raw) == period {
			return period, true
		}
	}
	return "", false
}

// Lookback returns the age of the opening snapshot for a fixed-window period.
// ALL_TIME has no fixed lookback: it anchors on the first valid snapshot (§53).
func (p Period) Lookback() (time.Duration, bool) {
	switch p {
	case Period24h:
		return 24 * time.Hour, true
	case Period7d:
		return 7 * 24 * time.Hour, true
	case Period30d:
		return 30 * 24 * time.Hour, true
	default:
		return 0, false
	}
}

// Status reports whether a performance result could be produced (§53).
type Status string

const (
	// StatusAvailable means both endpoints were valid.
	StatusAvailable Status = "AVAILABLE"
	// StatusPartial means a result exists but its inputs were incomplete.
	StatusPartial Status = "PARTIAL"
	// StatusUnavailable means no historical snapshot anchors the period.
	StatusUnavailable Status = "UNAVAILABLE"
)

// Endpoint is one side of a performance comparison.
type Endpoint struct {
	SnapshotID uuid.UUID
	CapturedAt time.Time
	ValueUSD   shared.Decimal
	Status     shared.ValuationStatus
}

// SeriesPoint is one point of the historical chart.
type SeriesPoint struct {
	CapturedAt    time.Time
	TotalValueUSD shared.NullDecimal
	Status        shared.ValuationStatus
}

// Driver attributes part of the change to one asset. Contribution is computed
// by the backend and only explained by the AI (§56).
type Driver struct {
	Asset           asset.Asset
	AllocationPct   shared.NullDecimal
	ContributionUSD shared.NullDecimal
	ContributionPct shared.NullDecimal
	ChangePct       shared.NullDecimal
}

// Performance is a computed snapshot-based performance result.
type Performance struct {
	WalletID    uuid.UUID
	Period      Period
	Status      Status
	DataQuality shared.DataQuality
	Currency    shared.Currency

	Opening *Endpoint
	Closing *Endpoint

	ChangeUSD shared.NullDecimal
	ChangePct shared.NullDecimal

	Series  []SeriesPoint
	Drivers []Driver

	// UnavailableReason carries the domain error code explaining a missing
	// result, so the frontend can present it without guessing.
	UnavailableReason *string

	// CalculationID identifies this specific computation, letting an AI claim
	// point at the evidence that produced it (§53, §73).
	CalculationID      *uuid.UUID
	CalculationVersion *int
}
