import { NoticeSeverity, type DataNotice } from '@/api/types'
import { Info, Warning } from '@/components/ui/Icon'
import { noticeMessage } from '@/lib/copy/notices'
import { cn } from '@/lib/utils/cn'

/**
 * Renders backend data-quality notices.
 *
 * Warnings are surfaced prominently because they change what the user is
 * entitled to believe about the numbers above them. Informational notices (such
 * as NFTs being excluded) stay quiet.
 */
export function DataQualityBanner({
  notices,
  className,
}: {
  notices: DataNotice[]
  className?: string
}) {
  const warnings = notices.filter(
    (notice) => notice.severity === NoticeSeverity.WARNING,
  )
  if (warnings.length === 0) return null

  return (
    <div
      role="status"
      className={cn(
        'flex items-start gap-3 rounded-card border border-caution/25 bg-caution-quiet/60 px-4 py-3',
        className,
      )}
    >
      <Warning className="mt-0.5 size-4 shrink-0 text-caution" />
      <div className="space-y-1 text-[13px] leading-relaxed text-fg-muted">
        {warnings.map((notice) => (
          <p key={notice.code}>{noticeMessage(notice)}</p>
        ))}
      </div>
    </div>
  )
}

/** Quiet footnote for informational notices. */
export function DataQualityFootnotes({
  notices,
  className,
}: {
  notices: DataNotice[]
  className?: string
}) {
  const infos = notices.filter(
    (notice) => notice.severity === NoticeSeverity.INFO,
  )
  if (infos.length === 0) return null

  return (
    <p
      className={cn(
        'flex items-center gap-1.5 text-[12px] text-fg-subtle',
        className,
      )}
    >
      <Info className="size-3.5 shrink-0" />
      {infos.map((notice) => noticeMessage(notice)).join(' ')}
    </p>
  )
}
