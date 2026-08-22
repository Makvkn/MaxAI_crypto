-- name: GetAssetByID :one
SELECT id, chain_id, contract_address, symbol, name, decimals, asset_type,
       icon_url, market_data_provider, market_data_id, created_at, updated_at
FROM assets
WHERE id = $1;

-- name: GetAssetsByIDs :many
SELECT id, chain_id, contract_address, symbol, name, decimals, asset_type,
       icon_url, market_data_provider, market_data_id, created_at, updated_at
FROM assets
WHERE id = ANY($1::uuid[]);

-- name: FindAssetByIdentity :one
SELECT id, chain_id, contract_address, symbol, name, decimals, asset_type,
       icon_url, market_data_provider, market_data_id, created_at, updated_at
FROM assets
WHERE chain_id = $1
  AND contract_address IS NOT DISTINCT FROM $2;

-- name: UpsertAsset :one
INSERT INTO assets (id, chain_id, contract_address, symbol, name, decimals, asset_type,
                    icon_url, market_data_provider, market_data_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT ON CONSTRAINT assets_identity_key DO UPDATE
SET symbol = EXCLUDED.symbol,
    name = EXCLUDED.name,
    decimals = EXCLUDED.decimals,
    asset_type = EXCLUDED.asset_type,
    icon_url = COALESCE(EXCLUDED.icon_url, assets.icon_url),
    updated_at = NOW()
RETURNING id, chain_id, contract_address, symbol, name, decimals, asset_type,
          icon_url, market_data_provider, market_data_id, created_at, updated_at;

-- name: SetAssetMarketDataMapping :exec
UPDATE assets
SET market_data_provider = $2,
    market_data_id = $3,
    updated_at = NOW()
WHERE id = $1;

-- name: ListUnmappedAssets :many
SELECT id, chain_id, contract_address, symbol, name, decimals, asset_type,
       icon_url, market_data_provider, market_data_id, created_at, updated_at
FROM assets
WHERE market_data_provider IS NULL
   OR market_data_id IS NULL
ORDER BY created_at
LIMIT $1;
