package performance

import (
	"testing"
	"time"
)

func TestParsePeriodAcceptsOnlySupportedValues(t *testing.T) {
	for _, raw := range []string{"24h", "7d", "30d", "all"} {
		if _, ok := ParsePeriod(raw); !ok {
			t.Errorf("ParsePeriod(%q) was rejected, want accepted", raw)
		}
	}
	for _, raw := range []string{"", "1h", "all_time", "24H", "365d"} {
		if _, ok := ParsePeriod(raw); ok {
			t.Errorf("ParsePeriod(%q) was accepted, want rejected", raw)
		}
	}
}

func TestAllTimeHasNoFixedLookback(t *testing.T) {
	if _, ok := PeriodAllTime.Lookback(); ok {
		t.Error("all-time reported a fixed lookback; it must anchor on the first valid snapshot")
	}

	fixed := map[Period]time.Duration{
		Period24h: 24 * time.Hour,
		Period7d:  7 * 24 * time.Hour,
		Period30d: 30 * 24 * time.Hour,
	}
	for period, want := range fixed {
		got, ok := period.Lookback()
		if !ok {
			t.Errorf("%s has no lookback", period)
			continue
		}
		if got != want {
			t.Errorf("%s lookback = %v, want %v", period, got, want)
		}
	}
}
