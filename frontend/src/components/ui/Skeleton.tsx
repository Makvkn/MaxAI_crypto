import { cn } from '@/lib/utils/cn'

/**
 * Contextual loading placeholder.
 *
 * Skeletons are preferred over spinners so the layout does not jump and the
 * user can see what is about to arrive.
 */
export function Skeleton({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={cn(
        'animate-pulse-soft rounded-md bg-surface-raised',
        className,
      )}
      aria-hidden="true"
      {...props}
    />
  )
}

/** Repeated skeleton rows for list and table placeholders. */
export function SkeletonRows({
  rows = 4,
  className,
}: {
  rows?: number
  className?: string
}) {
  return (
    <div className={cn('space-y-3', className)} aria-hidden="true">
      {Array.from({ length: rows }).map((_, index) => (
        <div key={index} className="flex items-center gap-3">
          <Skeleton className="size-8 rounded-full" />
          <Skeleton className="h-3 flex-1" />
          <Skeleton className="h-3 w-16" />
        </div>
      ))}
    </div>
  )
}
