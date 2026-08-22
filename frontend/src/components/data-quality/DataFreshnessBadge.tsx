import { DataFreshness, type Timestamp } from '@/api/types'
import { Badge, type BadgeTone } from '@/components/ui/Badge'
import { Clock } from '@/components/ui/Icon'
import { Tooltip } from '@/components/ui/Tooltip'
import { formatRelativeTime } from '@/lib/dates/format'
import { freshnessLabels } from '@/lib/copy/labels'

const TONES: Record<DataFreshness, BadgeTone> = {
  [DataFreshness.FRESH]: 'positive',
  [DataFreshness.RECENT]: 'neutral',
  [DataFreshness.STALE]: 'caution',
  [DataFreshness.VERY_STALE]: 'negative',
}

/**
 * How old the underlying data is. Stale data is never hidden — it is labelled,
 * with the exact time available on hover and to screen readers.
 */
export function DataFreshnessBadge({
  freshness,
  asOf,
}: {
  freshness: DataFreshness
  asOf: Timestamp | null
}) {
  const relative = asOf ? formatRelativeTime(asOf) : 'time unknown'

  return (
    <Tooltip content={`Portfolio data last updated ${relative}.`}>
      <Badge tone={TONES[freshness]} icon={<Clock className="size-3" />}>
        <span>{freshnessLabels[freshness]}</span>
        <span className="sr-only"> — updated {relative}</span>
      </Badge>
    </Tooltip>
  )
}
