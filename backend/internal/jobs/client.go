package jobs

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/maxaicrypto/backend/internal/app/config"
	"github.com/maxaicrypto/backend/internal/domain/apperr"
	"github.com/maxaicrypto/backend/internal/domain/wallet"
)

// Client enqueues background work. It implements sync.Enqueuer so the HTTP
// layer depends on an application port rather than on Asynq (§57).
type Client struct {
	client *asynq.Client
	cfg    config.SyncConfig
}

// NewClient builds an enqueuer over the shared Redis connection.
func NewClient(opt asynq.RedisConnOpt, cfg config.SyncConfig) *Client {
	return &Client{client: asynq.NewClient(opt), cfg: cfg}
}

// Close releases the underlying connection.
func (c *Client) Close() error { return c.client.Close() }

// EnqueueInitialSync queues the first synchronization of a new wallet on the
// critical queue: the user is watching the sync screen (§95).
func (c *Client) EnqueueInitialSync(ctx context.Context, walletID uuid.UUID) error {
	task, err := NewInitialWalletSync(walletID)
	if err != nil {
		return apperr.Wrap(apperr.CodeInternal, err)
	}
	return c.enqueue(ctx, task,
		asynq.Queue(QueueCritical),
		asynq.TaskID(SyncTaskID(walletID, wallet.TriggerInitial)),
		asynq.MaxRetry(c.cfg.MaxRetries),
		asynq.Timeout(c.cfg.JobTimeout),
	)
}

// EnqueueSync queues a manual or scheduled synchronization.
func (c *Client) EnqueueSync(ctx context.Context, walletID uuid.UUID, trigger wallet.SyncTrigger) error {
	task, err := NewWalletSync(walletID, trigger)
	if err != nil {
		return apperr.Wrap(apperr.CodeInternal, err)
	}
	queue := QueueLow
	if trigger == wallet.TriggerManual {
		queue = QueueDefault
	}
	return c.enqueue(ctx, task,
		asynq.Queue(queue),
		asynq.TaskID(SyncTaskID(walletID, trigger)),
		asynq.MaxRetry(c.cfg.MaxRetries),
		asynq.Timeout(c.cfg.JobTimeout),
	)
}

// EnqueueSnapshot queues snapshot creation after a successful sync (§50).
func (c *Client) EnqueueSnapshot(ctx context.Context, walletID uuid.UUID, syncRunID *uuid.UUID) error {
	task, err := NewPortfolioSnapshot(walletID, syncRunID)
	if err != nil {
		return apperr.Wrap(apperr.CodeInternal, err)
	}
	return c.enqueue(ctx, task,
		asynq.Queue(QueueDefault),
		asynq.TaskID(SnapshotTaskID(walletID, syncRunID)),
		asynq.MaxRetry(c.cfg.MaxRetries),
		asynq.Timeout(c.cfg.JobTimeout),
	)
}

// EnqueuePriceRefresh queues a market price refresh.
func (c *Client) EnqueuePriceRefresh(ctx context.Context, assetIDs []uuid.UUID) error {
	task, err := NewPriceRefresh(assetIDs)
	if err != nil {
		return apperr.Wrap(apperr.CodeInternal, err)
	}
	return c.enqueue(ctx, task, asynq.Queue(QueueLow), asynq.MaxRetry(c.cfg.MaxRetries))
}

// enqueue submits a task, treating a duplicate deterministic ID as success:
// the work the caller wanted is already scheduled (§60).
func (c *Client) enqueue(ctx context.Context, task *asynq.Task, opts ...asynq.Option) error {
	if _, err := c.client.EnqueueContext(ctx, task, opts...); err != nil {
		if errors.Is(err, asynq.ErrTaskIDConflict) || errors.Is(err, asynq.ErrDuplicateTask) {
			return nil
		}
		return apperr.Wrap(apperr.CodeInternal, err)
	}
	return nil
}
