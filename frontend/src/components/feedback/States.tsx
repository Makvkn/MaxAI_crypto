import type { ReactNode } from 'react'
import { Button } from '@/components/ui/Button'
import { Refresh, Warning } from '@/components/ui/Icon'
import { errorCopy } from '@/lib/errors/messages'
import { cn } from '@/lib/utils/cn'

/**
 * Shared empty and error presentation.
 *
 * `ErrorState` derives its copy from the domain error code, so a component
 * never has to decide how to phrase a backend failure — and raw API messages,
 * stack traces and provider names never reach the screen.
 */

export function EmptyState({
  title,
  description,
  action,
  icon,
  className,
}: {
  title: string
  description?: ReactNode
  action?: ReactNode
  icon?: ReactNode
  className?: string
}) {
  return (
    <div
      className={cn(
        'flex flex-col items-center justify-center gap-3 px-6 py-12 text-center',
        className,
      )}
    >
      {icon ? <div className="text-xl text-fg-subtle">{icon}</div> : null}
      <p className="text-sm font-medium text-fg">{title}</p>
      {description ? (
        <p className="max-w-sm text-[13px] leading-relaxed text-fg-subtle">
          {description}
        </p>
      ) : null}
      {action}
    </div>
  )
}

export function ErrorState({
  error,
  onRetry,
  retrying,
  className,
  compact,
}: {
  error: unknown
  onRetry?: () => void
  retrying?: boolean
  className?: string
  compact?: boolean
}) {
  const copy = errorCopy(error)

  return (
    <div
      role="alert"
      className={cn(
        'flex flex-col items-start gap-3',
        compact ? 'px-4 py-4' : 'px-6 py-10 text-center items-center',
        className,
      )}
    >
      <Warning className="size-5 text-caution" />
      <div className={compact ? '' : 'space-y-1'}>
        <p className="text-sm font-medium text-fg">{copy.title}</p>
        <p
          className={cn(
            'text-[13px] leading-relaxed text-fg-subtle',
            compact ? 'mt-1' : 'max-w-sm',
          )}
        >
          {copy.description}
        </p>
      </div>
      {onRetry && copy.retryable ? (
        <Button
          variant="secondary"
          size="sm"
          onClick={onRetry}
          loading={retrying}
          iconLeft={<Refresh className="size-3.5" />}
        >
          Try again
        </Button>
      ) : null}
    </div>
  )
}
