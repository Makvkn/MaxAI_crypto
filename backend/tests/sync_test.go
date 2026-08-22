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
	walletrepo "github.com/maxaicrypto/backend/internal/infrastructure/postgres/wallet"
	"github.com/maxaicrypto/backend/tests/testsupport"
)

func TestWalletSyncWorkerCompletes(t *testing.T) {
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

	jobID := jobs.SyncTaskID(walletID, wallet.TriggerInitial)
	result, err := deps.Sync.Run(ctx, appsync.Request{
		WalletID: walletID,
		Trigger:  wallet.TriggerInitial,
		JobID:    jobID,
	})
	if err != nil {
		t.Fatalf("sync run failed: %v", err)
	}
	if result.Status != wallet.SyncReady && result.Status != wallet.SyncPartial {
		t.Fatalf("sync status = %q, want READY or PARTIAL", result.Status)
	}
	if result.SnapshotID == nil {
		t.Fatal("expected snapshot to be created")
	}

	syncRepo := walletrepo.NewSyncRepository(pool)
	state, err := syncRepo.Get(ctx, walletID)
	if err != nil {
		t.Fatalf("load sync state: %v", err)
	}
	if state.Status != wallet.SyncReady && state.Status != wallet.SyncPartial {
		t.Fatalf("persisted sync status = %q", state.Status)
	}
	if len(state.StagesCompleted) == 0 {
		t.Fatal("expected completed stages to be recorded")
	}
}
