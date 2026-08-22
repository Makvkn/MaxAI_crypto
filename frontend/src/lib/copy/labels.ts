import {
  AssetVisibility,
  DataFreshness,
  DataQuality,
  PerformancePeriod,
  SyncStage,
  SyncStatus,
  TransactionStatus,
  TransactionType,
  ValuationStatus,
  type AIToolName,
} from '@/api/types'

/**
 * Domain enum -> user-facing label.
 *
 * All copy for backend states lives here so no JSX contains a raw enum string
 * and wording stays consistent across the product.
 */

export const transactionTypeLabels: Record<TransactionType, string> = {
  [TransactionType.TRANSFER]: 'Transfer',
  [TransactionType.SWAP]: 'Swap',
  [TransactionType.STAKE]: 'Stake',
  [TransactionType.UNSTAKE]: 'Unstake',
  [TransactionType.CLAIM]: 'Claim',
  [TransactionType.APPROVE]: 'Approval',
  [TransactionType.CONTRACT_INTERACTION]: 'Contract interaction',
  // Never relabelled as something more specific.
  [TransactionType.UNKNOWN]: 'Unknown',
}

export const transactionStatusLabels: Record<TransactionStatus, string> = {
  [TransactionStatus.SUCCESS]: 'Confirmed',
  [TransactionStatus.FAILED]: 'Failed',
  [TransactionStatus.PENDING]: 'Pending',
}

export const freshnessLabels: Record<DataFreshness, string> = {
  [DataFreshness.FRESH]: 'Live',
  [DataFreshness.RECENT]: 'Recent',
  [DataFreshness.STALE]: 'Stale',
  [DataFreshness.VERY_STALE]: 'Very stale',
}

export const dataQualityLabels: Record<DataQuality, string> = {
  [DataQuality.COMPLETE]: 'Complete',
  [DataQuality.PARTIAL]: 'Partial',
  [DataQuality.STALE]: 'Stale',
  [DataQuality.UNAVAILABLE]: 'Unavailable',
}

export const valuationStatusLabels: Record<ValuationStatus, string> = {
  [ValuationStatus.COMPLETE]: 'Fully valued',
  [ValuationStatus.PARTIAL]: 'Partially valued',
  [ValuationStatus.UNAVAILABLE]: 'Not valued',
}

export const visibilityLabels: Record<AssetVisibility, string> = {
  [AssetVisibility.VISIBLE]: 'Visible',
  [AssetVisibility.HIDDEN_DUST]: 'Dust',
  [AssetVisibility.HIDDEN_SPAM]: 'Spam',
  [AssetVisibility.UNKNOWN]: 'Unrecognised',
}

export const syncStatusLabels: Record<SyncStatus, string> = {
  [SyncStatus.PENDING]: 'Queued',
  [SyncStatus.SYNCING]: 'Analysing',
  [SyncStatus.READY]: 'Ready',
  [SyncStatus.PARTIAL]: 'Partially ready',
  [SyncStatus.FAILED]: 'Failed',
}

/** Only shown for stages the backend actually reports. */
export const syncStageLabels: Record<SyncStage, string> = {
  [SyncStage.FETCHING_BALANCES]: 'Fetching balances',
  [SyncStage.FETCHING_TRANSACTIONS]: 'Fetching transactions',
  [SyncStage.NORMALIZING_ASSETS]: 'Normalising assets',
  [SyncStage.FETCHING_PRICES]: 'Fetching market prices',
  [SyncStage.CALCULATING_PORTFOLIO]: 'Calculating portfolio',
  [SyncStage.PREPARING_ANALYSIS]: 'Preparing analysis',
}

export const periodLabels: Record<PerformancePeriod, string> = {
  [PerformancePeriod.H24]: '24h',
  [PerformancePeriod.D7]: '7d',
  [PerformancePeriod.D30]: '30d',
  [PerformancePeriod.ALL]: 'All time',
}

export const periodDescriptions: Record<PerformancePeriod, string> = {
  [PerformancePeriod.H24]: 'the last 24 hours',
  [PerformancePeriod.D7]: 'the last 7 days',
  [PerformancePeriod.D30]: 'the last 30 days',
  [PerformancePeriod.ALL]: 'all recorded history',
}

/** What the AI is doing while a backend tool runs. */
export const aiToolLabels: Record<AIToolName | string, string> = {
  get_portfolio: 'Reading portfolio',
  get_positions: 'Reading positions',
  get_portfolio_performance: 'Measuring performance',
  get_transaction: 'Reading transaction',
  get_historical_portfolio: 'Reading portfolio history',
  get_asset_price: 'Checking market price',
  simulate_scenario: 'Running scenario',
}

export function aiToolLabel(tool: string): string {
  return aiToolLabels[tool] ?? 'Working'
}

/** What a claim is anchored to. Keeps evidence readable if it becomes clickable. */
export const evidenceTypeLabels: Record<string, string> = {
  calculation: 'Calculation',
  portfolio: 'Portfolio',
  portfolio_performance: 'Performance',
  portfolio_snapshot: 'Snapshot',
  position: 'Position',
  transaction: 'Transaction',
  price: 'Price',
  scenario: 'Scenario',
}

export function evidenceTypeLabel(type: string): string {
  return evidenceTypeLabels[type] ?? 'Backend fact'
}
