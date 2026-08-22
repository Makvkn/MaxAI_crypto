import { useSyncExternalStore } from 'react'
import { tokenStore } from '@/api'
import type { StoredTokens } from '@/api/tokenStore'

/** Reactive view of the persisted token pair. */
export function useTokens(): StoredTokens | null {
  return useSyncExternalStore(
    (listener) => tokenStore.subscribe(listener),
    () => tokenStore.get(),
    () => null,
  )
}
