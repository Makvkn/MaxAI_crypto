/**
 * Assets and prices as normalised domain data.
 *
 * An asset is never identified by symbol alone — `chain_id` plus
 * `contract_address` (NULL for native assets) is the identity. Market-data
 * provider identifiers are deliberately absent from the DTO: which provider
 * priced an asset is a backend concern.
 *
 * PROVISIONAL CONTRACT LAYER — see `src/api/types/index.ts`.
 */
import type { ChainId } from './chain'
import type { AssetType, DataFreshness, PriceStatus } from './enums'
import type { CurrencyCode, Decimal, Timestamp } from './primitives'

export interface Asset {
  id: string
  chain_id: ChainId
  /** NULL for a chain's native asset. */
  contract_address: string | null
  symbol: string
  name: string
  decimals: number
  asset_type: AssetType
  icon_url: string | null
  /**
   * Whether the backend holds a reliable market-data mapping for this asset.
   * `false` means the price is unknown — not zero.
   */
  has_market_data: boolean
}

/**
 * A price is never a timeless number: it always carries when it was taken and
 * how fresh it is.
 */
export interface Price {
  asset_id: string
  value_usd: Decimal | null
  currency: CurrencyCode
  status: PriceStatus
  freshness: DataFreshness
  as_of: Timestamp | null
  change_24h_pct: Decimal | null
}
