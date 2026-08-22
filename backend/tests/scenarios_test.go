//go:build integration

package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/maxaicrypto/backend/tests/testsupport"
)

func TestSimulateScenarioAfterSync(t *testing.T) {
	pool := testsupport.Postgres(t)
	router, _, deps := newWalletsTestStack(t, pool)
	accessToken := guestAccessToken(t, router)
	walletID := createSyncedWallet(t, router, accessToken, deps)

	rec := httptest.NewRecorder()
	portfolioReq := httptest.NewRequest(http.MethodGet, "/api/v1/wallets/"+walletID.String()+"/portfolio", nil)
	portfolioReq.Header.Set("Authorization", "Bearer "+accessToken)
	router.ServeHTTP(rec, portfolioReq)
	if rec.Code != http.StatusOK {
		t.Fatalf("get portfolio status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var portfolio struct {
		TotalValueUSD *string `json:"total_value_usd"`
		Positions []struct {
			Asset struct {
				ID     string `json:"id"`
				Symbol string `json:"symbol"`
			} `json:"asset"`
			ValueUSD *string `json:"value_usd"`
		} `json:"positions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &portfolio); err != nil {
		t.Fatalf("decode portfolio: %v", err)
	}
	if len(portfolio.Positions) == 0 || portfolio.Positions[0].Asset.ID == "" {
		t.Fatal("expected a portfolio position with asset id")
	}
	assetID := portfolio.Positions[0].Asset.ID

	body, _ := json.Marshal(map[string]string{
		"wallet_id":  walletID.String(),
		"type":       "ASSET_PRICE_CHANGE",
		"asset_id":   assetID,
		"change_pct": "-20",
	})
	rec = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/scenarios", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("simulate scenario status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var result struct {
		Type       string `json:"type"`
		ChangePct  string `json:"change_pct"`
		Asset      struct {
			Symbol string `json:"symbol"`
		} `json:"asset"`
		Baseline struct {
			PortfolioValueUSD *string `json:"portfolio_value_usd"`
		} `json:"baseline"`
		Projection struct {
			PortfolioChangePct *string `json:"portfolio_change_pct"`
			AssetImpactUSD     *string `json:"asset_impact_usd"`
		} `json:"projection"`
		CalculationID string `json:"calculation_id"`
		Explanation   *struct {
			Intent string `json:"intent"`
			Answer string `json:"answer"`
		} `json:"explanation"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode scenario: %v", err)
	}
	if result.Type != "ASSET_PRICE_CHANGE" {
		t.Fatalf("type = %q", result.Type)
	}
	if result.ChangePct != "-20" {
		t.Fatalf("change_pct = %q", result.ChangePct)
	}
	if result.Asset.Symbol != "ETH" {
		t.Fatalf("asset symbol = %q, want ETH", result.Asset.Symbol)
	}
	if result.CalculationID == "" {
		t.Fatal("expected calculation_id")
	}
	if result.Projection.AssetImpactUSD == nil {
		t.Fatal("expected asset_impact_usd")
	}
	if result.Explanation == nil || result.Explanation.Intent != "SCENARIO_SIMULATION" {
		t.Fatalf("explanation = %+v", result.Explanation)
	}

	rec = httptest.NewRecorder()
	usageReq := httptest.NewRequest(http.MethodGet, "/api/v1/ai/usage", nil)
	usageReq.Header.Set("Authorization", "Bearer "+accessToken)
	router.ServeHTTP(rec, usageReq)
	var usage struct {
		Used int `json:"used"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &usage)
	if usage.Used != 1 {
		t.Fatalf("used = %d, want 1 after scenario", usage.Used)
	}
}

func TestSimulateScenarioRejectsUnknownAsset(t *testing.T) {
	pool := testsupport.Postgres(t)
	router, _, deps := newWalletsTestStack(t, pool)
	accessToken := guestAccessToken(t, router)
	walletID := createSyncedWallet(t, router, accessToken, deps)

	body, _ := json.Marshal(map[string]string{
		"wallet_id":  walletID.String(),
		"type":       "ASSET_PRICE_CHANGE",
		"asset_id":   "00000000-0000-0000-0000-000000000099",
		"change_pct": "-20",
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/scenarios", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest && rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 400/422, body = %s", rec.Code, rec.Body.String())
	}
}
