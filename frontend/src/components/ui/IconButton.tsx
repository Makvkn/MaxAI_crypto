import type { ButtonHTMLAttributes, ReactNode } from 'react'
import { cn } from '@/lib/utils/cn'

/** Icon-only button. `label` is required: it becomes the accessible name. */
export function IconButton({
  label,
  children,
  className,
  size = 'md',
  ...props
}: ButtonHTMLAttributes<HTMLButtonElement> & {
  label: string
  children: ReactNode
  size?: 'sm' | 'md'
}) {
  return (
    <button
      type="button"
      aria-label={label}
      title={label}
      className={cn(
        'inline-flex items-center justify-center rounded-lg border border-transparent text-fg-muted transition-colors hover:bg-surface-raised hover:text-fg focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent disabled:opacity-40',
        size === 'sm' ? 'size-7 text-[14px]' : 'size-9 text-[17px]',
        className,
      )}
      {...props}
    >
      {children}
    </button>
  )
}
