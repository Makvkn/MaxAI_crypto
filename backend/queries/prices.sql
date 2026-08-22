-- name: UpsertPrice :exec
INSERT INTO prices (asset_id, as_of, currency, value_usd, status, source, change_24h_pct)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (asset_id, as_of) DO UPDATE
SET value_usd = EXCLUDED.value_usd,
    status = EXCLUDED.status,
    source = EXCLUDED.source,
    change_24h_pct = EXCLUDED.change_24h_pct;

-- name: GetLatestPrices :many
SELECT DISTINCT ON (asset_id)
       asset_id, as_of, currency, value_usd, status, source, change_24h_pct, created_at
FROM prices
WHERE asset_id = ANY($1::uuid[])
ORDER BY asset_id, as_of DESC;

-- name: GetClosestPrice :one
SELECT asset_id, as_of, currency, value_usd, status, source, change_24h_pct, created_at
FROM prices
WHERE asset_id = $1
ORDER BY ABS(EXTRACT(EPOCH FROM (as_of - $2::timestamptz)))
LIMIT 1;
