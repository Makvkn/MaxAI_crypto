import { useEffect, useId, useRef, useState, type ReactNode } from 'react'
import { cn } from '@/lib/utils/cn'

export interface MenuItem {
  id: string
  label: ReactNode
  description?: ReactNode
  selected?: boolean
  disabled?: boolean
  onSelect: () => void
}

/**
 * Dropdown menu with keyboard support (Enter/Space to open, arrows to move,
 * Escape to close) and outside-click dismissal.
 */
export function Menu({
  trigger,
  items,
  label,
  align = 'start',
  className,
}: {
  trigger: (props: {
    open: boolean
    onClick: () => void
    'aria-expanded': boolean
    'aria-haspopup': 'menu'
    'aria-controls': string
  }) => ReactNode
  items: MenuItem[]
  label: string
  align?: 'start' | 'end'
  className?: string
}) {
  const menuId = useId()
  const [open, setOpen] = useState(false)
  const [activeIndex, setActiveIndex] = useState(0)
  const containerRef = useRef<HTMLDivElement>(null)
  const listRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!open) return

    function onPointerDown(event: PointerEvent) {
      if (!containerRef.current?.contains(event.target as Node)) setOpen(false)
    }
    function onKeyDown(event: KeyboardEvent) {
      if (event.key === 'Escape') setOpen(false)
    }

    document.addEventListener('pointerdown', onPointerDown)
    document.addEventListener('keydown', onKeyDown)
    listRef.current?.focus({ preventScroll: true })

    return () => {
      document.removeEventListener('pointerdown', onPointerDown)
      document.removeEventListener('keydown', onKeyDown)
    }
  }, [open])

  const enabled = items.filter((item) => !item.disabled)

  function onListKeyDown(event: React.KeyboardEvent<HTMLDivElement>) {
    if (event.key === 'ArrowDown') {
      event.preventDefault()
      setActiveIndex((index) => (index + 1) % enabled.length)
    } else if (event.key === 'ArrowUp') {
      event.preventDefault()
      setActiveIndex((index) => (index - 1 + enabled.length) % enabled.length)
    } else if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault()
      const item = enabled[activeIndex]
      if (item) {
        item.onSelect()
        setOpen(false)
      }
    }
  }

  return (
    <div ref={containerRef} className="relative">
      {trigger({
        open,
        onClick: () => setOpen((value) => !value),
        'aria-expanded': open,
        'aria-haspopup': 'menu',
        'aria-controls': menuId,
      })}

      {open ? (
        <div
          id={menuId}
          ref={listRef}
          role="menu"
          aria-label={label}
          tabIndex={-1}
          onKeyDown={onListKeyDown}
          className={cn(
            'absolute z-50 mt-2 min-w-56 animate-fade-in overflow-hidden rounded-xl border border-line-strong bg-surface-raised p-1 shadow-2xl outline-none',
            align === 'end' ? 'right-0' : 'left-0',
            className,
          )}
        >
          {items.map((item, index) => (
            <button
              key={item.id}
              type="button"
              role="menuitem"
              disabled={item.disabled}
              aria-current={item.selected || undefined}
              onMouseEnter={() => setActiveIndex(index)}
              onClick={() => {
                item.onSelect()
                setOpen(false)
              }}
              className={cn(
                'flex w-full flex-col items-start gap-0.5 rounded-lg px-3 py-2 text-left text-sm transition-colors disabled:opacity-40',
                index === activeIndex
                  ? 'bg-surface-hover text-fg'
                  : 'text-fg-muted hover:text-fg',
              )}
            >
              <span className="flex w-full items-center justify-between gap-3">
                <span className="truncate">{item.label}</span>
                {item.selected ? (
                  <span className="text-accent" aria-hidden="true">
                    •
                  </span>
                ) : null}
              </span>
              {item.description ? (
                <span className="text-[12px] text-fg-subtle">
                  {item.description}
                </span>
              ) : null}
            </button>
          ))}
        </div>
      ) : null}
    </div>
  )
}
