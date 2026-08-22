import type { Decimal } from '@/api/types'
import { formatUsd, formatUsdDelta, UNKNOWN_VALUE } from '@/lib/formatting/money'
import {
  deltaDirection,
  formatPercentDelta,
} from '@/lib/formatting/percent'
import { cn } from '@/lib/utils/cn'
import { ArrowDownRight, ArrowUpRight } from '@/components/ui/Icon'

/**
 * Financial value rendering.
 *
 * One rule runs through this file: an unknown value is shown as an em dash with
 * an accessible explanation. It is never rendered as `$0`, and nothing here
 * derives a number — every input comes from the backend.
 */

export function UnknownValue({
  reason = 'Not available',
  className,
}: {
  reason?: string
  className?: string
}) {
  return (
    <span
      className={cn('text-fg-subtle', className)}
      title={reason}
      aria-label={reason}
    >
      {UNKNOWN_VALUE}
    </span>
  )
}

export function Money({
  value,
  compact,
  className,
  unknownReason,
}: {
  value: Decimal | null | undefined
  compact?: boolean
  className?: string
  unknownReason?: string
}) {
  if (value === null || value === undefined || value === '') {
    return <UnknownValue reason={unknownReason} className={className} />
  }
  return (
    <span className={cn('tabular', className)}>
      {formatUsd(value, compact ? { compact: true } : undefined)}
    </span>
  )
}

/**
 * A signed change, in percent and optionally in USD, coloured by direction.
 * Direction comes from the backend value; the component only reads its sign.
 */
export function Delta({
  percent,
  amount,
  size = 'md',
  showIcon = true,
  className,
}: {
  percent: Decimal | null | undefined
  amount?: Decimal | null | undefined
  size?: 'sm' | 'md' | 'lg'
  showIcon?: boolean
  className?: string
}) {
  const direction = deltaDirection(percent)

  if (direction === 'unknown') {
    return <UnknownValue reason="Change is not available" className={className} />
  }

  const tone =
    direction === 'up'
      ? 'text-positive'
      : direction === 'down'
        ? 'text-negative'
        : 'text-fg-muted'

  const Icon = direction === 'down' ? ArrowDownRight : ArrowUpRight

  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 tabular',
        tone,
        size === 'sm' && 'text-[13px]',
        size === 'lg' && 'text-lg',
        className,
      )}
    >
      {showIcon && direction !== 'flat' ? (
        <Icon className="size-[1.05em] shrink-0" />
      ) : null}
      <span>{formatPercentDelta(percent)}</span>
      {amount !== undefined && amount !== null ? (
        <span className="text-fg-subtle">({formatUsdDelta(amount)})</span>
      ) : null}
    </span>
  )
}
