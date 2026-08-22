import {
  AssetVisibility,
  DataFreshness,
  DataNoticeCode,
  DataQuality,
  NoticeSeverity,
  PriceStatus,
  ValuationStatus,
  type DataNotice,
  type Portfolio,
  type Wallet,
  type WalletPosition,
} from '../../types'
import * as d from '../support/decimal'
import type { MockVariant } from '../variants'
import { toAsset } from './assets'
import {
  UNPRICED_VISIBLE_ASSET,
  holdingsForChain,
  type MockAssetSpec,
} from './catalog'

/**
 * Portfolio valuation — MOCK BACKEND SIMULATION.
 *
 * Mirrors what `PortfolioService` would do: value every priced position,
 * exclude the unpriced ones, and report the resulting quality honestly.
 */

const CALCULATION_VERSION = 3

/** Freshness thresholds; a backend configuration value in the real system. */
export function freshnessFor(ageMinutes: number): DataFreshness {
  if (ageMinutes < 5) return DataFreshness.FRESH
  if (ageMinutes < 15) return DataFreshness.RECENT
  if (ageMinutes <= 60) return DataFreshness.STALE
  return DataFreshness.VERY_STALE
}

export function syncAgeMinutes(variant: MockVariant): number {
  if (variant.veryStale) return 184
  if (variant.stale) return 42
  return 3
}

/** Spam is excluded from valuation entirely; dust is not. */
function countsTowardValuation(spec: MockAssetSpec): boolean {
  return spec.visibility !== AssetVisibility.HIDDEN_SPAM
}

export function holdingsFor(
  wallet: Wallet,
  variant: MockVariant,
): MockAssetSpec[] {
  if (variant.empty) return []
  const base = holdingsForChain(wallet.chain_id)
  return variant.partialValuation ? [...base, UNPRICED_VISIBLE_ASSET] : base
}

export interface BuiltPortfolio {
  portfolio: Portfolio
  /** Positions keyed by asset id, for reuse by transactions and scenarios. */
  bySymbol: Map<string, WalletPosition>
}

export function buildPortfolio(
  wallet: Wallet,
  variant: MockVariant,
  now: Date,
): BuiltPortfolio {
  const specs = holdingsFor(wallet, variant)
  const ageMinutes = syncAgeMinutes(variant)
  const lastSyncedAt = new Date(now.getTime() - ageMinutes * 60_000)
  const freshness = freshnessFor(ageMinutes)
  const priceTimestamp = lastSyncedAt.toISOString()

  const valued = specs.map((spec) => {
    const priced = spec.price_usd !== null
    const valueUsd =
      priced && countsTowardValuation(spec)
        ? d.multiply(spec.balance, spec.price_usd as string, 2)
        : null
    return { spec, valueUsd }
  })

  const total = d.add(
    valued.map((entry) => entry.valueUsd).filter((v): v is string => v !== null),
    2,
  )
  const hasValuedPositions = valued.some((entry) => entry.valueUsd !== null)

  const positions: WalletPosition[] = valued.map(({ spec, valueUsd }) => {
    const priceValue = spec.price_usd
    const changePct = spec.change_24h_pct

    // 24h delta of the position, derived from the previous day's valuation.
    let changeUsd: string | null = null
    if (valueUsd !== null && changePct !== null) {
      const previous = d.divide(
        valueUsd,
        d.add(['1', d.multiply(changePct, '0.01', 18)], 18),
        2,
      )
      changeUsd = previous === null ? null : d.subtract(valueUsd, previous, 2)
    }

    const asset = toAsset(wallet.chain_id, spec)

    return {
      asset,
      balance: spec.balance,
      balance_raw: rawBalance(spec.balance, spec.decimals),
      price:
        priceValue === null
          ? {
              asset_id: asset.id,
              value_usd: null,
              currency: 'USD',
              status: PriceStatus.UNAVAILABLE,
              freshness,
              as_of: null,
              change_24h_pct: null,
            }
          : {
              asset_id: asset.id,
              value_usd: priceValue,
              currency: 'USD',
              status: PriceStatus.AVAILABLE,
              freshness,
              as_of: priceTimestamp,
              change_24h_pct: changePct,
            },
      value_usd: valueUsd,
      allocation_pct:
        valueUsd === null || d.isZero(total)
          ? null
          : d.percentOf(valueUsd, total, 2),
      change_24h_pct: changePct,
      change_24h_usd: changeUsd,
      visibility: spec.visibility,
      valuation_status:
        valueUsd === null ? ValuationStatus.UNAVAILABLE : ValuationStatus.COMPLETE,
      updated_at: lastSyncedAt.toISOString(),
    }
  })

  const unpricedVisible = positions.filter(
    (position) =>
      position.visibility === AssetVisibility.VISIBLE &&
      position.value_usd === null,
  ).length

  const portfolioChangeUsd = d.add(
    positions
      .map((position) => position.change_24h_usd)
      .filter((value): value is string => value !== null),
    2,
  )
  const openingValue = d.subtract(total, portfolioChangeUsd, 2)

  const unavailable = variant.portfolioUnavailable || (!hasValuedPositions && !variant.empty)

  const valuationStatus = unavailable
    ? ValuationStatus.UNAVAILABLE
    : unpricedVisible > 0
      ? ValuationStatus.PARTIAL
      : ValuationStatus.COMPLETE

  const dataQuality = unavailable
    ? DataQuality.UNAVAILABLE
    : unpricedVisible > 0
      ? DataQuality.PARTIAL
      : freshness === DataFreshness.STALE ||
          freshness === DataFreshness.VERY_STALE
        ? DataQuality.STALE
        : DataQuality.COMPLETE

  const portfolio: Portfolio = {
    wallet_id: wallet.id,
    currency: 'USD',
    total_value_usd: unavailable ? null : total,
    valuation_status: valuationStatus,
    data_quality: dataQuality,
    data_freshness: freshness,
    change_24h_usd: unavailable || variant.empty ? null : portfolioChangeUsd,
    change_24h_pct:
      unavailable || variant.empty || d.isZero(openingValue)
        ? null
        : d.percentOf(portfolioChangeUsd, openingValue, 2),
    as_of: now.toISOString(),
    last_synced_at: lastSyncedAt.toISOString(),
    calculation_version: CALCULATION_VERSION,
    positions: unavailable ? positions.map(stripValuation) : positions,
    visible_positions_count: positions.filter(
      (position) => position.visibility === AssetVisibility.VISIBLE,
    ).length,
    hidden_positions_count: positions.filter(
      (position) => position.visibility !== AssetVisibility.VISIBLE,
    ).length,
    exclusions: {
      unpriced_positions: positions.filter(
        (position) => position.value_usd === null,
      ).length,
      nfts_excluded: true,
      defi_positions_excluded: true,
    },
    notices: buildNotices({
      unpricedVisible,
      freshness,
      ageMinutes,
      syncPartial: variant.syncPartial,
    }),
  }

  return {
    portfolio,
    bySymbol: new Map(
      positions.map((position) => [position.asset.symbol, position]),
    ),
  }
}

function stripValuation(position: WalletPosition): WalletPosition {
  return {
    ...position,
    value_usd: null,
    allocation_pct: null,
    change_24h_usd: null,
    valuation_status: ValuationStatus.UNAVAILABLE,
  }
}

function buildNotices(input: {
  unpricedVisible: number
  freshness: DataFreshness
  ageMinutes: number
  syncPartial: boolean
}): DataNotice[] {
  const notices: DataNotice[] = []

  if (input.unpricedVisible > 0) {
    notices.push({
      code: DataNoticeCode.UNPRICED_ASSETS_EXCLUDED,
      severity: NoticeSeverity.WARNING,
      params: { count: input.unpricedVisible },
    })
  }

  if (
    input.freshness === DataFreshness.STALE ||
    input.freshness === DataFreshness.VERY_STALE
  ) {
    notices.push({
      code: DataNoticeCode.DATA_STALE,
      severity: NoticeSeverity.WARNING,
      params: { minutes: input.ageMinutes },
    })
  }

  if (input.syncPartial) {
    notices.push({
      code: DataNoticeCode.SYNC_PARTIALLY_FAILED,
      severity: NoticeSeverity.WARNING,
    })
  }

  notices.push({
    code: DataNoticeCode.NFTS_EXCLUDED_FROM_VALUATION,
    severity: NoticeSeverity.INFO,
  })

  return notices
}

/** Unscaled on-chain amount, as a provider would report it. */
function rawBalance(balance: string, decimals: number): string {
  const [intPart = '0', fracPart = ''] = balance.split('.')
  const padded = `${fracPart}${'0'.repeat(decimals)}`.slice(0, decimals)
  return `${intPart}${padded}`.replace(/^0+(?=\d)/, '')
}
