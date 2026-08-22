package asset

import (
	"context"
	"strings"

	"github.com/shopspring/decimal"
)

// RulesVisibilityClassifier applies deterministic spam/dust rules (§43).
type RulesVisibilityClassifier struct {
	dustUSD decimal.Decimal
}

// NewRulesVisibilityClassifier builds the default classifier.
// Positions valued below dustUSD are HIDDEN_DUST; obvious spam markers become
// HIDDEN_SPAM; everything else stays VISIBLE (or UNKNOWN without price for
// non-native tokens that also lack market data — still not treated as spam).
func NewRulesVisibilityClassifier() *RulesVisibilityClassifier {
	return &RulesVisibilityClassifier{dustUSD: decimal.NewFromFloat(0.01)}
}

// Classify implements VisibilityClassifier.
func (c *RulesVisibilityClassifier) Classify(_ context.Context, input VisibilityInput) (Visibility, error) {
	symbol := strings.ToUpper(strings.TrimSpace(input.Asset.Symbol))
	name := strings.ToLower(strings.TrimSpace(input.Asset.Name))

	if looksLikeSpam(symbol, name) {
		return VisibilityHiddenSpam, nil
	}

	if input.ValueUSD != nil {
		value, err := decimal.NewFromString(*input.ValueUSD)
		if err == nil && value.GreaterThan(decimal.Zero) && value.LessThan(c.dustUSD) {
			return VisibilityHiddenDust, nil
		}
	}

	if !input.HasMarketData && input.Asset.Type == TypeToken {
		return VisibilityUnknown, nil
	}

	return VisibilityVisible, nil
}

func looksLikeSpam(symbol, nameLower string) bool {
	if symbol == "" || symbol == "UNKNOWN" {
		return true
	}
	markers := []string{
		"airdrop", "claim reward", "visit ", "http", "www.", ".com", ".io",
		"telegram", "t.me/", "free mint", "congratulations",
	}
	for _, marker := range markers {
		if strings.Contains(nameLower, marker) {
			return true
		}
	}
	if len(symbol) > 20 {
		return true
	}
	return false
}

var _ VisibilityClassifier = (*RulesVisibilityClassifier)(nil)
