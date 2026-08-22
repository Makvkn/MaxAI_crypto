/**
 * Portfolio and positions.
 *
 * Every figure here is computed by the backend. `null` means "unknown" and
 * must never be rendered as `$0`.
 *
 * PROVISIONAL CONTRACT LAYER — see `src/api/types/index.ts`.
 */
import type { Asset, Price } from './asset'
import type {
  AssetVisibility,
  DataFreshness,
  DataNoticeCode,
  DataQuality,
  NoticeSeverity,
  ValuationStatus,
} from './enums'
import type { CurrencyCode, Decimal, Timestamp } from './primitives'

/**
 * Machine-readable data-quality notice. The frontend owns the wording; the
 * backend owns the facts.
 */
export interface DataNotice {
  code: DataNoticeCode
  severity: NoticeSeverity
  /** Substitution values for the frontend copy, e.g. `{ count: 1 }`. */
  params?: Record<string, string | number>
}

/** Current holding of one asset in one wallet, with backend valuation. */
export interface WalletPosition {
  asset: Asset
  /** Balance scaled by `asset.decimals`. Always displayable. */
  balance: Decimal
  /** Unscaled on-chain balance, kept for traceability. */
  balance_raw: string
  /** `null` when no reliable market price exists. */
  price: Price | null
  value_usd: Decimal | null
  allocation_pct: Decimal | null
  change_24h_pct: Decimal | null
  change_24h_usd: Decimal | null
  visibility: AssetVisibility
  valuation_status: ValuationStatus
  updated_at: Timestamp
}

/** What the backend left out of the valuation, and why. */
export interface PortfolioExclusions {
  /** Positions excluded because their price is unknown. */
  unpriced_positions: number
  nfts_excluded: boolean
  defi_positions_excluded: boolean
}

export interface Portfolio {
  wallet_id: string
  currency: CurrencyCode
  /** `null` when valuation_status is `UNAVAILABLE`. */
  total_value_usd: Decimal | null
  valuation_status: ValuationStatus
  data_quality: DataQuality
  data_freshness: DataFreshness
  change_24h_usd: Decimal | null
  change_24h_pct: Decimal | null
  /** When this valuation was produced. */
  as_of: Timestamp
  /** When the wallet was last synchronised against its chain. */
  last_synced_at: Timestamp | null
  calculation_version: number
  positions: WalletPosition[]
  visible_positions_count: number
  hidden_positions_count: number
  exclusions: PortfolioExclusions
  notices: DataNotice[]
}
