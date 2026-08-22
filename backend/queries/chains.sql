-- Queries are written by hand and reviewed as SQL; no ORM generates them (§5.3).

-- name: ListChains :many
SELECT id, name, native_asset_id, address_format, is_supported, created_at, updated_at
FROM chains
ORDER BY name;

-- name: GetChain :one
SELECT id, name, native_asset_id, address_format, is_supported, created_at, updated_at
FROM chains
WHERE id = $1;

-- name: GetNativeAsset :one
SELECT a.id, a.chain_id, a.contract_address, a.symbol, a.name, a.decimals,
       a.asset_type, a.icon_url, a.market_data_provider, a.market_data_id,
       a.created_at, a.updated_at
FROM assets AS a
JOIN chains AS c ON c.native_asset_id = a.id
WHERE c.id = $1;
