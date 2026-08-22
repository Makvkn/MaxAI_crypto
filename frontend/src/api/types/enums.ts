/**
 * Backend state machines and enumerations, modelled as const objects plus
 * string-literal unions (the project compiles with `erasableSyntaxOnly`, so
 * TypeScript `enum` is unavailable — and unions map 1:1 onto OpenAPI enums).
 *
 * PROVISIONAL CONTRACT LAYER — see `src/api/types/index.ts`.
 */

/** Wallet lifecycle. `DELETED` is a soft delete on the domain level. */
export const WalletStatus = {
  ACTIVE: 'ACTIVE',
  SYNCING: 'SYNCING',
  ERROR: 'ERROR',
  PAUSED: 'PAUSED',
  DELETED: 'DELETED',
} as const
export type WalletStatus = (typeof WalletStatus)[keyof typeof WalletStatus]

/**
 * Wallet synchronisation state machine. Distinct from {@link WalletStatus}
 * and from {@link DataFreshness} — `STALE` is *not* a sync status.
 *
 * PENDING -> SYNCING -> READY | PARTIAL | FAILED
 */
export const SyncStatus = {
  PENDING: 'PENDING',
  SYNCING: 'SYNCING',
  READY: 'READY',
  PARTIAL: 'PARTIAL',
  FAILED: 'FAILED',
} as const
export type SyncStatus = (typeof SyncStatus)[keyof typeof SyncStatus]

/**
 * Stages the backend reports while `InitialWalletSyncJob` runs.
 *
 * The UI renders only stages the backend has actually reported. There is no
 * client-side timer advancing through these.
 */
export const SyncStage = {
  FETCHING_BALANCES: 'FETCHING_BALANCES',
  FETCHING_TRANSACTIONS: 'FETCHING_TRANSACTIONS',
  NORMALIZING_ASSETS: 'NORMALIZING_ASSETS',
  FETCHING_PRICES: 'FETCHING_PRICES',
  CALCULATING_PORTFOLIO: 'CALCULATING_PORTFOLIO',
  PREPARING_ANALYSIS: 'PREPARING_ANALYSIS',
} as const
export type SyncStage = (typeof SyncStage)[keyof typeof SyncStage]

/** Age of the underlying data. Thresholds are a backend configuration value. */
export const DataFreshness = {
  FRESH: 'FRESH',
  RECENT: 'RECENT',
  STALE: 'STALE',
  VERY_STALE: 'VERY_STALE',
} as const
export type DataFreshness = (typeof DataFreshness)[keyof typeof DataFreshness]

/**
 * How much the user is allowed to believe a figure. Not cosmetic: it changes
 * what the UI is permitted to display.
 */
export const DataQuality = {
  COMPLETE: 'COMPLETE',
  PARTIAL: 'PARTIAL',
  STALE: 'STALE',
  UNAVAILABLE: 'UNAVAILABLE',
} as const
export type DataQuality = (typeof DataQuality)[keyof typeof DataQuality]

/** Whether a valuation could be produced from priced positions. */
export const ValuationStatus = {
  COMPLETE: 'COMPLETE',
  PARTIAL: 'PARTIAL',
  UNAVAILABLE: 'UNAVAILABLE',
} as const
export type ValuationStatus =
  (typeof ValuationStatus)[keyof typeof ValuationStatus]

/** Whether a reliable market price exists for an asset. */
export const PriceStatus = {
  AVAILABLE: 'AVAILABLE',
  UNAVAILABLE: 'UNAVAILABLE',
} as const
export type PriceStatus = (typeof PriceStatus)[keyof typeof PriceStatus]

/** Asset visibility, decided by deterministic backend rules. */
export const AssetVisibility = {
  VISIBLE: 'VISIBLE',
  HIDDEN_DUST: 'HIDDEN_DUST',
  HIDDEN_SPAM: 'HIDDEN_SPAM',
  UNKNOWN: 'UNKNOWN',
} as const
export type AssetVisibility =
  (typeof AssetVisibility)[keyof typeof AssetVisibility]

export const AssetType = {
  NATIVE: 'NATIVE',
  TOKEN: 'TOKEN',
  UNKNOWN: 'UNKNOWN',
} as const
export type AssetType = (typeof AssetType)[keyof typeof AssetType]

/** Canonical transaction types. Classified by the backend, never inferred. */
export const TransactionType = {
  TRANSFER: 'TRANSFER',
  SWAP: 'SWAP',
  STAKE: 'STAKE',
  UNSTAKE: 'UNSTAKE',
  CLAIM: 'CLAIM',
  APPROVE: 'APPROVE',
  CONTRACT_INTERACTION: 'CONTRACT_INTERACTION',
  UNKNOWN: 'UNKNOWN',
} as const
export type TransactionType =
  (typeof TransactionType)[keyof typeof TransactionType]

export const TransactionStatus = {
  SUCCESS: 'SUCCESS',
  FAILED: 'FAILED',
  PENDING: 'PENDING',
} as const
export type TransactionStatus =
  (typeof TransactionStatus)[keyof typeof TransactionStatus]

/** Periods supported by snapshot-based portfolio performance. */
export const PerformancePeriod = {
  H24: '24h',
  D7: '7d',
  D30: '30d',
  ALL: 'all',
} as const
export type PerformancePeriod =
  (typeof PerformancePeriod)[keyof typeof PerformancePeriod]

/**
 * Availability of a performance figure. The product term is "Portfolio
 * Performance" — never "PnL", because the MVP has no accounting engine.
 */
export const PerformanceStatus = {
  AVAILABLE: 'AVAILABLE',
  PARTIAL: 'PARTIAL',
  UNAVAILABLE: 'UNAVAILABLE',
} as const
export type PerformanceStatus =
  (typeof PerformanceStatus)[keyof typeof PerformanceStatus]

/** AI routing intents resolved by the backend orchestrator. */
export const AIIntent = {
  PORTFOLIO_SUMMARY: 'PORTFOLIO_SUMMARY',
  PORTFOLIO_PERFORMANCE: 'PORTFOLIO_PERFORMANCE',
  PORTFOLIO_ALLOCATION: 'PORTFOLIO_ALLOCATION',
  TRANSACTION_EXPLANATION: 'TRANSACTION_EXPLANATION',
  SCENARIO_SIMULATION: 'SCENARIO_SIMULATION',
  GENERAL_PORTFOLIO_QUESTION: 'GENERAL_PORTFOLIO_QUESTION',
  UNSUPPORTED: 'UNSUPPORTED',
} as const
export type AIIntent = (typeof AIIntent)[keyof typeof AIIntent]

/** Domain tools the backend orchestrator may execute. Never run client-side. */
export const AIToolName = {
  GET_PORTFOLIO: 'get_portfolio',
  GET_POSITIONS: 'get_positions',
  GET_PORTFOLIO_PERFORMANCE: 'get_portfolio_performance',
  GET_TRANSACTION: 'get_transaction',
  GET_HISTORICAL_PORTFOLIO: 'get_historical_portfolio',
  GET_ASSET_PRICE: 'get_asset_price',
  SIMULATE_SCENARIO: 'simulate_scenario',
} as const
export type AIToolName = (typeof AIToolName)[keyof typeof AIToolName]

export const AIToolCallStatus = {
  RUNNING: 'RUNNING',
  COMPLETED: 'COMPLETED',
  FAILED: 'FAILED',
} as const
export type AIToolCallStatus =
  (typeof AIToolCallStatus)[keyof typeof AIToolCallStatus]

export const MessageRole = {
  USER: 'USER',
  ASSISTANT: 'ASSISTANT',
} as const
export type MessageRole = (typeof MessageRole)[keyof typeof MessageRole]

export const MessageStatus = {
  PENDING: 'PENDING',
  STREAMING: 'STREAMING',
  COMPLETED: 'COMPLETED',
  FAILED: 'FAILED',
} as const
export type MessageStatus = (typeof MessageStatus)[keyof typeof MessageStatus]

/** Kinds of backend evidence an AI claim can be anchored to. */
export const AIEvidenceType = {
  CALCULATION: 'calculation',
  PORTFOLIO: 'portfolio',
  PORTFOLIO_PERFORMANCE: 'portfolio_performance',
  PORTFOLIO_SNAPSHOT: 'portfolio_snapshot',
  POSITION: 'position',
  TRANSACTION: 'transaction',
  PRICE: 'price',
  SCENARIO: 'scenario',
} as const
export type AIEvidenceType =
  (typeof AIEvidenceType)[keyof typeof AIEvidenceType]

/** Entities an AI answer can point at, so answers can become interactive. */
export const AIReferenceType = {
  ASSET: 'asset',
  TRANSACTION: 'transaction',
  PORTFOLIO: 'portfolio',
  SCENARIO: 'scenario',
} as const
export type AIReferenceType =
  (typeof AIReferenceType)[keyof typeof AIReferenceType]

/** MVP scenario shapes handled by the backend `ScenarioService`. */
export const ScenarioType = {
  ASSET_PRICE_CHANGE: 'ASSET_PRICE_CHANGE',
} as const
export type ScenarioType = (typeof ScenarioType)[keyof typeof ScenarioType]

export const UserKind = {
  GUEST: 'GUEST',
  REGISTERED: 'REGISTERED',
} as const
export type UserKind = (typeof UserKind)[keyof typeof UserKind]

export const AuthMethod = {
  GUEST: 'GUEST',
  GOOGLE: 'GOOGLE',
  EMAIL: 'EMAIL',
} as const
export type AuthMethod = (typeof AuthMethod)[keyof typeof AuthMethod]

export const SubscriptionPlan = {
  FREE: 'FREE',
  PRO: 'PRO',
} as const
export type SubscriptionPlan =
  (typeof SubscriptionPlan)[keyof typeof SubscriptionPlan]

export const SubscriptionStatus = {
  ACTIVE: 'ACTIVE',
  CANCELED: 'CANCELED',
  EXPIRED: 'EXPIRED',
} as const
export type SubscriptionStatus =
  (typeof SubscriptionStatus)[keyof typeof SubscriptionStatus]

/** Severity of a data-quality notice attached to a payload. */
export const NoticeSeverity = {
  INFO: 'INFO',
  WARNING: 'WARNING',
} as const
export type NoticeSeverity = (typeof NoticeSeverity)[keyof typeof NoticeSeverity]

/**
 * Machine-readable notices the backend attaches to financial payloads. The
 * frontend maps these to user-facing copy so tone stays under product control.
 */
export const DataNoticeCode = {
  UNPRICED_ASSETS_EXCLUDED: 'UNPRICED_ASSETS_EXCLUDED',
  NFTS_EXCLUDED_FROM_VALUATION: 'NFTS_EXCLUDED_FROM_VALUATION',
  DEFI_POSITIONS_EXCLUDED: 'DEFI_POSITIONS_EXCLUDED',
  DATA_STALE: 'DATA_STALE',
  HISTORY_INCOMPLETE: 'HISTORY_INCOMPLETE',
  SYNC_PARTIALLY_FAILED: 'SYNC_PARTIALLY_FAILED',
} as const
export type DataNoticeCode =
  (typeof DataNoticeCode)[keyof typeof DataNoticeCode]
