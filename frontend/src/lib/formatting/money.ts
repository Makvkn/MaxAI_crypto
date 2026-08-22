import type { Decimal } from '@/api/types'

/**
 * Presentation formatting for backend-provided values.
 *
 * These helpers never change a value's meaning: they round for display only.
 * A `null` input means the backend does not know the value, and is rendered as
 * an em dash — never as `$0`.
 */

/** What the UI shows when a value is genuinely unknown. */
export const UNKNOWN_VALUE = '—'

const usd = new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'USD',
  minimumFractionDigits: 2,
  maximumFractionDigits: 2,
})

const usdWhole = new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'USD',
  maximumFractionDigits: 0,
})

const usdCompact = new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'USD',
  notation: 'compact',
  maximumFractionDigits: 1,
})

const usdPrecise = new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'USD',
  minimumFractionDigits: 2,
  maximumFractionDigits: 6,
})

/**
 * `"24850.123456"` -> `"$24,850.12"`.
 *
 * Returns the unknown marker for `null`/`undefined`, so callers cannot
 * accidentally display a fabricated zero.
 */
export function formatUsd(
  value: Decimal | null | undefined,
  options?: { compact?: boolean; whole?: boolean },
): string {
  if (value === null || value === undefined || value === '') return UNKNOWN_VALUE
  const amount = Number(value)
  if (!Number.isFinite(amount)) return UNKNOWN_VALUE

  if (options?.compact) return usdCompact.format(amount)
  if (options?.whole) return usdWhole.format(amount)
  return usd.format(amount)
}

/** Prices can be very small (`$0.0000241`), so more decimals are kept. */
export function formatPrice(value: Decimal | null | undefined): string {
  if (value === null || value === undefined || value === '') return UNKNOWN_VALUE
  const amount = Number(value)
  if (!Number.isFinite(amount)) return UNKNOWN_VALUE
  if (amount !== 0 && Math.abs(amount) < 0.01) {
    return `$${amount.toPrecision(3)}`
  }
  return usdPrecise.format(amount)
}

/** Signed money for deltas: `+$412.30` / `-$1,092.44`. */
export function formatUsdDelta(
  value: Decimal | null | undefined,
  options?: { compact?: boolean },
): string {
  if (value === null || value === undefined || value === '') return UNKNOWN_VALUE
  const amount = Number(value)
  if (!Number.isFinite(amount)) return UNKNOWN_VALUE

  const formatted = formatUsd(
    Math.abs(amount).toString(),
    options?.compact ? { compact: true } : undefined,
  )
  if (amount === 0) return formatted
  return `${amount > 0 ? '+' : '-'}${formatted}`
}

/**
 * Crypto balances keep enough precision to stay honest: large balances are
 * rounded, small ones are not.
 */
export function formatBalance(
  value: Decimal | null | undefined,
  options?: { symbol?: string },
): string {
  if (value === null || value === undefined || value === '') return UNKNOWN_VALUE
  const amount = Number(value)
  if (!Number.isFinite(amount)) return UNKNOWN_VALUE

  const absolute = Math.abs(amount)
  const maximumFractionDigits =
    absolute >= 10_000 ? 2 : absolute >= 1 ? 4 : absolute >= 0.0001 ? 6 : 8

  const formatted = new Intl.NumberFormat('en-US', {
    maximumFractionDigits,
  }).format(amount)

  return options?.symbol ? `${formatted} ${options.symbol}` : formatted
}

/** Plain counts and axis labels. */
export function formatNumber(
  value: number | null | undefined,
  options?: { compact?: boolean },
): string {
  if (value === null || value === undefined || !Number.isFinite(value)) {
    return UNKNOWN_VALUE
  }
  return new Intl.NumberFormat('en-US', {
    notation: options?.compact ? 'compact' : 'standard',
    maximumFractionDigits: options?.compact ? 1 : 0,
  }).format(value)
}
