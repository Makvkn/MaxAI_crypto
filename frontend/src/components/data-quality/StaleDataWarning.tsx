import { DataFreshness, type Timestamp } from '@/api/types'
import { Clock, Refresh } from '@/components/ui/Icon'
import { Button } from '@/components/ui/Button'
import { minutesSince } from '@/lib/dates/format'

/**
 * Inline stale-data warning with an explicit refresh affordance.
 *
 * The MVP has no realtime updates, so asking for a refresh is a legitimate user
 * action rather than something the UI should hide.
 */
export function StaleDataWarning({
  freshness,
  lastSyncedAt,
  onRefresh,
  refreshing,
}: {
  freshness: DataFreshness
  lastSyncedAt: Timestamp | null
  onRefresh?: () => void
  refreshing?: boolean
}) {
  if (
    freshness !== DataFreshness.STALE &&
    freshness !== DataFreshness.VERY_STALE
  ) {
    return null
  }

  const minutes = minutesSince(lastSyncedAt)

  return (
    <div
      role="status"
      className="flex flex-wrap items-center justify-between gap-3 rounded-card border border-line bg-base-elevated px-4 py-3"
    >
      <p className="flex items-center gap-2 text-[13px] text-fg-muted">
        <Clock className="size-4 shrink-0 text-caution" />
        {minutes === null
          ? 'This view is based on data that is no longer fresh.'
          : `Based on portfolio data last updated ${minutes} minutes ago.`}
      </p>
      {onRefresh ? (
        <Button
          variant="quiet"
          size="sm"
          onClick={onRefresh}
          loading={refreshing}
          iconLeft={<Refresh className="size-3.5" />}
        >
          Refresh
        </Button>
      ) : null}
    </div>
  )
}
