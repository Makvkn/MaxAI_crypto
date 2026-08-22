-- MaxAI Crypto initial schema.
--
-- Conventions enforced throughout (backend spec §111, §113, §186):
--   * every financial value is NUMERIC; FLOAT/REAL/DOUBLE PRECISION are never
--     used for canonical financial data;
--   * NULL in a financial column means "unknown", which is different from 0;
--   * every timestamp is TIMESTAMPTZ and stored in UTC;
--   * enums are TEXT with CHECK constraints so new values do not require a
--     type rewrite;
--   * domain integrity is expressed as constraints, not only in Go.

BEGIN;

-- ---------------------------------------------------------------- chains

CREATE TABLE chains (
    id              TEXT PRIMARY KEY,
    name            TEXT        NOT NULL,
    -- Set once the chain's native asset row exists; the foreign key is added
    -- after the assets table is created.
    native_asset_id UUID,
    address_format  TEXT        NOT NULL
        CHECK (address_format IN ('EVM', 'BITCOIN_LIKE', 'SOLANA', 'TRON', 'XRPL')),
    is_supported    BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ----------------------------------------------------------------- users

CREATE TABLE users (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kind         TEXT        NOT NULL CHECK (kind IN ('GUEST', 'REGISTERED')),
    email        TEXT,
    display_name TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ
);

-- Email uniqueness is case-insensitive and ignores deleted accounts.
CREATE UNIQUE INDEX users_email_key
    ON users (LOWER(email))
    WHERE email IS NOT NULL AND deleted_at IS NULL;

-- A user record is never coupled to a single authentication provider (§11).
CREATE TABLE auth_identities (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    provider       TEXT        NOT NULL CHECK (provider IN ('guest', 'google', 'email')),
    subject        TEXT        NOT NULL,
    email          TEXT,
    -- Only ever a hash produced by a modern password algorithm (§14).
    password_hash  TEXT,
    email_verified BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT auth_identities_provider_subject_key UNIQUE (provider, subject),
    -- A password may only exist on the email provider.
    CONSTRAINT auth_identities_password_only_for_email
        CHECK (password_hash IS NULL OR provider = 'email')
);

CREATE INDEX auth_identities_user_id_idx ON auth_identities (user_id);

-- Refresh sessions are server-side state; only the hash of the secret is
-- persisted, and rotation is recorded so reuse is detectable (§13).
CREATE TABLE refresh_sessions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash   TEXT        NOT NULL UNIQUE,
    issued_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at   TIMESTAMPTZ NOT NULL,
    revoked_at   TIMESTAMPTZ,
    rotated_to   UUID REFERENCES refresh_sessions (id) ON DELETE SET NULL,
    user_agent   TEXT,
    ip_address   INET,
    last_used_at TIMESTAMPTZ,

    CONSTRAINT refresh_sessions_expiry_after_issue CHECK (expires_at > issued_at)
);

CREATE INDEX refresh_sessions_user_id_idx ON refresh_sessions (user_id)
    WHERE revoked_at IS NULL;

-- ---------------------------------------------------------------- assets

CREATE TABLE assets (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chain_id             TEXT        NOT NULL REFERENCES chains (id),
    -- NULL for the chain's native asset, a normalized contract address for
    -- tokens (§31).
    contract_address     TEXT,
    symbol               TEXT        NOT NULL,
    name                 TEXT        NOT NULL,
    decimals             INTEGER     NOT NULL CHECK (decimals >= 0 AND decimals <= 38),
    asset_type           TEXT        NOT NULL CHECK (asset_type IN ('NATIVE', 'TOKEN', 'UNKNOWN')),
    icon_url             TEXT,
    -- Market-data mapping. Both columns are NULL when no reliable mapping
    -- exists, which makes the price unknown rather than zero (§33, §40).
    market_data_provider TEXT CHECK (market_data_provider IN ('coingecko')),
    market_data_id       TEXT,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Asset identity is chain plus contract address. NULLS NOT DISTINCT makes
    -- the native asset of a chain unique too, so a symbol can never be the
    -- identity (§31).
    CONSTRAINT assets_identity_key UNIQUE NULLS NOT DISTINCT (chain_id, contract_address),
    CONSTRAINT assets_native_has_no_contract
        CHECK ((asset_type = 'NATIVE') = (contract_address IS NULL)),
    CONSTRAINT assets_market_mapping_is_complete
        CHECK ((market_data_provider IS NULL) = (market_data_id IS NULL))
);

CREATE INDEX assets_unmapped_idx ON assets (chain_id)
    WHERE market_data_id IS NULL;

ALTER TABLE chains
    ADD CONSTRAINT chains_native_asset_id_fkey
    FOREIGN KEY (native_asset_id) REFERENCES assets (id);

-- --------------------------------------------------------------- wallets

CREATE TABLE wallets (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    chain_id   TEXT        NOT NULL REFERENCES chains (id),
    -- Canonical, chain-normalized address. Only public addresses are ever
    -- stored; keys and seed phrases are never collected (§2, §188).
    address    TEXT        NOT NULL,
    label      TEXT,
    status     TEXT        NOT NULL
        CHECK (status IN ('ACTIVE', 'SYNCING', 'ERROR', 'PAUSED', 'DELETED')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT wallets_deleted_status_agree
        CHECK ((deleted_at IS NULL) OR status = 'DELETED')
);

-- The same address may be re-added after deletion, but never duplicated while
-- active.
CREATE UNIQUE INDEX wallets_active_identity_key
    ON wallets (user_id, chain_id, address)
    WHERE deleted_at IS NULL;

CREATE INDEX wallets_user_id_idx ON wallets (user_id, deleted_at);
CREATE INDEX wallets_chain_address_idx ON wallets (chain_id, address);

-- Synchronization state is separate from the wallet lifecycle (§18).
CREATE TABLE wallet_sync_states (
    wallet_id        UUID PRIMARY KEY REFERENCES wallets (id) ON DELETE CASCADE,
    status           TEXT        NOT NULL
        CHECK (status IN ('PENDING', 'SYNCING', 'READY', 'PARTIAL', 'FAILED')),
    -- Only stages the backend has actually reached are recorded; fabricated
    -- progress is forbidden (§19).
    stage            TEXT CHECK (stage IN (
        'FETCHING_BALANCES', 'FETCHING_TRANSACTIONS', 'NORMALIZING_ASSETS',
        'FETCHING_PRICES', 'CALCULATING_PORTFOLIO', 'PREPARING_ANALYSIS')),
    stages_completed TEXT[]      NOT NULL DEFAULT '{}',
    started_at       TIMESTAMPTZ,
    completed_at     TIMESTAMPTZ,
    last_synced_at   TIMESTAMPTZ,
    -- Domain-level failure reason; provider errors are mapped before they get
    -- here (§28).
    error_code       TEXT,
    error_message    TEXT,
    sync_job_id      TEXT,
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Scheduler lookup: wallets whose last successful sync is oldest first.
CREATE INDEX wallet_sync_states_due_idx
    ON wallet_sync_states (last_synced_at NULLS FIRST)
    WHERE status <> 'SYNCING';

-- One row per synchronization attempt, for observability (§122).
CREATE TABLE wallet_sync_runs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id   UUID        NOT NULL REFERENCES wallets (id) ON DELETE CASCADE,
    -- Deterministic per attempt, which is what makes a retried job resume the
    -- same run instead of opening a second one (§60).
    job_id      TEXT        NOT NULL UNIQUE,
    trigger     TEXT        NOT NULL CHECK (trigger IN ('INITIAL', 'SCHEDULED', 'MANUAL')),
    provider    TEXT,
    status      TEXT        NOT NULL
        CHECK (status IN ('PENDING', 'SYNCING', 'READY', 'PARTIAL', 'FAILED')),
    started_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMPTZ,
    error_code  TEXT,
    error_text  TEXT
);

CREATE INDEX wallet_sync_runs_wallet_idx ON wallet_sync_runs (wallet_id, started_at DESC);

-- ------------------------------------------------------------- positions

CREATE TABLE wallet_positions (
    wallet_id          UUID        NOT NULL REFERENCES wallets (id) ON DELETE CASCADE,
    asset_id           UUID        NOT NULL REFERENCES assets (id),
    -- On-chain integer amount in the asset's smallest unit.
    balance_raw        NUMERIC(78, 0) NOT NULL,
    -- balance_raw scaled by the asset's decimals.
    balance_normalized NUMERIC     NOT NULL,
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (wallet_id, asset_id),
    CONSTRAINT wallet_positions_balance_not_negative CHECK (balance_normalized >= 0)
);

-- ---------------------------------------------------------- transactions

CREATE TABLE transactions (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id     UUID        NOT NULL REFERENCES wallets (id) ON DELETE CASCADE,
    chain_id      TEXT        NOT NULL REFERENCES chains (id),

    tx_hash       TEXT        NOT NULL,
    -- Disambiguates several wallet-relevant movements inside one hash (§48).
    log_index     INTEGER     NOT NULL DEFAULT 0,
    block_number  BIGINT,
    timestamp     TIMESTAMPTZ NOT NULL,

    status        TEXT        NOT NULL CHECK (status IN ('SUCCESS', 'FAILED', 'PENDING')),
    -- Backend-owned classification. An unclassifiable transaction stays
    -- UNKNOWN; the LLM may never promote it (§46, §47).
    type          TEXT        NOT NULL CHECK (type IN (
        'TRANSFER', 'SWAP', 'STAKE', 'UNSTAKE', 'CLAIM',
        'APPROVE', 'CONTRACT_INTERACTION', 'UNKNOWN')),

    from_address  TEXT,
    to_address    TEXT,

    asset_in_id   UUID REFERENCES assets (id),
    amount_in     NUMERIC,
    asset_out_id  UUID REFERENCES assets (id),
    amount_out    NUMERIC,
    fee_asset_id  UUID REFERENCES assets (id),
    fee_amount    NUMERIC,

    protocol      TEXT,
    counterparty  TEXT,
    -- A reference to the provider record, not a stored provider payload (§163).
    raw_reference TEXT,

    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Retrying a sync must never create duplicates (§48).
    CONSTRAINT transactions_identity_key UNIQUE (wallet_id, chain_id, tx_hash, log_index),
    CONSTRAINT transactions_amount_in_needs_asset
        CHECK ((amount_in IS NULL) = (asset_in_id IS NULL)),
    CONSTRAINT transactions_amount_out_needs_asset
        CHECK ((amount_out IS NULL) = (asset_out_id IS NULL)),
    CONSTRAINT transactions_fee_needs_asset
        CHECK ((fee_amount IS NULL) = (fee_asset_id IS NULL))
);

-- Matches the stable cursor ordering `timestamp DESC, id DESC` (§109).
CREATE INDEX transactions_wallet_timeline_idx
    ON transactions (wallet_id, timestamp DESC, id DESC);
CREATE INDEX transactions_wallet_type_timeline_idx
    ON transactions (wallet_id, type, timestamp DESC, id DESC);
CREATE INDEX transactions_chain_hash_idx ON transactions (chain_id, tx_hash);

-- ---------------------------------------------------------------- prices

CREATE TABLE prices (
    asset_id       UUID        NOT NULL REFERENCES assets (id) ON DELETE CASCADE,
    as_of          TIMESTAMPTZ NOT NULL,
    currency       TEXT        NOT NULL DEFAULT 'USD' CHECK (currency = 'USD'),
    -- NULL means the price is unknown at this instant, never zero (§40).
    value_usd      NUMERIC,
    status         TEXT        NOT NULL CHECK (status IN ('AVAILABLE', 'UNAVAILABLE')),
    source         TEXT        NOT NULL CHECK (source IN ('coingecko')),
    change_24h_pct NUMERIC,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (asset_id, as_of),
    -- An AVAILABLE price must actually carry a value.
    CONSTRAINT prices_available_has_value
        CHECK ((status = 'AVAILABLE') = (value_usd IS NOT NULL))
);

CREATE INDEX prices_asset_recent_idx ON prices (asset_id, as_of DESC);

-- ------------------------------------------------------------- snapshots

CREATE TABLE portfolio_snapshots (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id           UUID        NOT NULL REFERENCES wallets (id) ON DELETE CASCADE,
    captured_at         TIMESTAMPTZ NOT NULL,
    -- NULL when no valuation could be produced at all (§41).
    total_value_usd     NUMERIC,
    status              TEXT        NOT NULL
        CHECK (status IN ('COMPLETE', 'PARTIAL', 'UNAVAILABLE')),
    data_quality        TEXT        NOT NULL
        CHECK (data_quality IN ('COMPLETE', 'PARTIAL', 'STALE', 'UNAVAILABLE')),
    -- Records which algorithm produced this snapshot, so results stay
    -- reproducible after the logic changes (§51).
    calculation_version INTEGER     NOT NULL,
    -- Links the snapshot to the sync attempt that produced it. The unique
    -- index below is what stops a retried job from writing a second snapshot
    -- for the same run (§60).
    sync_run_id         UUID REFERENCES wallet_sync_runs (id) ON DELETE SET NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT portfolio_snapshots_value_matches_status
        CHECK ((status = 'UNAVAILABLE') OR (total_value_usd IS NOT NULL))
);

CREATE INDEX portfolio_snapshots_wallet_history_idx
    ON portfolio_snapshots (wallet_id, captured_at DESC);
CREATE UNIQUE INDEX portfolio_snapshots_sync_run_key
    ON portfolio_snapshots (sync_run_id)
    WHERE sync_run_id IS NOT NULL;

CREATE TABLE portfolio_snapshot_positions (
    snapshot_id     UUID        NOT NULL REFERENCES portfolio_snapshots (id) ON DELETE CASCADE,
    asset_id        UUID        NOT NULL REFERENCES assets (id),
    balance         NUMERIC     NOT NULL,
    -- NULL price means the position was held but could not be valued; its
    -- balance is still part of the historical record (§39, §40).
    price_usd       NUMERIC,
    value_usd       NUMERIC,
    allocation_pct  NUMERIC,
    -- Captured so a past valuation stays auditable (§161, §162).
    price_timestamp TIMESTAMPTZ,
    price_source    TEXT CHECK (price_source IN ('coingecko')),

    PRIMARY KEY (snapshot_id, asset_id),
    CONSTRAINT snapshot_positions_value_needs_price
        CHECK ((value_usd IS NULL) = (price_usd IS NULL))
);

-- --------------------------------------------------------- conversations

CREATE TABLE conversations (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id              UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    -- MVP conversations are scoped to the wallet being analysed (§182).
    wallet_id            UUID        NOT NULL REFERENCES wallets (id) ON DELETE CASCADE,
    title                TEXT        NOT NULL,
    message_count        INTEGER     NOT NULL DEFAULT 0,
    last_message_preview TEXT,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX conversations_user_recent_idx ON conversations (user_id, updated_at DESC, id DESC);
CREATE INDEX conversations_wallet_idx ON conversations (wallet_id);

CREATE TABLE conversation_messages (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID        NOT NULL REFERENCES conversations (id) ON DELETE CASCADE,
    role            TEXT        NOT NULL CHECK (role IN ('USER', 'ASSISTANT')),
    status          TEXT        NOT NULL
        CHECK (status IN ('PENDING', 'STREAMING', 'COMPLETED', 'FAILED')),
    content         TEXT        NOT NULL DEFAULT '',
    -- The validated structured AI response (§74). Stored as JSONB because its
    -- shape is owned by the API contract, not by the relational schema.
    response        JSONB,
    tool_calls      JSONB       NOT NULL DEFAULT '[]'::JSONB,
    error_code      TEXT,
    error_message   TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT conversation_messages_response_only_for_assistant
        CHECK (response IS NULL OR role = 'ASSISTANT')
);

CREATE INDEX conversation_messages_timeline_idx
    ON conversation_messages (conversation_id, created_at DESC, id DESC);

-- --------------------------------------------------------------- ai usage

CREATE TABLE ai_usage (
    user_id    UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    -- UTC day boundary (§86).
    usage_date DATE        NOT NULL,
    used       INTEGER     NOT NULL DEFAULT 0 CHECK (used >= 0),
    plan       TEXT        NOT NULL CHECK (plan IN ('FREE', 'PRO')),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (user_id, usage_date)
);

-- One row per settled AI operation. The unique idempotency key is what stops a
-- retried request from consuming a second unit (§60, §87).
CREATE TABLE ai_usage_operations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    usage_date      DATE        NOT NULL,
    operation       TEXT        NOT NULL CHECK (operation IN (
        'AI_INSIGHT', 'ASK_AI', 'TRANSACTION_EXPLANATION', 'SCENARIO_SIMULATION')),
    idempotency_key TEXT        NOT NULL UNIQUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ai_usage_operations_user_day_idx ON ai_usage_operations (user_id, usage_date);

-- ---------------------------------------------------------- subscriptions

CREATE TABLE subscriptions (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id            UUID        NOT NULL UNIQUE REFERENCES users (id) ON DELETE CASCADE,
    plan               TEXT        NOT NULL CHECK (plan IN ('FREE', 'PRO')),
    status             TEXT        NOT NULL CHECK (status IN ('ACTIVE', 'CANCELED', 'EXPIRED')),
    current_period_end TIMESTAMPTZ,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ------------------------------------------------------------- scenarios

-- Deterministic scenario results are persisted so an AI claim can cite the
-- exact calculation that produced it (§51, §73, §85).
CREATE TABLE scenario_calculations (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    wallet_id           UUID        NOT NULL REFERENCES wallets (id) ON DELETE CASCADE,
    type                TEXT        NOT NULL CHECK (type IN ('ASSET_PRICE_CHANGE')),
    asset_id            UUID        NOT NULL REFERENCES assets (id),
    change_pct          NUMERIC     NOT NULL,

    baseline_portfolio_value_usd  NUMERIC,
    baseline_asset_value_usd      NUMERIC,
    baseline_asset_allocation_pct NUMERIC,

    projected_portfolio_value_usd NUMERIC,
    projected_asset_value_usd     NUMERIC,
    asset_impact_usd              NUMERIC,
    portfolio_change_usd          NUMERIC,
    portfolio_change_pct          NUMERIC,

    data_quality        TEXT        NOT NULL
        CHECK (data_quality IN ('COMPLETE', 'PARTIAL', 'STALE', 'UNAVAILABLE')),
    calculation_version INTEGER     NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX scenario_calculations_wallet_idx ON scenario_calculations (wallet_id, created_at DESC);

COMMIT;
