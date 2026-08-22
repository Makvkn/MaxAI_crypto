import '@testing-library/jest-dom/vitest'
import { afterEach, beforeEach, vi } from 'vitest'
import { cleanup } from '@testing-library/react'

/**
 * Node exposes a disabled `localStorage` global that shadows jsdom's, so tests
 * get an in-memory implementation instead. The application only depends on the
 * `Storage` surface it actually uses.
 */
function createMemoryStorage(): Storage {
  let entries = new Map<string, string>()

  return {
    get length() {
      return entries.size
    },
    key: (index) => [...entries.keys()][index] ?? null,
    getItem: (key) => entries.get(key) ?? null,
    setItem: (key, value) => void entries.set(key, String(value)),
    removeItem: (key) => void entries.delete(key),
    clear: () => {
      entries = new Map()
    },
  }
}

if (!globalThis.localStorage) {
  vi.stubGlobal('localStorage', createMemoryStorage())
}

// jsdom lacks the observers and matchMedia that layout components rely on.
beforeEach(() => {
  vi.stubGlobal(
    'matchMedia',
    vi.fn((query: string) => ({
      matches: false,
      media: query,
      onchange: null,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      addListener: vi.fn(),
      removeListener: vi.fn(),
      dispatchEvent: vi.fn(),
    })),
  )

  vi.stubGlobal(
    'ResizeObserver',
    class {
      observe() {}
      unobserve() {}
      disconnect() {}
    },
  )
})

afterEach(() => {
  cleanup()
  localStorage.clear()
})
