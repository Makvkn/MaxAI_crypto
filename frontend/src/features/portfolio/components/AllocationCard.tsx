import {
  AssetVisibility,
  ValuationStatus,
  type Portfolio,
  type WalletPosition,
} from '@/api/types'
import { Card, CardBody, CardHeader } from '@/components/ui/Card'
import { EmptyState } from '@/components/feedback/States'
import { Money } from '@/components/finance/Money'
import { formatPercent } from '@/lib/formatting/percent'

/** Palette for allocation segments — ordered, so the bar stays readable. */
const SEGMENT_COLORS = [
  'var(--color-accent)',
  '#5ad1c8',
  '#8f7ff0',
  '#e0a458',
  '#35c48f',
  '#d97a9a',
  '#7c93b8',
  '#b0894f',
]

/**
 * Asset allocation.
 *
 * `allocation_pct` is a backend figure; this component only sorts and draws it.
 * Positions the backend could not value carry no share and are listed without
 * one instead of being folded into a made-up remainder.
 */
export function AllocationCard({ portfolio }: { portfolio: Portfolio }) {
  const priced = portfolio.positions
    .filter(
      (position) =>
        position.visibility === AssetVisibility.VISIBLE &&
        position.allocation_pct !== null,
    )
    .sort((a, b) => Number(b.allocation_pct) - Number(a.allocation_pct))

  const shown = priced.slice(0, 8)
  const remaining = priced.length - shown.length

  if (portfolio.valuation_status === ValuationStatus.UNAVAILABLE) {
    return (
      <Card>
        <CardHeader title="Allocation" />
        <EmptyState
          title="Allocation is unavailable"
          description="Allocation is a share of a portfolio value, and that value could not be calculated."
        />
      </Card>
    )
  }

  if (portfolio.positions.length === 0) {
    return (
      <Card>
        <CardHeader title="Allocation" />
        <EmptyState
          title="No assets to allocate"
          description="This wallet has no holdings, so there is nothing to break down."
        />
      </Card>
    )
  }

  if (shown.length === 0) {
    return (
      <Card>
        <CardHeader title="Allocation" />
        <EmptyState
          title="Nothing to allocate yet"
          description="No held asset currently has both a balance and a reliable market price."
        />
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader
        title="Allocation"
        action={
          remaining > 0 ? (
            <span className="text-[12px] text-fg-subtle">
              top {shown.length} of {priced.length}
            </span>
          ) : null
        }
      />
      <CardBody>
        <div
          className="flex h-2 w-full gap-0.5 overflow-hidden rounded-full bg-base-elevated"
          role="img"
          aria-label={`Allocation across ${shown.length} assets`}
        >
          {shown.map((position, index) => (
            <span
              key={position.asset.id}
              className="h-full first:rounded-l-full last:rounded-r-full"
              style={{
                width: `${position.allocation_pct}%`,
                backgroundColor: segmentColor(index),
              }}
            />
          ))}
        </div>

        <ul className="mt-5 space-y-2.5">
          {shown.map((position, index) => (
            <AllocationRow
              key={position.asset.id}
              position={position}
              color={segmentColor(index)}
            />
          ))}
        </ul>

        {remaining > 0 ? (
          <p className="mt-4 text-[12px] text-fg-subtle">
            {remaining} smaller {remaining === 1 ? 'position' : 'positions'} are
            listed in Assets.
          </p>
        ) : null}
      </CardBody>
    </Card>
  )
}

function AllocationRow({
  position,
  color,
}: {
  position: WalletPosition
  color: string
}) {
  return (
    <li className="flex items-center gap-3 text-sm">
      <span
        className="size-2 shrink-0 rounded-full"
        style={{ backgroundColor: color }}
        aria-hidden="true"
      />
      <span className="min-w-0 flex-1 truncate text-fg-muted">
        {position.asset.symbol}
      </span>
      <span className="w-14 text-right text-fg-subtle tabular">
        {formatPercent(position.allocation_pct)}
      </span>
      <span className="w-24 text-right text-fg">
        <Money value={position.value_usd} />
      </span>
    </li>
  )
}

function segmentColor(index: number): string {
  return SEGMENT_COLORS[index % SEGMENT_COLORS.length] ?? 'var(--color-accent)'
}
