import type { Timestamp } from './types'

/**
 * Session token persistence.
 *
 * Kept behind an interface because the transport may change: if the backend
 * moves refresh tokens into httpOnly cookies, only this module is replaced.
 */
export interface StoredTokens {
  access_token: string
  refresh_token: string
  expires_at: Timestamp
}

export interface TokenStore {
  get(): StoredTokens | null
  set(tokens: StoredTokens): void
  clear(): void
  /** Notifies on cross-tab and in-tab changes. Returns an unsubscribe fn. */
  subscribe(listener: (tokens: StoredTokens | null) => void): () => void
}

const STORAGE_KEY = 'maxai.session.tokens'

function isStoredTokens(value: unknown): value is StoredTokens {
  if (typeof value !== 'object' || value === null) return false
  const candidate = value as Partial<StoredTokens>
  return (
    typeof candidate.access_token === 'string' &&
    typeof candidate.refresh_token === 'string' &&
    typeof candidate.expires_at === 'string'
  )
}

/**
 * localStorage-backed store with an in-memory mirror, so reads stay cheap and
 * the app keeps working if storage is unavailable (private mode, quota).
 */
export function createTokenStore(storageKey = STORAGE_KEY): TokenStore {
  const listeners = new Set<(tokens: StoredTokens | null) => void>()
  let cache: StoredTokens | null = read()

  function storage(): Storage | null {
    try {
      return globalThis.localStorage ?? null
    } catch {
      return null
    }
  }

  function read(): StoredTokens | null {
    const store = storage()
    if (!store) return null
    try {
      const raw = store.getItem(storageKey)
      if (!raw) return null
      const parsed: unknown = JSON.parse(raw)
      return isStoredTokens(parsed) ? parsed : null
    } catch {
      return null
    }
  }

  function emit() {
    for (const listener of listeners) listener(cache)
  }

  if (typeof globalThis.addEventListener === 'function') {
    globalThis.addEventListener('storage', (event) => {
      if (!(event instanceof StorageEvent) || event.key !== storageKey) return
      cache = read()
      emit()
    })
  }

  return {
    get: () => cache,
    set(tokens) {
      cache = tokens
      try {
        storage()?.setItem(storageKey, JSON.stringify(tokens))
      } catch {
        // Non-fatal: the session survives in memory for this tab.
      }
      emit()
    },
    clear() {
      cache = null
      try {
        storage()?.removeItem(storageKey)
      } catch {
        // Non-fatal.
      }
      emit()
    },
    subscribe(listener) {
      listeners.add(listener)
      return () => listeners.delete(listener)
    },
  }
}
