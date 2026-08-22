import type {
  AIIntent,
  ChainId,
  DataQuality,
  PerformancePeriod,
  SyncStatus,
  TransactionType,
  ValuationStatus,
} from '@/api/types'

/**
 * One event schema for the whole product.
 *
 * Event names and payloads are declared here so no component invents its own
 * naming, and so the funnel from the product spec is measurable end to end:
 *
 *   landing_visit -> wallet_entered -> analysis_started -> portfolio_loaded
 *   -> ai_insight_viewed -> first_ai_question -> second_ai_question
 *   -> scenario_used -> return_visit
 */
export interface AnalyticsEventMap {
  landing_visit: { referrer?: string }
  return_visit: { wallet_id?: string }

  chain_selected: { chain_id: ChainId }
  wallet_entered: { chain_id: ChainId }
  analysis_started: { chain_id: ChainId }
  analysis_failed: { chain_id: ChainId; error_code: string }

  sync_completed: { wallet_id: string; sync_status: SyncStatus }

  portfolio_loaded: {
    wallet_id: string
    data_quality: DataQuality
    valuation_status: ValuationStatus
  }
  performance_period_changed: { wallet_id: string; period: PerformancePeriod }
  hidden_assets_expanded: { wallet_id: string; count: number }

  ai_insight_viewed: { wallet_id: string; intent: AIIntent }
  ai_question_asked: { wallet_id: string; question_index: number }
  first_ai_question: { wallet_id: string }
  second_ai_question: { wallet_id: string }
  ai_answer_unsupported: { wallet_id: string; reason: string | null }
  ai_limit_reached: { limit: number }

  transaction_opened: { wallet_id: string; transaction_type: TransactionType }
  transaction_explained: { wallet_id: string; transaction_type: TransactionType }
  transactions_page_loaded: { wallet_id: string; page_index: number }

  scenario_used: { wallet_id: string; asset_symbol: string; change_pct: string }

  account_upgraded: { method: string }
  signed_in: { method: string }
}

export type AnalyticsEventName = keyof AnalyticsEventMap
