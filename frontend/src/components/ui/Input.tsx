import { useId, type InputHTMLAttributes, type ReactNode } from 'react'
import { cn } from '@/lib/utils/cn'

interface InputProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'id'> {
  label: string
  /** Hidden visually but kept for screen readers. */
  hideLabel?: boolean
  hint?: ReactNode
  error?: string | null
  adornment?: ReactNode
  monospace?: boolean
}

/** Labelled text input with description/error wiring for assistive tech. */
export function Input({
  label,
  hideLabel,
  hint,
  error,
  adornment,
  monospace,
  className,
  ...props
}: InputProps) {
  const id = useId()
  const hintId = `${id}-hint`
  const errorId = `${id}-error`
  const describedBy =
    [error ? errorId : null, hint ? hintId : null].filter(Boolean).join(' ') ||
    undefined

  return (
    <div className="w-full">
      <label
        htmlFor={id}
        className={cn(
          'mb-2 block text-[13px] font-medium text-fg-muted',
          hideLabel && 'sr-only',
        )}
      >
        {label}
      </label>

      <div className="relative">
        <input
          id={id}
          className={cn(
            'w-full rounded-lg border bg-base-elevated px-3.5 py-3 text-sm text-fg transition-colors',
            'placeholder:text-fg-subtle/70',
            'focus:outline-none focus-visible:border-accent focus-visible:ring-2 focus-visible:ring-accent/25',
            error
              ? 'border-negative/60'
              : 'border-line-strong hover:border-line-strong/80',
            monospace && 'font-mono text-[13px] tracking-tight',
            adornment && 'pr-11',
            className,
          )}
          aria-invalid={error ? true : undefined}
          aria-describedby={describedBy}
          {...props}
        />
        {adornment ? (
          <span className="absolute top-1/2 right-3 -translate-y-1/2 text-fg-subtle">
            {adornment}
          </span>
        ) : null}
      </div>

      {error ? (
        <p id={errorId} role="alert" className="mt-2 text-[13px] text-negative">
          {error}
        </p>
      ) : hint ? (
        <p id={hintId} className="mt-2 text-[13px] text-fg-subtle">
          {hint}
        </p>
      ) : null}
    </div>
  )
}
