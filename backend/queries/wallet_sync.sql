-- name: CreateWalletSyncState :one
INSERT INTO wallet_sync_states (wallet_id, status)
VALUES ($1, 'PENDING')
RETURNING wallet_id, status, stage, stages_completed, started_at, completed_at,
          last_synced_at, error_code, error_message, sync_job_id, updated_at;

-- name: GetWalletSyncState :one
SELECT wallet_id, status, stage, stages_completed, started_at, completed_at,
       last_synced_at, error_code, error_message, sync_job_id, updated_at
FROM wallet_sync_states
WHERE wallet_id = $1;

-- name: ListWalletSyncStates :many
SELECT wallet_id, status, stage, stages_completed, started_at, completed_at,
       last_synced_at, error_code, error_message, sync_job_id, updated_at
FROM wallet_sync_states
WHERE wallet_id = ANY($1::uuid[]);

-- name: UpdateWalletSyncState :one
UPDATE wallet_sync_states
SET status = $2,
    stage = $3,
    stages_completed = $4,
    started_at = $5,
    completed_at = $6,
    last_synced_at = $7,
    error_code = $8,
    error_message = $9,
    sync_job_id = $10,
    updated_at = NOW()
WHERE wallet_id = $1
RETURNING wallet_id, status, stage, stages_completed, started_at, completed_at,
          last_synced_at, error_code, error_message, sync_job_id, updated_at;

-- name: StartWalletSyncRun :one
INSERT INTO wallet_sync_runs (wallet_id, job_id, trigger, provider, status, started_at)
VALUES ($1, $2, $3, $4, 'SYNCING', NOW())
RETURNING id, wallet_id, job_id, trigger, provider, status, started_at, finished_at, error_code, error_text;

-- name: FinishWalletSyncRun :exec
UPDATE wallet_sync_runs
SET status = $2,
    finished_at = $3,
    error_code = $4,
    error_text = $5
WHERE id = $1;

-- name: GetWalletSyncRunByJobID :one
SELECT id, wallet_id, job_id, trigger, provider, status, started_at, finished_at, error_code, error_text
FROM wallet_sync_runs
WHERE job_id = $1;
