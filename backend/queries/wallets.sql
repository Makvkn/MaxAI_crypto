-- name: CreateWallet :one
INSERT INTO wallets (user_id, chain_id, address, label, status)
VALUES ($1, $2, $3, $4, 'ACTIVE')
RETURNING id, user_id, chain_id, address, label, status, created_at, updated_at, deleted_at;

-- name: GetWalletByID :one
SELECT id, user_id, chain_id, address, label, status, created_at, updated_at, deleted_at
FROM wallets
WHERE id = $1
  AND deleted_at IS NULL;

-- name: FindWalletByUserChainAddress :one
SELECT id, user_id, chain_id, address, label, status, created_at, updated_at, deleted_at
FROM wallets
WHERE user_id = $1
  AND chain_id = $2
  AND address = $3
  AND deleted_at IS NULL;

-- name: ListWalletsByUser :many
SELECT id, user_id, chain_id, address, label, status, created_at, updated_at, deleted_at
FROM wallets
WHERE user_id = $1
  AND deleted_at IS NULL
  AND (
    sqlc.narg('cursor_at')::timestamptz IS NULL
    OR (created_at, id) < (sqlc.narg('cursor_at')::timestamptz, sqlc.narg('cursor_id')::uuid)
  )
ORDER BY created_at DESC, id DESC
LIMIT $2;

-- name: CountWalletsByUser :one
SELECT COUNT(*)::int AS count
FROM wallets
WHERE user_id = $1
  AND deleted_at IS NULL;

-- name: SoftDeleteWallet :exec
UPDATE wallets
SET status = 'DELETED',
    deleted_at = $2,
    updated_at = $2
WHERE id = $1
  AND deleted_at IS NULL;

-- name: UpdateWalletStatus :exec
UPDATE wallets
SET status = $2,
    updated_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL;

-- name: ListWalletsDueForSync :many
SELECT w.id, w.user_id, w.chain_id, w.address, w.label, w.status, w.created_at, w.updated_at, w.deleted_at
FROM wallets AS w
JOIN wallet_sync_states AS s ON s.wallet_id = w.id
WHERE w.deleted_at IS NULL
  AND s.status <> 'SYNCING'
  AND (s.last_synced_at IS NULL OR s.last_synced_at < $1)
ORDER BY s.last_synced_at NULLS FIRST
LIMIT $2;
