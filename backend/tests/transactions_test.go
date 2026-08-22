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

func TestListTransactionsAfterSync(t *testing.T) {
	pool := testsupport.Postgres(t)
	router, _, deps := newWalletsTestStack(t, pool)
	accessToken := guestAccessToken(t, router)

	walletID := createTestWallet(t, router, accessToken, "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	runSync(t, deps, walletID)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/wallets/"+walletID.String()+"/transactions", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list transactions status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var page struct {
		Items []struct {
			TxHash    string `json:"tx_hash"`
			Type      string `json:"type"`
			AssetIn   *struct {
				Symbol string `json:"symbol"`
			} `json:"asset_in"`
			ExplorerURL *string `json:"explorer_url"`
		} `json:"items"`
		HasMore bool `json:"has_more"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
		t.Fatalf("decode transactions: %v", err)
	}
	if len(page.Items) == 0 {
		t.Fatal("expected at least one transaction")
	}
	if page.Items[0].Type != "UNKNOWN" {
		t.Fatalf("type = %q, want UNKNOWN from fixture classifier", page.Items[0].Type)
	}
	if page.Items[0].AssetIn == nil || page.Items[0].AssetIn.Symbol != "ETH" {
		t.Fatal("expected inbound ETH transfer")
	}
	if page.Items[0].ExplorerURL == nil || *page.Items[0].ExplorerURL == "" {
		t.Fatal("expected explorer_url")
	}
}

func TestGetTransactionAfterSync(t *testing.T) {
	pool := testsupport.Postgres(t)
	router, _, deps := newWalletsTestStack(t, pool)
	accessToken := guestAccessToken(t, router)

	walletID := createTestWallet(t, router, accessToken, "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	runSync(t, deps, walletID)

	listRec := httptest.NewRecorder()
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/wallets/"+walletID.String()+"/transactions", nil)
	listReq.Header.Set("Authorization", "Bearer "+accessToken)
	router.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status = %d", listRec.Code)
	}
	var page struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	if err := json.Unmarshal(listRec.Body.Bytes(), &page); err != nil {
		t.Fatalf("decode list: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/wallets/"+walletID.String()+"/transactions/"+page.Items[0].ID, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("get transaction status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func createTestWallet(t *testing.T, router http.Handler, accessToken, address string) uuid.UUID {
	t.Helper()
	body, _ := json.Marshal(map[string]string{
		"chain_id": "ethereum",
		"address":  address,
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
	return uuid.MustParse(created.ID)
}

func runSync(t *testing.T, deps testStack, walletID uuid.UUID) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if _, err := deps.Sync.Run(ctx, appsync.Request{
		WalletID: walletID,
		Trigger:  wallet.TriggerInitial,
		JobID:    jobs.SyncTaskID(walletID, wallet.TriggerInitial),
	}); err != nil {
		t.Fatalf("sync run failed: %v", err)
	}
}
