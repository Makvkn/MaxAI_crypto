import type { Decimal } from '@/api/types'
import { UNKNOWN_VALUE } from './money'

/**
 * Percentages arrive from the backend already expressed in percent
 * (`"-4.21"` means −4.21%). These helpers format; they never derive a rate.
 */

export function formatPercent(
  value: Decimal | null | undefined,
  options?: { decimals?: number },
): string {
  if (value === null || value === undefined || value === '') return UNKNOWN_VALUE
  const amount = Number(value)
  if (!Number.isFinite(amount)) return UNKNOWN_VALUE
  return `${Math.abs(amount).toFixed(options?.decimals ?? 2)}%`
}

/** Signed percentage for deltas: `+6.42%` / `-4.21%`. */
export function formatPercentDelta(
  value: Decimal | null | undefined,
  options?: { decimals?: number },
): string {
  if (value === null || value === undefined || value === '') return UNKNOWN_VALUE
  const amount = Number(value)
  if (!Number.isFinite(amount)) return UNKNOWN_VALUE
  const sign = amount > 0 ? '+' : amount < 0 ? '-' : ''
  return `${sign}${Math.abs(amount).toFixed(options?.decimals ?? 2)}%`
}

export type DeltaDirection = 'up' | 'down' | 'flat' | 'unknown'

/** Direction of a delta, for colour and iconography only. */
export function deltaDirection(
  value: Decimal | null | undefined,
): DeltaDirection {
  if (value === null || value === undefined || value === '') return 'unknown'
  const amount = Number(value)
  if (!Number.isFinite(amount)) return 'unknown'
  if (amount > 0) return 'up'
  if (amount < 0) return 'down'
  return 'flat'
}
