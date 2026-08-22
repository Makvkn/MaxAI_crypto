-- name: CreatePortfolioSnapshot :one
INSERT INTO portfolio_snapshots (
    wallet_id, captured_at, total_value_usd, status, data_quality, calculation_version, sync_run_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, wallet_id, captured_at, total_value_usd, status, data_quality,
          calculation_version, sync_run_id, created_at;

-- name: InsertSnapshotPosition :exec
INSERT INTO portfolio_snapshot_positions (
    snapshot_id, asset_id, balance, price_usd, value_usd, allocation_pct,
    price_timestamp, price_source
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (snapshot_id, asset_id) DO UPDATE
SET balance = EXCLUDED.balance,
    price_usd = EXCLUDED.price_usd,
    value_usd = EXCLUDED.value_usd,
    allocation_pct = EXCLUDED.allocation_pct,
    price_timestamp = EXCLUDED.price_timestamp,
    price_source = EXCLUDED.price_source;

-- name: GetLatestValidSnapshot :one
SELECT id, wallet_id, captured_at, total_value_usd, status, data_quality,
       calculation_version, sync_run_id, created_at
FROM portfolio_snapshots
WHERE wallet_id = $1
  AND status <> 'UNAVAILABLE'
  AND total_value_usd IS NOT NULL
ORDER BY captured_at DESC
LIMIT 1;

-- name: GetSnapshotBySyncRunID :one
SELECT id, wallet_id, captured_at, total_value_usd, status, data_quality,
       calculation_version, sync_run_id, created_at
FROM portfolio_snapshots
WHERE sync_run_id = $1;
