// Package jobs defines the background job contract: task types, payloads and
// deterministic identifiers (§58, §59, §60).
package jobs

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/maxaicrypto/backend/internal/domain/wallet"
)

// Task types from §58.
const (
	TypeInitialWalletSync = "wallet:initial_sync"
	TypeWalletSync        = "wallet:sync"
	TypePortfolioSnapshot = "portfolio:snapshot"
	TypePriceRefresh      = "price:refresh"
	TypeSyncScheduler     = "wallet:sync_scheduler"
)

// Queues separate latency-sensitive work from periodic maintenance so a large
// scheduled batch cannot delay a user-triggered sync.
const (
	QueueCritical = "critical"
	QueueDefault  = "default"
	QueueLow      = "low"
)

// WalletSyncPayload identifies the wallet to synchronize and why.
type WalletSyncPayload struct {
	WalletID uuid.UUID          `json:"wallet_id"`
	Trigger  wallet.SyncTrigger `json:"trigger"`
}

// PortfolioSnapshotPayload identifies the wallet to snapshot. SyncRunID links
// the snapshot to the run that produced it, which also makes a retried job
// recognise an already-written snapshot (§60).
type PortfolioSnapshotPayload struct {
	WalletID  uuid.UUID  `json:"wallet_id"`
	SyncRunID *uuid.UUID `json:"sync_run_id,omitempty"`
}

// PriceRefreshPayload optionally narrows a refresh to specific assets. An empty
// list means every asset currently held by any wallet.
type PriceRefreshPayload struct {
	AssetIDs []uuid.UUID `json:"asset_ids,omitempty"`
}

// NewInitialWalletSync builds the task enqueued right after wallet creation, so
// the HTTP request returns without waiting for the sync (§57).
func NewInitialWalletSync(walletID uuid.UUID) (*asynq.Task, error) {
	payload, err := json.Marshal(WalletSyncPayload{WalletID: walletID, Trigger: wallet.TriggerInitial})
	if err != nil {
		return nil, fmt.Errorf("marshal initial sync payload: %w", err)
	}
	return asynq.NewTask(TypeInitialWalletSync, payload), nil
}

// NewWalletSync builds a synchronization task for an existing wallet.
func NewWalletSync(walletID uuid.UUID, trigger wallet.SyncTrigger) (*asynq.Task, error) {
	payload, err := json.Marshal(WalletSyncPayload{WalletID: walletID, Trigger: trigger})
	if err != nil {
		return nil, fmt.Errorf("marshal wallet sync payload: %w", err)
	}
	return asynq.NewTask(TypeWalletSync, payload), nil
}

// NewPortfolioSnapshot builds a snapshot task.
func NewPortfolioSnapshot(walletID uuid.UUID, syncRunID *uuid.UUID) (*asynq.Task, error) {
	payload, err := json.Marshal(PortfolioSnapshotPayload{WalletID: walletID, SyncRunID: syncRunID})
	if err != nil {
		return nil, fmt.Errorf("marshal snapshot payload: %w", err)
	}
	return asynq.NewTask(TypePortfolioSnapshot, payload), nil
}

// NewPriceRefresh builds a price refresh task.
func NewPriceRefresh(assetIDs []uuid.UUID) (*asynq.Task, error) {
	payload, err := json.Marshal(PriceRefreshPayload{AssetIDs: assetIDs})
	if err != nil {
		return nil, fmt.Errorf("marshal price refresh payload: %w", err)
	}
	return asynq.NewTask(TypePriceRefresh, payload), nil
}

// SyncTaskID is the deterministic identifier of a wallet sync task. Asynq drops
// a duplicate enqueue of the same ID, so a user hammering refresh cannot stack
// up identical syncs (§60).
func SyncTaskID(walletID uuid.UUID, trigger wallet.SyncTrigger) string {
	return fmt.Sprintf("%s:%s:%s", TypeWalletSync, walletID, trigger)
}

// SnapshotTaskID is the deterministic identifier of a snapshot task, scoped to
// the sync run that produced the data.
func SnapshotTaskID(walletID uuid.UUID, syncRunID *uuid.UUID) string {
	if syncRunID == nil {
		return fmt.Sprintf("%s:%s", TypePortfolioSnapshot, walletID)
	}
	return fmt.Sprintf("%s:%s:%s", TypePortfolioSnapshot, walletID, *syncRunID)
}
