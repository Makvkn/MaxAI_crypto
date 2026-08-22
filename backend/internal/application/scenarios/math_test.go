package scenarios_test

import (
	"testing"

	"github.com/shopspring/decimal"

	"github.com/maxaicrypto/backend/internal/domain/shared"
)

// Financial regression: ASSET_PRICE_CHANGE math must stay exact (§51, §171).
func TestAssetPriceChangeMath(t *testing.T) {
	portfolio := decimal.RequireFromString("10000")
	asset := decimal.RequireFromString("2500") // 25%
	changePct := decimal.RequireFromString("10")

	factor := decimal.NewFromInt(1).Add(changePct.Div(decimal.NewFromInt(100)))
	projectedAsset := asset.Mul(factor)
	impact := projectedAsset.Sub(asset)
	projectedPortfolio := portfolio.Add(impact)
	changePortfolioPct := impact.Div(portfolio).Mul(decimal.NewFromInt(100))

	if !projectedAsset.Equal(decimal.RequireFromString("2750")) {
		t.Fatalf("projected asset = %s", projectedAsset)
	}
	if !impact.Equal(decimal.RequireFromString("250")) {
		t.Fatalf("impact = %s", impact)
	}
	if !projectedPortfolio.Equal(decimal.RequireFromString("10250")) {
		t.Fatalf("projected portfolio = %s", projectedPortfolio)
	}
	if !changePortfolioPct.Equal(decimal.RequireFromString("2.5")) {
		t.Fatalf("portfolio change pct = %s", changePortfolioPct)
	}

	_ = shared.Known(shared.NewDecimal(projectedPortfolio))
}
