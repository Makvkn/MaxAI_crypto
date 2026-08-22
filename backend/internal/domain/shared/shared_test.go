package shared

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDecimalSerializesAsStringWithoutLosingPrecision(t *testing.T) {
	const exact = "24850.123456789012345678"

	d, err := ParseDecimal(exact)
	if err != nil {
		t.Fatalf("ParseDecimal returned an error: %v", err)
	}

	encoded, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("Marshal returned an error: %v", err)
	}
	if string(encoded) != `"`+exact+`"` {
		t.Fatalf("encoded = %s, want %q", encoded, exact)
	}

	var round Decimal
	if err := json.Unmarshal(encoded, &round); err != nil {
		t.Fatalf("Unmarshal returned an error: %v", err)
	}
	if round.String() != exact {
		t.Errorf("round trip = %s, want %s", round.String(), exact)
	}
}

func TestDecimalRejectsJSONNumbers(t *testing.T) {
	var d Decimal
	if err := json.Unmarshal([]byte("24850.12"), &d); err == nil {
		t.Fatal("Unmarshal accepted a JSON number, want an error")
	}
}

func TestUnknownIsDistinctFromZero(t *testing.T) {
	unknown, err := json.Marshal(Unknown())
	if err != nil {
		t.Fatalf("Marshal returned an error: %v", err)
	}
	if string(unknown) != "null" {
		t.Errorf("unknown encoded as %s, want null", unknown)
	}

	zero, err := json.Marshal(Known(Zero))
	if err != nil {
		t.Fatalf("Marshal returned an error: %v", err)
	}
	if string(zero) != `"0"` {
		t.Errorf("zero encoded as %s, want \"0\"", zero)
	}
}

func TestCursorRoundTrip(t *testing.T) {
	original := NewCursor(time.Date(2026, 8, 22, 10, 30, 0, 0, time.UTC), "tx_42")

	decoded, err := ParseCursor(original.Encode())
	if err != nil {
		t.Fatalf("ParseCursor returned an error: %v", err)
	}
	if !decoded.SortKey.Equal(original.SortKey) || decoded.TieBreaker != original.TieBreaker {
		t.Errorf("decoded = %+v, want %+v", decoded, original)
	}
}

func TestParseCursorRejectsGarbage(t *testing.T) {
	for _, raw := range []string{"not-base64!!", "e30", "MTIz"} {
		if _, err := ParseCursor(raw); err == nil {
			t.Errorf("ParseCursor(%q) succeeded, want an error", raw)
		}
	}
}

func TestNewPageOmitsCursorOnTheLastPage(t *testing.T) {
	page := NewPage([]string{"a"}, NewCursor(time.Now(), "a"), false)

	if page.NextCursor != nil {
		t.Errorf("next_cursor = %v, want nil on the last page", *page.NextCursor)
	}
	if page.HasMore {
		t.Error("has_more = true on the last page")
	}
}

func TestFreshnessClassification(t *testing.T) {
	thresholds := FreshnessThresholds{
		FreshMax:  5 * time.Minute,
		RecentMax: 15 * time.Minute,
		StaleMax:  60 * time.Minute,
	}

	cases := []struct {
		age  time.Duration
		want DataFreshness
	}{
		{time.Minute, FreshnessFresh},
		{5 * time.Minute, FreshnessRecent},
		{14 * time.Minute, FreshnessRecent},
		{15 * time.Minute, FreshnessStale},
		{59 * time.Minute, FreshnessStale},
		{61 * time.Minute, FreshnessVeryStale},
	}
	for _, tc := range cases {
		if got := thresholds.Classify(tc.age); got != tc.want {
			t.Errorf("Classify(%v) = %s, want %s", tc.age, got, tc.want)
		}
	}
}
