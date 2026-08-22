import { useRef } from 'react'
import { cn } from '@/lib/utils/cn'

export interface SegmentOption<T extends string> {
  value: T
  label: string
}

/**
 * Radio-group segmented control with arrow-key navigation. Used for the
 * performance period, where the choice is a filter rather than a page.
 */
export function SegmentedControl<T extends string>({
  options,
  value,
  onChange,
  label,
  size = 'md',
  className,
}: {
  options: readonly SegmentOption<T>[]
  value: T
  onChange: (value: T) => void
  label: string
  size?: 'sm' | 'md'
  className?: string
}) {
  const containerRef = useRef<HTMLDivElement>(null)

  function onKeyDown(event: React.KeyboardEvent<HTMLDivElement>) {
    const currentIndex = options.findIndex((option) => option.value === value)
    if (currentIndex === -1) return

    let nextIndex: number | null = null
    if (event.key === 'ArrowRight' || event.key === 'ArrowDown') {
      nextIndex = (currentIndex + 1) % options.length
    } else if (event.key === 'ArrowLeft' || event.key === 'ArrowUp') {
      nextIndex = (currentIndex - 1 + options.length) % options.length
    } else if (event.key === 'Home') {
      nextIndex = 0
    } else if (event.key === 'End') {
      nextIndex = options.length - 1
    }

    if (nextIndex === null) return
    event.preventDefault()
    const next = options[nextIndex]
    if (!next) return
    onChange(next.value)
    containerRef.current
      ?.querySelectorAll<HTMLButtonElement>('[role="radio"]')
      ?.item(nextIndex)
      ?.focus()
  }

  return (
    <div
      ref={containerRef}
      role="radiogroup"
      aria-label={label}
      onKeyDown={onKeyDown}
      className={cn(
        'inline-flex items-center gap-0.5 rounded-lg border border-line bg-base-elevated p-0.5',
        className,
      )}
    >
      {options.map((option) => {
        const selected = option.value === value
        return (
          <button
            key={option.value}
            type="button"
            role="radio"
            aria-checked={selected}
            tabIndex={selected ? 0 : -1}
            onClick={() => onChange(option.value)}
            className={cn(
              'rounded-[6px] font-medium transition-colors focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-accent',
              size === 'sm' ? 'px-2.5 py-1 text-[12px]' : 'px-3 py-1.5 text-[13px]',
              selected
                ? 'bg-surface-raised text-fg'
                : 'text-fg-subtle hover:text-fg-muted',
            )}
          >
            {option.label}
          </button>
        )
      })}
    </div>
  )
}
