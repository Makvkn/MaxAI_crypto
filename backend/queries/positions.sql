-- name: ListPositionsByWallet :many
SELECT wallet_id, asset_id, balance_raw, balance_normalized, updated_at
FROM wallet_positions
WHERE wallet_id = $1;

-- name: GetPositionByAsset :one
SELECT wallet_id, asset_id, balance_raw, balance_normalized, updated_at
FROM wallet_positions
WHERE wallet_id = $1
  AND asset_id = $2;

-- name: DeletePositionsByWallet :exec
DELETE FROM wallet_positions
WHERE wallet_id = $1;

-- name: InsertPosition :exec
INSERT INTO wallet_positions (wallet_id, asset_id, balance_raw, balance_normalized, updated_at)
VALUES ($1, $2, $3, $4, $5);
