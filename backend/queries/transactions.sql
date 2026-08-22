-- name: UpsertTransaction :one
INSERT INTO transactions (
    wallet_id, chain_id, tx_hash, log_index, block_number, timestamp, status, type,
    from_address, to_address, asset_in_id, amount_in, asset_out_id, amount_out,
    fee_asset_id, fee_amount, protocol, counterparty, raw_reference
)
VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8,
    $9, $10, $11, $12, $13, $14,
    $15, $16, $17, $18, $19
)
ON CONFLICT ON CONSTRAINT transactions_identity_key DO UPDATE
SET block_number = EXCLUDED.block_number,
    timestamp = EXCLUDED.timestamp,
    status = EXCLUDED.status,
    type = EXCLUDED.type,
    from_address = EXCLUDED.from_address,
    to_address = EXCLUDED.to_address,
    asset_in_id = EXCLUDED.asset_in_id,
    amount_in = EXCLUDED.amount_in,
    asset_out_id = EXCLUDED.asset_out_id,
    amount_out = EXCLUDED.amount_out,
    fee_asset_id = EXCLUDED.fee_asset_id,
    fee_amount = EXCLUDED.fee_amount,
    protocol = EXCLUDED.protocol,
    counterparty = EXCLUDED.counterparty,
    raw_reference = EXCLUDED.raw_reference,
    updated_at = NOW()
RETURNING id;

-- name: ListTransactionsByWallet :many
SELECT id, wallet_id, chain_id, tx_hash, log_index, block_number, timestamp, status, type,
       from_address, to_address, asset_in_id, amount_in, asset_out_id, amount_out,
       fee_asset_id, fee_amount, protocol, counterparty, raw_reference, created_at, updated_at
FROM transactions
WHERE wallet_id = $1
  AND ($2::text IS NULL OR type = $2)
  AND (
    sqlc.narg('cursor_at')::timestamptz IS NULL
    OR (timestamp, id) < (sqlc.narg('cursor_at')::timestamptz, sqlc.narg('cursor_id')::uuid)
  )
ORDER BY timestamp DESC, id DESC
LIMIT $3;

-- name: GetTransactionByID :one
SELECT id, wallet_id, chain_id, tx_hash, log_index, block_number, timestamp, status, type,
       from_address, to_address, asset_in_id, amount_in, asset_out_id, amount_out,
       fee_asset_id, fee_amount, protocol, counterparty, raw_reference, created_at, updated_at
FROM transactions
WHERE id = $1;
