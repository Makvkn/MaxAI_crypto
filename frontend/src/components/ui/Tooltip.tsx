import { useId, useState, type ReactNode } from 'react'
import { cn } from '@/lib/utils/cn'

/**
 * Tooltip triggered by hover *and* focus, described via `aria-describedby` so
 * the content is available to screen readers rather than being purely visual.
 */
export function Tooltip({
  content,
  children,
  className,
  side = 'top',
}: {
  content: ReactNode
  children: ReactNode
  className?: string
  side?: 'top' | 'bottom'
}) {
  const id = useId()
  const [open, setOpen] = useState(false)

  return (
    <span className="relative inline-flex">
      <span
        aria-describedby={open ? id : undefined}
        onMouseEnter={() => setOpen(true)}
        onMouseLeave={() => setOpen(false)}
        onFocus={() => setOpen(true)}
        onBlur={() => setOpen(false)}
        className="inline-flex"
      >
        {children}
      </span>

      <span
        id={id}
        role="tooltip"
        hidden={!open}
        className={cn(
          'pointer-events-none absolute left-1/2 z-50 w-max max-w-64 -translate-x-1/2 rounded-lg border border-line-strong bg-surface-raised px-3 py-2 text-left text-[12px] leading-relaxed font-normal tracking-normal text-fg-muted normal-case shadow-xl',
          side === 'top' ? 'bottom-full mb-2' : 'top-full mt-2',
          className,
        )}
      >
        {content}
      </span>
    </span>
  )
}
