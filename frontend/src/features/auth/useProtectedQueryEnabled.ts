import { useSession } from './sessionContext'

/**
 * Central gate for authenticated API queries.
 *
 * Protected requests must wait until auth bootstrap completes and the session
 * is validated — not merely until tokens exist in storage.
 */
export function useProtectedQueryEnabled(enabled = true): boolean {
  const { isAuthInitialized, isAuthenticated } = useSession()
  return enabled && isAuthInitialized && isAuthenticated
}
