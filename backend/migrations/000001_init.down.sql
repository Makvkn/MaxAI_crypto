BEGIN;

DROP TABLE IF EXISTS scenario_calculations;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS ai_usage_operations;
DROP TABLE IF EXISTS ai_usage;
DROP TABLE IF EXISTS conversation_messages;
DROP TABLE IF EXISTS conversations;
DROP TABLE IF EXISTS portfolio_snapshot_positions;
DROP TABLE IF EXISTS portfolio_snapshots;
DROP TABLE IF EXISTS prices;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS wallet_positions;
DROP TABLE IF EXISTS wallet_sync_runs;
DROP TABLE IF EXISTS wallet_sync_states;
DROP TABLE IF EXISTS wallets;

-- Break the chains/assets cycle before dropping either side.
ALTER TABLE chains DROP CONSTRAINT IF EXISTS chains_native_asset_id_fkey;
DROP TABLE IF EXISTS assets;
DROP TABLE IF EXISTS chains;

DROP TABLE IF EXISTS refresh_sessions;
DROP TABLE IF EXISTS auth_identities;
DROP TABLE IF EXISTS users;

COMMIT;
