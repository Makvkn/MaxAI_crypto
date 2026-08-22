import {
  ApiErrorCode,
  DataQuality,
  PerformancePeriod,
  PerformanceStatus,
  ValuationStatus,
  type PerformanceDriver,
  type Portfolio,
  type PortfolioPerformance,
  type PortfolioSnapshotPoint,
  type Wallet,
} from '../../types'
import * as d from '../support/decimal'
import { createRandom, hashString } from '../support/random'
import type { MockVariant } from '../variants'

/**
 * Portfolio snapshots and snapshot-based performance — MOCK BACKEND
 * SIMULATION of `SnapshotService` + `PortfolioService`.
 *
 * The series is generated deterministically per wallet so history is stable
 * across reloads. The frontend never fabricates or interpolates these points.
 */

interface PeriodShape {
  points: number
  intervalMs: number
  /** Deterministic total return over the period, in percent. */
  drift: number
  volatility: number
}

const HOUR = 3_600_000
const DAY = 24 * HOUR

const PERIOD_SHAPES: Record<PerformancePeriod, PeriodShape> = {
  [PerformancePeriod.H24]: {
    points: 96,
    intervalMs: 15 * 60_000,
    drift: 0,
    volatility: 0.0035,
  },
  [PerformancePeriod.D7]: {
    points: 84,
    intervalMs: 2 * HOUR,
    drift: 6.4,
    volatility: 0.006,
  },
  [PerformancePeriod.D30]: {
    points: 60,
    intervalMs: 12 * HOUR,
    drift: -9.8,
    volatility: 0.011,
  },
  [PerformancePeriod.ALL]: {
    points: 90,
    intervalMs: DAY,
    drift: 71.5,
    volatility: 0.019,
  },
}

export function buildPerformance(
  wallet: Wallet,
  portfolio: Portfolio,
  period: PerformancePeriod,
  variant: MockVariant,
  now: Date,
): PortfolioPerformance {
  const base: PortfolioPerformance = {
    wallet_id: wallet.id,
    period,
    status: PerformanceStatus.UNAVAILABLE,
    data_quality: portfolio.data_quality,
    currency: 'USD',
    opening: null,
    closing: null,
    change_usd: null,
    change_pct: null,
    series: [],
    drivers: [],
    unavailable_reason: null,
    calculation_id: null,
    calculation_version: null,
  }

  const currentValue = portfolio.total_value_usd

  // No valuation, no snapshots, or an empty wallet: nothing to compare.
  if (currentValue === null || variant.noHistory) {
    return {
      ...base,
      unavailable_reason: variant.noHistory
        ? ApiErrorCode.PERFORMANCE_DATA_UNAVAILABLE
        : ApiErrorCode.PORTFOLIO_DATA_UNAVAILABLE,
      data_quality: DataQuality.UNAVAILABLE,
    }
  }

  const shape = PERIOD_SHAPES[period]
  const openingValue =
    period === PerformancePeriod.H24 && portfolio.change_24h_usd !== null
      ? d.subtract(currentValue, portfolio.change_24h_usd, 2)
      : d.divide(
          currentValue,
          d.add(['1', d.multiply(String(shape.drift), '0.01', 18)], 18),
          2,
        )

  if (openingValue === null || d.isZero(openingValue)) {
    return { ...base, unavailable_reason: ApiErrorCode.PERFORMANCE_DATA_UNAVAILABLE }
  }

  const series = buildSeries({
    walletId: wallet.id,
    period,
    shape,
    openingValue,
    currentValue,
    now,
    // A partial valuation leaves a gap in history rather than a fake number.
    withGap: variant.partialValuation,
  })

  const firstPoint = series.find((point) => point.total_value_usd !== null)
  const lastPoint = [...series]
    .reverse()
    .find((point) => point.total_value_usd !== null)

  if (!firstPoint || !lastPoint) {
    return { ...base, unavailable_reason: ApiErrorCode.PERFORMANCE_DATA_UNAVAILABLE }
  }

  const changeUsd = d.subtract(
    lastPoint.total_value_usd as string,
    firstPoint.total_value_usd as string,
    2,
  )

  const partial =
    portfolio.valuation_status === ValuationStatus.PARTIAL ||
    portfolio.data_quality === DataQuality.PARTIAL

  return {
    ...base,
    status: partial ? PerformanceStatus.PARTIAL : PerformanceStatus.AVAILABLE,
    data_quality: portfolio.data_quality,
    opening: {
      captured_at: firstPoint.captured_at,
      value_usd: firstPoint.total_value_usd as string,
      status: firstPoint.status,
    },
    closing: {
      captured_at: lastPoint.captured_at,
      value_usd: lastPoint.total_value_usd as string,
      status: lastPoint.status,
    },
    change_usd: changeUsd,
    change_pct: d.percentOf(changeUsd, firstPoint.total_value_usd as string, 2),
    series,
    drivers: buildDrivers(portfolio, period),
    calculation_id: `calc_${wallet.id}_${period}`,
    calculation_version: portfolio.calculation_version,
  }
}

function buildSeries(input: {
  walletId: string
  period: PerformancePeriod
  shape: PeriodShape
  openingValue: string
  currentValue: string
  now: Date
  withGap: boolean
}): PortfolioSnapshotPoint[] {
  const { shape, openingValue, currentValue, now } = input
  const random = createRandom(hashString(`${input.walletId}:${input.period}`))

  const open = Number(openingValue)
  const close = Number(currentValue)

  // Random walk normalised so the endpoints match the reported values exactly.
  const walk: number[] = [0]
  for (let index = 1; index < shape.points; index += 1) {
    walk.push((walk[index - 1] as number) + (random() - 0.5) * shape.volatility)
  }
  const walkEnd = walk[walk.length - 1] as number

  return walk.map((offset, index) => {
    const progress = index / (shape.points - 1)
    const detrended = offset - walkEnd * progress
    const value = (open + (close - open) * progress) * (1 + detrended)
    const capturedAt = new Date(
      now.getTime() - (shape.points - 1 - index) * shape.intervalMs,
    )

    const isGap = input.withGap && index === Math.floor(shape.points * 0.42)

    return {
      captured_at: capturedAt.toISOString(),
      total_value_usd: isGap ? null : value.toFixed(2),
      status: isGap ? ValuationStatus.UNAVAILABLE : ValuationStatus.COMPLETE,
    }
  })
}

/**
 * Per-asset contribution to the period result. For 24h these come straight
 * from the position deltas; longer periods use a deterministic per-asset return.
 */
function buildDrivers(
  portfolio: Portfolio,
  period: PerformancePeriod,
): PerformanceDriver[] {
  const priced = portfolio.positions.filter(
    (position) => position.value_usd !== null && position.allocation_pct !== null,
  )

  return priced
    .map((position) => {
      const random = createRandom(
        hashString(`${position.asset.id}:${period}:driver`),
      )
      const changePct =
        period === PerformancePeriod.H24
          ? position.change_24h_pct
          : (random() * 40 - 18).toFixed(2)

      const contributionUsd =
        period === PerformancePeriod.H24
          ? position.change_24h_usd
          : changePct === null
            ? null
            : d.multiply(
                position.value_usd as string,
                d.multiply(changePct, '0.01', 18),
                2,
              )

      return {
        asset: position.asset,
        allocation_pct: position.allocation_pct,
        contribution_usd: contributionUsd,
        contribution_pct:
          contributionUsd === null || portfolio.total_value_usd === null
            ? null
            : d.percentOf(contributionUsd, portfolio.total_value_usd, 2),
        change_pct: changePct,
      }
    })
    .sort((a, b) => magnitude(b.contribution_usd) - magnitude(a.contribution_usd))
    .slice(0, 5)
}

function magnitude(value: string | null): number {
  return value === null ? 0 : Math.abs(Number(value))
}
