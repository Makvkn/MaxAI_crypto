import type { AIUsage } from '@/api/types'
import { Tooltip } from '@/components/ui/Tooltip'
import { formatTime } from '@/lib/dates/format'
import { cn } from '@/lib/utils/cn'

/**
 * Daily AI budget, as reported by the backend.
 *
 * Purely informational: the limit is enforced server-side, and this never gates
 * a request on a client-side count.
 */
export function AiUsageMeter({
  usage,
  className,
}: {
  usage: AIUsage
  className?: string
}) {
  const exhausted = usage.remaining <= 0

  return (
    <Tooltip
      content={`AI operations reset at ${formatTime(usage.resets_at)}. The dashboard stays fully available either way.`}
    >
      <span
        className={cn(
          'inline-flex items-center gap-2 text-[12px]',
          exhausted ? 'text-caution' : 'text-fg-subtle',
          className,
        )}
      >
        <span className="tabular">
          {usage.used} / {usage.limit}
        </span>
        <span>AI operations used today</span>
      </span>
    </Tooltip>
  )
}
