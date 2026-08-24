/**
 * Explicit authentication lifecycle state.
 *
 * AuthStatus is the single source of truth for whether the app has finished
 * bootstrapping and whether a validated session exists. React Query manages
 * session data; this module manages authentication readiness.
 */

export type AuthStatus = 'bootstrapping' | 'authenticated' | 'unauthenticated'

export function initialAuthStatus(hasTokens: boolean): AuthStatus {
  return hasTokens ? 'bootstrapping' : 'unauthenticated'
}

export function isAuthReady(status: AuthStatus): boolean {
  return status !== 'bootstrapping'
}

export function isAuthenticatedStatus(status: AuthStatus): boolean {
  return status === 'authenticated'
}

export function isAuthLoading(status: AuthStatus): boolean {
  return status === 'bootstrapping'
}
