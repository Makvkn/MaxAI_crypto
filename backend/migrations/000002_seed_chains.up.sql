-- Reference data for the eight MVP chains and their native assets (§20).
--
-- TRON is the blockchain; TRX is its native asset.
--
-- Chain codes are the stable machine identifiers the API accepts. The
-- divergence from the §21 examples for bnb and xrpl is recorded in
-- openapi/DECISIONS.md.
--
-- Native asset UUIDs are fixed literals so that every environment resolves the
-- same identity and reruns stay idempotent.

BEGIN;

INSERT INTO chains (id, name, address_format) VALUES
    ('ethereum', 'Ethereum',    'EVM'),
    ('bitcoin',  'Bitcoin',     'BITCOIN_LIKE'),
    ('bnb',      'BNB Chain',   'EVM'),
    ('solana',   'Solana',      'SOLANA'),
    ('litecoin', 'Litecoin',    'BITCOIN_LIKE'),
    ('xrpl',     'XRP Ledger',  'XRPL'),
    ('tron',     'TRON',        'TRON'),
    ('dogecoin', 'Dogecoin',    'BITCOIN_LIKE')
ON CONFLICT (id) DO NOTHING;

INSERT INTO assets (id, chain_id, contract_address, symbol, name, decimals, asset_type,
                    market_data_provider, market_data_id) VALUES
    ('00000000-0000-4000-a000-000000000001', 'ethereum', NULL, 'ETH',  'Ether',     18, 'NATIVE', 'coingecko', 'ethereum'),
    ('00000000-0000-4000-a000-000000000002', 'bitcoin',  NULL, 'BTC',  'Bitcoin',    8, 'NATIVE', 'coingecko', 'bitcoin'),
    ('00000000-0000-4000-a000-000000000003', 'bnb',      NULL, 'BNB',  'BNB',       18, 'NATIVE', 'coingecko', 'binancecoin'),
    ('00000000-0000-4000-a000-000000000004', 'solana',   NULL, 'SOL',  'Solana',     9, 'NATIVE', 'coingecko', 'solana'),
    ('00000000-0000-4000-a000-000000000005', 'litecoin', NULL, 'LTC',  'Litecoin',   8, 'NATIVE', 'coingecko', 'litecoin'),
    ('00000000-0000-4000-a000-000000000006', 'xrpl',     NULL, 'XRP',  'XRP',        6, 'NATIVE', 'coingecko', 'ripple'),
    ('00000000-0000-4000-a000-000000000007', 'tron',     NULL, 'TRX',  'TRON',       6, 'NATIVE', 'coingecko', 'tron'),
    ('00000000-0000-4000-a000-000000000008', 'dogecoin', NULL, 'DOGE', 'Dogecoin',   8, 'NATIVE', 'coingecko', 'dogecoin')
ON CONFLICT ON CONSTRAINT assets_identity_key DO NOTHING;

UPDATE chains AS c
SET native_asset_id = a.id
FROM assets AS a
WHERE a.chain_id = c.id
  AND a.asset_type = 'NATIVE'
  AND c.native_asset_id IS NULL;

COMMIT;
