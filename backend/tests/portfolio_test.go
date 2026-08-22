//go:build integration

package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	appsync "github.com/maxaicrypto/backend/internal/application/sync"
	"github.com/maxaicrypto/backend/internal/domain/wallet"
	"github.com/maxaicrypto/backend/internal/jobs"
	"github.com/maxaicrypto/backend/tests/testsupport"
)

func TestGetPortfolioAfterSync(t *testing.T) {
	pool := testsupport.Postgres(t)
	router, _, deps := newWalletsTestStack(t, pool)
	accessToken := guestAccessToken(t, router)

	body, _ := json.Marshal(map[string]string{
		"chain_id": "ethereum",
		"address":  "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/wallets", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create wallet status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var created struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode wallet: %v", err)
	}
	walletID := uuid.MustParse(created.ID)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, err := deps.Sync.Run(ctx, appsync.Request{
		WalletID: walletID,
		Trigger:  wallet.TriggerInitial,
		JobID:    jobs.SyncTaskID(walletID, wallet.TriggerInitial),
	})
	if err != nil {
		t.Fatalf("sync run failed: %v", err)
	}

	rec = httptest.NewRecorder()
	portfolioReq := httptest.NewRequest(http.MethodGet, "/api/v1/wallets/"+walletID.String()+"/portfolio", nil)
	portfolioReq.Header.Set("Authorization", "Bearer "+accessToken)
	router.ServeHTTP(rec, portfolioReq)
	if rec.Code != http.StatusOK {
		t.Fatalf("get portfolio status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var portfolio struct {
		TotalValueUSD *string `json:"total_value_usd"`
		ValuationStatus string `json:"valuation_status"`
		Positions []struct {
			Asset struct {
				Symbol string `json:"symbol"`
			} `json:"asset"`
			Balance string `json:"balance"`
		} `json:"positions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &portfolio); err != nil {
		t.Fatalf("decode portfolio: %v", err)
	}
	if portfolio.ValuationStatus != "COMPLETE" && portfolio.ValuationStatus != "PARTIAL" {
		t.Fatalf("valuation_status = %q", portfolio.ValuationStatus)
	}
	if portfolio.TotalValueUSD == nil || *portfolio.TotalValueUSD == "" {
		t.Fatal("expected total_value_usd to be set")
	}
	if len(portfolio.Positions) == 0 {
		t.Fatal("expected at least one position")
	}
	if portfolio.Positions[0].Asset.Symbol != "ETH" {
		t.Fatalf("position symbol = %q, want ETH", portfolio.Positions[0].Asset.Symbol)
	}
}

func TestGetPortfolioBeforeSyncReturnsNotReady(t *testing.T) {
	pool := testsupport.Postgres(t)
	router, _, _ := newWalletsTestStack(t, pool)
	accessToken := guestAccessToken(t, router)

	body, _ := json.Marshal(map[string]string{
		"chain_id": "ethereum",
		"address":  "0x1234567890123456789012345678901234567890",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/wallets", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create wallet status = %d", rec.Code)
	}

	var created struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &created)

	rec = httptest.NewRecorder()
	portfolioReq := httptest.NewRequest(http.MethodGet, "/api/v1/wallets/"+created.ID+"/portfolio", nil)
	portfolioReq.Header.Set("Authorization", "Bearer "+accessToken)
	router.ServeHTTP(rec, portfolioReq)
	if rec.Code != http.StatusConflict {
		t.Fatalf("get portfolio status = %d, want 409 WALLET_NOT_READY", rec.Code)
	}
	var errBody struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &errBody); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if errBody.Error.Code != "WALLET_NOT_READY" {
		t.Fatalf("error code = %q, want WALLET_NOT_READY", errBody.Error.Code)
	}
}
