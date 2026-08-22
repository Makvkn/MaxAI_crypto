import { useEffect, type RefObject } from 'react'

const FOCUSABLE = [
  'a[href]',
  'button:not([disabled])',
  'input:not([disabled])',
  'select:not([disabled])',
  'textarea:not([disabled])',
  '[tabindex]:not([tabindex="-1"])',
].join(',')

/**
 * Keeps Tab focus inside a container while it is active, and restores focus to
 * the previously focused element on teardown. Used by dialogs and drawers.
 */
export function useFocusTrap(
  ref: RefObject<HTMLElement | null>,
  active: boolean,
): void {
  useEffect(() => {
    if (!active) return
    const container = ref.current
    if (!container) return

    const previouslyFocused = document.activeElement as HTMLElement | null

    const focusables = () =>
      Array.from(container.querySelectorAll<HTMLElement>(FOCUSABLE)).filter(
        (element) => element.offsetParent !== null || element === document.activeElement,
      )

    const initial = focusables()[0] ?? container
    initial.focus({ preventScroll: true })

    function onKeyDown(event: KeyboardEvent) {
      if (event.key !== 'Tab') return
      const elements = focusables()
      if (elements.length === 0) {
        event.preventDefault()
        return
      }

      const first = elements[0] as HTMLElement
      const last = elements[elements.length - 1] as HTMLElement
      const active = document.activeElement

      if (event.shiftKey && (active === first || active === container)) {
        event.preventDefault()
        last.focus()
      } else if (!event.shiftKey && active === last) {
        event.preventDefault()
        first.focus()
      }
    }

    container.addEventListener('keydown', onKeyDown)
    return () => {
      container.removeEventListener('keydown', onKeyDown)
      previouslyFocused?.focus({ preventScroll: true })
    }
  }, [ref, active])
}
