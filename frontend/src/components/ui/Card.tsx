import type { HTMLAttributes, ReactNode } from 'react'
import { cn } from '@/lib/utils/cn'

/** Surface primitive for dashboard panels. */
export function Card({
  className,
  ...props
}: HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={cn(
        'rounded-card border border-line bg-surface',
        'shadow-[0_1px_0_0_rgba(255,255,255,0.02)_inset]',
        className,
      )}
      {...props}
    />
  )
}

export function CardHeader({
  title,
  description,
  action,
  className,
  ...props
}: HTMLAttributes<HTMLDivElement> & {
  title: ReactNode
  description?: ReactNode
  action?: ReactNode
}) {
  return (
    <div
      className={cn(
        'flex items-start justify-between gap-4 border-b border-line px-5 py-4',
        className,
      )}
      {...props}
    >
      <div className="min-w-0">
        <h2 className="text-[13px] font-medium tracking-[0.06em] text-fg-muted uppercase">
          {title}
        </h2>
        {description ? (
          <p className="mt-1 text-sm text-fg-subtle">{description}</p>
        ) : null}
      </div>
      {action ? <div className="shrink-0">{action}</div> : null}
    </div>
  )
}

export function CardBody({
  className,
  ...props
}: HTMLAttributes<HTMLDivElement>) {
  return <div className={cn('px-5 py-4', className)} {...props} />
}

export function CardFooter({
  className,
  ...props
}: HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={cn('border-t border-line px-5 py-3 text-sm', className)}
      {...props}
    />
  )
}
