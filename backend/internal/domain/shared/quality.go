package shared

import "time"

// DataQuality describes the overall trustworthiness of a response. The four
// values are distinct concepts and must not be conflated: STALE describes
// freshness, PARTIAL describes completeness, UNAVAILABLE describes missing
// required data (§42).
type DataQuality string

const (
	DataQualityComplete    DataQuality = "COMPLETE"
	DataQualityPartial     DataQuality = "PARTIAL"
	DataQualityStale       DataQuality = "STALE"
	DataQualityUnavailable DataQuality = "UNAVAILABLE"
)

// DataFreshness is the age classification of the underlying data (§37).
type DataFreshness string

const (
	FreshnessFresh     DataFreshness = "FRESH"
	FreshnessRecent    DataFreshness = "RECENT"
	FreshnessStale     DataFreshness = "STALE"
	FreshnessVeryStale DataFreshness = "VERY_STALE"
)

// ValuationStatus describes whether a monetary total could be produced from the
// available inputs (§41).
type ValuationStatus string

const (
	ValuationComplete    ValuationStatus = "COMPLETE"
	ValuationPartial     ValuationStatus = "PARTIAL"
	ValuationUnavailable ValuationStatus = "UNAVAILABLE"
)

// Currency is the settlement currency of a monetary value. The MVP values
// everything in USD.
type Currency string

const CurrencyUSD Currency = "USD"

// FreshnessThresholds are the configured age boundaries between freshness
// buckets. They are configuration values, never constants (§37).
type FreshnessThresholds struct {
	FreshMax  time.Duration
	RecentMax time.Duration
	StaleMax  time.Duration
}

// Classify maps an age onto a freshness bucket.
func (t FreshnessThresholds) Classify(age time.Duration) DataFreshness {
	switch {
	case age < t.FreshMax:
		return FreshnessFresh
	case age < t.RecentMax:
		return FreshnessRecent
	case age < t.StaleMax:
		return FreshnessStale
	default:
		return FreshnessVeryStale
	}
}

// ClassifyAt maps the age of observedAt relative to now onto a freshness
// bucket. A zero timestamp means the data was never observed.
func (t FreshnessThresholds) ClassifyAt(observedAt, now time.Time) DataFreshness {
	if observedAt.IsZero() {
		return FreshnessVeryStale
	}
	return t.Classify(now.Sub(observedAt))
}
