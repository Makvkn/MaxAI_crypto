import type { HTMLAttributes, ReactNode } from 'react'
import { cn } from '@/lib/utils/cn'

export type BadgeTone =
  | 'neutral'
  | 'accent'
  | 'positive'
  | 'negative'
  | 'caution'
  | 'muted'

const TONES: Record<BadgeTone, string> = {
  neutral: 'border-line-strong bg-neutral-quiet text-fg-muted',
  accent: 'border-accent/30 bg-accent-quiet text-accent',
  positive: 'border-positive/25 bg-positive-quiet text-positive',
  negative: 'border-negative/25 bg-negative-quiet text-negative',
  caution: 'border-caution/25 bg-caution-quiet text-caution',
  muted: 'border-transparent bg-transparent text-fg-subtle',
}

export function Badge({
  tone = 'neutral',
  icon,
  className,
  children,
  ...props
}: HTMLAttributes<HTMLSpanElement> & { tone?: BadgeTone; icon?: ReactNode }) {
  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-[11px] font-medium tracking-[0.04em] uppercase',
        TONES[tone],
        className,
      )}
      {...props}
    >
      {icon}
      {children}
    </span>
  )
}
