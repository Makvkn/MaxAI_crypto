//go:build integration

package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/maxaicrypto/backend/internal/domain/wallet"
	"github.com/maxaicrypto/backend/internal/jobs"
	"github.com/maxaicrypto/backend/tests/testsupport"
)

func TestCreateWalletEnqueuesInitialSync(t *testing.T) {
	pool := testsupport.Postgres(t)

	router, cfg := newWalletsTestRouter(t, pool)
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
		ID      string `json:"id"`
		Address string `json:"address"`
		Sync    struct {
			Status string `json:"status"`
		} `json:"sync"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode wallet: %v", err)
	}
	if created.Sync.Status != "PENDING" {
		t.Fatalf("sync.status = %q, want PENDING", created.Sync.Status)
	}
	if created.Address != "0x742d35cc6634c0532925a3b844bc9e7595f0beb0" {
		t.Fatalf("address = %q, want normalized lowercase", created.Address)
	}

	redisOpt, err := jobs.RedisOpt(cfg.Redis)
	if err != nil {
		t.Fatalf("redis opt: %v", err)
	}
	inspector := asynq.NewInspector(redisOpt)
	defer inspector.Close()

	walletID := uuid.MustParse(created.ID)
	taskID := jobs.SyncTaskID(walletID, wallet.TriggerInitial)
	if _, err := inspector.GetTaskInfo(jobs.QueueCritical, taskID); err != nil {
		t.Fatalf("expected initial sync task in critical queue: %v", err)
	}
}
