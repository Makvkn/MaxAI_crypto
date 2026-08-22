/**
 * Snapshot-based portfolio performance.
 *
 * The product term is "Portfolio Performance". This is not PnL: there is no
 * cost basis, no realized/unrealized split and no tax lots in the MVP.
 *
 * PROVISIONAL CONTRACT LAYER — see `src/api/types/index.ts`.
 */
import type { Asset } from './asset'
import type { ApiErrorCodeValue } from './errors'
import type {
  DataQuality,
  PerformancePeriod,
  PerformanceStatus,
  ValuationStatus,
} from './enums'
import type { CurrencyCode, Decimal, Timestamp } from './primitives'

/** One point of the historical portfolio series, backed by a real snapshot. */
export interface PortfolioSnapshotPoint {
  captured_at: Timestamp
  /** `null` for a snapshot whose valuation was unavailable. */
  total_value_usd: Decimal | null
  status: ValuationStatus
}

/** The snapshot anchoring one end of the period. */
export interface PerformanceEndpoint {
  captured_at: Timestamp
  value_usd: Decimal
  status: ValuationStatus
}

/** Per-asset contribution to the period result, computed by the backend. */
export interface PerformanceDriver {
  asset: Asset
  allocation_pct: Decimal | null
  contribution_usd: Decimal | null
  contribution_pct: Decimal | null
  change_pct: Decimal | null
}

export interface PortfolioPerformance {
  wallet_id: string
  period: PerformancePeriod
  status: PerformanceStatus
  data_quality: DataQuality
  currency: CurrencyCode
  /** Snapshot closest to the start of the period; `null` when missing. */
  opening: PerformanceEndpoint | null
  closing: PerformanceEndpoint | null
  change_usd: Decimal | null
  change_pct: Decimal | null
  /** Historical series for the chart. Never interpolated by the frontend. */
  series: PortfolioSnapshotPoint[]
  drivers: PerformanceDriver[]
  /** Domain reason when `status` is `UNAVAILABLE`. */
  unavailable_reason: ApiErrorCodeValue | null
  calculation_id: string | null
  calculation_version: number | null
}
