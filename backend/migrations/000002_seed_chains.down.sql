BEGIN;

UPDATE chains SET native_asset_id = NULL;

DELETE FROM assets WHERE asset_type = 'NATIVE' AND chain_id IN (
    'ethereum', 'bitcoin', 'bnb', 'solana', 'litecoin', 'xrpl', 'tron', 'dogecoin'
);

DELETE FROM chains WHERE id IN (
    'ethereum', 'bitcoin', 'bnb', 'solana', 'litecoin', 'xrpl', 'tron', 'dogecoin'
);

COMMIT;
