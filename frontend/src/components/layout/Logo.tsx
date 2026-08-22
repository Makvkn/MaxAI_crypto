import { Link } from 'react-router-dom'
import { cn } from '@/lib/utils/cn'

/** Wordmark. The glyph is a chart line resolving into a single point. */
export function Logo({
  className,
  to = '/',
}: {
  className?: string
  to?: string
}) {
  return (
    <Link
      to={to}
      className={cn(
        'group inline-flex items-center gap-2.5 rounded-md text-fg',
        className,
      )}
      aria-label="MaxAI Crypto — home"
    >
      <span className="relative grid size-7 place-items-center rounded-lg border border-line-strong bg-surface-raised">
        <svg viewBox="0 0 24 24" className="size-4" aria-hidden="true">
          <path
            d="M5 17V7l4.5 7L14 7v10"
            fill="none"
            stroke="var(--color-accent)"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
          <circle cx="19" cy="8" r="2" fill="var(--color-positive)" />
        </svg>
      </span>
      <span className="text-[15px] font-medium tracking-tight">
        MaxAI<span className="text-fg-subtle"> Crypto</span>
      </span>
    </Link>
  )
}
