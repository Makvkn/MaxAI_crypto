package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/hibiken/asynq"

	"github.com/maxaicrypto/backend/internal/application/portfolio"
	"github.com/maxaicrypto/backend/internal/application/pricing"
	appsync "github.com/maxaicrypto/backend/internal/application/sync"
	"github.com/maxaicrypto/backend/internal/domain/wallet"
	"github.com/maxaicrypto/backend/internal/infrastructure/observability"
)

// Handlers executes background work by delegating to application services. Job
// code owns scheduling, retries and logging; business rules stay in the
// services (§204).
type Handlers struct {
	sync      appsync.Service
	snapshots portfolio.SnapshotService
	prices    pricing.Service
	client    *Client
	logger    *slog.Logger
}

// NewHandlers wires the job handlers.
func NewHandlers(
	syncService appsync.Service,
	snapshots portfolio.SnapshotService,
	prices pricing.Service,
	client *Client,
	logger *slog.Logger,
) *Handlers {
	return &Handlers{sync: syncService, snapshots: snapshots, prices: prices, client: client, logger: logger}
}

// Register binds every task type to its handler.
func (h *Handlers) Register(mux *asynq.ServeMux) {
	mux.HandleFunc(TypeInitialWalletSync, h.handleWalletSync)
	mux.HandleFunc(TypeWalletSync, h.handleWalletSync)
	mux.HandleFunc(TypePortfolioSnapshot, h.handlePortfolioSnapshot)
	mux.HandleFunc(TypePriceRefresh, h.handlePriceRefresh)
	mux.HandleFunc(TypeSyncScheduler, h.handleSyncScheduler)
}

func (h *Handlers) handleWalletSync(ctx context.Context, task *asynq.Task) error {
	payload, err := decode[WalletSyncPayload](task)
	if err != nil {
		return err
	}
	if h.sync == nil {
		return errNotWired("wallet sync")
	}

	ctx, logger := h.jobContext(ctx, slog.String(observability.FieldWalletID, payload.WalletID.String()))
	jobID, _ := asynq.GetTaskID(ctx)

	result, err := h.sync.Run(ctx, appsync.Request{
		WalletID: payload.WalletID,
		Trigger:  payload.Trigger,
		JobID:    jobID,
	})
	if err != nil {
		logger.ErrorContext(ctx, "wallet sync failed", slog.Any(observability.FieldError, err))
		return err
	}

	logger.InfoContext(ctx, "wallet sync completed",
		slog.String(observability.FieldStatus, string(result.Status)),
		slog.Int("balances", result.BalancesFetched),
		slog.Int("transactions", result.TransactionsFetched),
	)
	return nil
}

func (h *Handlers) handlePortfolioSnapshot(ctx context.Context, task *asynq.Task) error {
	payload, err := decode[PortfolioSnapshotPayload](task)
	if err != nil {
		return err
	}
	if h.snapshots == nil {
		return errNotWired("portfolio snapshot")
	}

	ctx, logger := h.jobContext(ctx, slog.String(observability.FieldWalletID, payload.WalletID.String()))

	snapshot, err := h.snapshots.Create(ctx, payload.WalletID, payload.SyncRunID, time.Now().UTC())
	if err != nil {
		logger.ErrorContext(ctx, "snapshot creation failed", slog.Any(observability.FieldError, err))
		return err
	}

	logger.InfoContext(ctx, "snapshot created", slog.String("snapshot_id", snapshot.ID.String()))
	return nil
}

func (h *Handlers) handlePriceRefresh(ctx context.Context, task *asynq.Task) error {
	payload, err := decode[PriceRefreshPayload](task)
	if err != nil {
		return err
	}
	if h.prices == nil {
		return errNotWired("price refresh")
	}

	ctx, logger := h.jobContext(ctx)

	refreshed, err := h.prices.Refresh(ctx, payload.AssetIDs)
	if err != nil {
		logger.ErrorContext(ctx, "price refresh failed", slog.Any(observability.FieldError, err))
		return err
	}

	logger.InfoContext(ctx, "prices refreshed", slog.Int("assets", refreshed))
	return nil
}

// handleSyncScheduler fans out one sync task per wallet whose data is older
// than the configured interval (§62). Deterministic task IDs make a repeated
// fan-out collapse onto the already queued work (§60).
func (h *Handlers) handleSyncScheduler(ctx context.Context, _ *asynq.Task) error {
	if h.sync == nil {
		return errNotWired("sync scheduler")
	}

	ctx, logger := h.jobContext(ctx)

	walletIDs, err := h.sync.ListDue(ctx, schedulerBatchSize)
	if err != nil {
		logger.ErrorContext(ctx, "listing due wallets failed", slog.Any(observability.FieldError, err))
		return err
	}

	var enqueued int
	for _, walletID := range walletIDs {
		if err := h.client.EnqueueSync(ctx, walletID, wallet.TriggerScheduled); err != nil {
			// One wallet must not abort the whole batch; the next tick retries.
			logger.WarnContext(ctx, "enqueue scheduled sync failed",
				slog.String(observability.FieldWalletID, walletID.String()),
				slog.Any(observability.FieldError, err),
			)
			continue
		}
		enqueued++
	}

	logger.InfoContext(ctx, "scheduled syncs enqueued",
		slog.Int("due", len(walletIDs)),
		slog.Int("enqueued", enqueued),
	)
	return nil
}

// schedulerBatchSize bounds one fan-out so a large user base cannot flood the
// queue in a single tick.
const schedulerBatchSize = 500

// jobContext attaches the job identifier and attempt to the logger and puts it
// on the context, so service-level logs stay correlated (§123).
func (h *Handlers) jobContext(ctx context.Context, attrs ...slog.Attr) (context.Context, *slog.Logger) {
	jobID, _ := asynq.GetTaskID(ctx)
	retried, _ := asynq.GetRetryCount(ctx)

	args := make([]any, 0, len(attrs)+2)
	args = append(args, slog.String(observability.FieldJobID, jobID), slog.Int("attempt", retried+1))
	for _, attr := range attrs {
		args = append(args, attr)
	}

	logger := h.logger.With(args...)
	return observability.WithLogger(ctx, logger), logger
}

// errNotWired reports that a slice's application service has not been built
// yet. Retrying cannot help, so the task fails loudly instead of looping.
func errNotWired(name string) error {
	return fmt.Errorf("%w: %s service is not wired yet", asynq.SkipRetry, name)
}

// decode parses a task payload. A malformed payload never becomes valid, so it
// skips retries instead of consuming the retry budget.
func decode[T any](task *asynq.Task) (T, error) {
	var payload T
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return payload, fmt.Errorf("%w: unmarshal %s payload: %v", asynq.SkipRetry, task.Type(), err)
	}
	return payload, nil
}
