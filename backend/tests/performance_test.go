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

func TestGetPerformanceAllTimeAfterSync(t *testing.T) {
	pool := testsupport.Postgres(t)
	router, _, deps := newWalletsTestStack(t, pool)
	accessToken := guestAccessToken(t, router)

	walletID := createSyncedWallet(t, router, accessToken, deps)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/wallets/"+walletID.String()+"/performance?period=all", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("get performance status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var performance struct {
		Status    string  `json:"status"`
		Period    string  `json:"period"`
		ChangePct *string `json:"change_pct"`
		Series    []struct {
			TotalValueUSD *string `json:"total_value_usd"`
		} `json:"series"`
		CalculationID *string `json:"calculation_id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &performance); err != nil {
		t.Fatalf("decode performance: %v", err)
	}
	if performance.Period != "all" {
		t.Fatalf("period = %q, want all", performance.Period)
	}
	if performance.Status != "AVAILABLE" {
		t.Fatalf("status = %q, want AVAILABLE", performance.Status)
	}
	if performance.ChangePct == nil || *performance.ChangePct != "0" {
		t.Fatalf("change_pct = %v, want 0 for single snapshot", performance.ChangePct)
	}
	if len(performance.Series) != 1 {
		t.Fatalf("series length = %d, want 1", len(performance.Series))
	}
	if performance.CalculationID == nil || *performance.CalculationID == "" {
		t.Fatal("expected calculation_id to be set")
	}
}

func TestGetPerformance24hWithoutHistoryReturnsUnavailable(t *testing.T) {
	pool := testsupport.Postgres(t)
	router, _, deps := newWalletsTestStack(t, pool)
	accessToken := guestAccessToken(t, router)

	walletID := createSyncedWallet(t, router, accessToken, deps)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/wallets/"+walletID.String()+"/performance?period=24h", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("get performance status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var performance struct {
		Status            string  `json:"status"`
		UnavailableReason *string `json:"unavailable_reason"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &performance); err != nil {
		t.Fatalf("decode performance: %v", err)
	}
	if performance.Status != "UNAVAILABLE" {
		t.Fatalf("status = %q, want UNAVAILABLE", performance.Status)
	}
	if performance.UnavailableReason == nil || *performance.UnavailableReason != "PERFORMANCE_DATA_UNAVAILABLE" {
		t.Fatalf("unavailable_reason = %v", performance.UnavailableReason)
	}
}

func TestGetPerformance24hWithHistoricalSnapshot(t *testing.T) {
	pool := testsupport.Postgres(t)
	router, _, deps := newWalletsTestStack(t, pool)
	accessToken := guestAccessToken(t, router)

	walletID := createSyncedWallet(t, router, accessToken, deps)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := deps.Snapshots.Create(ctx, walletID, nil, time.Now().UTC().Add(-25*time.Hour)); err != nil {
		t.Fatalf("create historical snapshot: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/wallets/"+walletID.String()+"/performance?period=24h", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("get performance status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var performance struct {
		Status    string `json:"status"`
		ChangePct *string `json:"change_pct"`
		Series    []any  `json:"series"`
		Drivers   []struct {
			Asset struct {
				Symbol string `json:"symbol"`
			} `json:"asset"`
		} `json:"drivers"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &performance); err != nil {
		t.Fatalf("decode performance: %v", err)
	}
	if performance.Status != "AVAILABLE" {
		t.Fatalf("status = %q, want AVAILABLE", performance.Status)
	}
	if performance.ChangePct == nil {
		t.Fatal("expected change_pct to be set")
	}
	if len(performance.Series) < 2 {
		t.Fatalf("series length = %d, want at least 2", len(performance.Series))
	}
	if len(performance.Drivers) == 0 {
		t.Fatal("expected at least one driver")
	}
	if performance.Drivers[0].Asset.Symbol != "ETH" {
		t.Fatalf("driver symbol = %q, want ETH", performance.Drivers[0].Asset.Symbol)
	}
}

func TestGetPerformanceRequiresPeriod(t *testing.T) {
	pool := testsupport.Postgres(t)
	router, _, _ := newWalletsTestStack(t, pool)
	accessToken := guestAccessToken(t, router)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/wallets/"+uuid.New().String()+"/performance", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("get performance status = %d, want 400", rec.Code)
	}
}

func createSyncedWallet(t *testing.T, router http.Handler, accessToken string, deps testStack) uuid.UUID {
	t.Helper()

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
	if _, err := deps.Sync.Run(ctx, appsync.Request{
		WalletID: walletID,
		Trigger:  wallet.TriggerInitial,
		JobID:    jobs.SyncTaskID(walletID, wallet.TriggerInitial),
	}); err != nil {
		t.Fatalf("sync run failed: %v", err)
	}
	return walletID
}
