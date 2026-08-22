// Package shared holds the domain primitives that several aggregates depend
// on: exact decimal money, opaque pagination cursors and the data-quality
// vocabulary. It must never import other domain packages.
package shared

import (
	"encoding/json"
	"fmt"

	"github.com/shopspring/decimal"
)

// Decimal is the canonical numeric type for every financial value: balances,
// prices, portfolio values, percentages and fees. Floating point is forbidden
// for these values (§14, §112), and the JSON form is a string so that exactness
// survives the wire (§97).
type Decimal struct {
	value decimal.Decimal
}

// Zero is an exact zero, which is semantically different from an unknown value.
// Unknown is modelled by NullDecimal, never by zero (§40).
var Zero = Decimal{}

// NewDecimal wraps an underlying decimal value.
func NewDecimal(v decimal.Decimal) Decimal { return Decimal{value: v} }

// NewDecimalFromInt builds a Decimal from an integer.
func NewDecimalFromInt(v int64) Decimal { return Decimal{value: decimal.NewFromInt(v)} }

// ParseDecimal parses an exact decimal string. External APIs that return JSON
// floating-point numbers must be normalized through this function immediately
// on ingest (§112).
func ParseDecimal(s string) (Decimal, error) {
	v, err := decimal.NewFromString(s)
	if err != nil {
		return Decimal{}, fmt.Errorf("parse decimal %q: %w", s, err)
	}
	return Decimal{value: v}, nil
}

// Value exposes the underlying decimal for arithmetic.
func (d Decimal) Value() decimal.Decimal { return d.value }

// String renders the exact decimal representation.
func (d Decimal) String() string { return d.value.String() }

// IsZero reports whether the value is exactly zero.
func (d Decimal) IsZero() bool { return d.value.IsZero() }

// MarshalJSON emits a JSON string, not a JSON number.
func (d Decimal) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.value.String())
}

// UnmarshalJSON accepts a JSON string. Numbers are rejected because a JSON
// number has already lost exactness by the time it reaches this method.
func (d *Decimal) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("decimal values must be JSON strings: %w", err)
	}
	parsed, err := ParseDecimal(s)
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}

// NullDecimal is a decimal that may be unknown. It exists so the backend can
// preserve the difference between an actual zero and a missing value all the
// way to the API, where unknown is serialized as null (§40).
type NullDecimal struct {
	Decimal Decimal
	Valid   bool
}

// Known builds a present value.
func Known(d Decimal) NullDecimal { return NullDecimal{Decimal: d, Valid: true} }

// Unknown builds an absent value.
func Unknown() NullDecimal { return NullDecimal{} }

// MarshalJSON emits null when the value is unknown.
func (n NullDecimal) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return n.Decimal.MarshalJSON()
}

// UnmarshalJSON accepts null or a decimal string.
func (n *NullDecimal) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*n = NullDecimal{}
		return nil
	}
	var d Decimal
	if err := d.UnmarshalJSON(data); err != nil {
		return err
	}
	*n = Known(d)
	return nil
}
