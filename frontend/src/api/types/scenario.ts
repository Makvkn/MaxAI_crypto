/**
 * Scenario simulation.
 *
 * `ScenarioService` performs the deterministic calculation; the LLM only
 * explains the returned numbers. The frontend does neither.
 *
 * PROVISIONAL CONTRACT LAYER — see `src/api/types/index.ts`.
 */
import type { AIResponse } from './ai'
import type { Asset } from './asset'
import type { DataQuality, ScenarioType } from './enums'
import type { CurrencyCode, Decimal, Timestamp } from './primitives'

export interface ScenarioRequest {
  wallet_id: string
  type: ScenarioType
  asset_id: string
  /** Signed percentage change, e.g. `"-20"` for "ETH falls 20%". */
  change_pct: Decimal
}

/** Portfolio state before the scenario is applied. */
export interface ScenarioBaseline {
  portfolio_value_usd: Decimal | null
  asset_value_usd: Decimal | null
  asset_allocation_pct: Decimal | null
}

/** Portfolio state the backend projects for the scenario. */
export interface ScenarioProjection {
  portfolio_value_usd: Decimal | null
  asset_value_usd: Decimal | null
  asset_impact_usd: Decimal | null
  portfolio_change_usd: Decimal | null
  portfolio_change_pct: Decimal | null
}

export interface ScenarioResult {
  id: string
  wallet_id: string
  type: ScenarioType
  currency: CurrencyCode
  asset: Asset
  change_pct: Decimal
  baseline: ScenarioBaseline
  projection: ScenarioProjection
  data_quality: DataQuality
  calculation_id: string
  calculation_version: number
  created_at: Timestamp
  /** AI interpretation of the deterministic result. */
  explanation: AIResponse | null
}
