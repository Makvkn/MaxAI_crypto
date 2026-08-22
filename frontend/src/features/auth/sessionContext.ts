import { createContext, useContext } from 'react'
import type { EmailCredentials, User } from '@/api/types'

/**
 * Session contract exposed to the application.
 *
 * The backend owns authentication and authorisation; this describes only the
 * transitions the product needs. Permission-driven UI is convenience — every
 * request is still authorised server-side.
 */
export interface SessionContextValue {
  user: User | null
  isAuthenticated: boolean
  isGuest: boolean
  isLoading: boolean
  /** Creates an anonymous account if none exists yet. */
  ensureSession: () => Promise<User>
  signInWithEmail: (credentials: EmailCredentials) => Promise<User>
  registerWithEmail: (credentials: EmailCredentials) => Promise<User>
  signInWithGoogle: () => Promise<User>
  /** Attaches credentials to the current guest account, keeping `user.id`. */
  upgradeWithEmail: (credentials: EmailCredentials) => Promise<User>
  upgradeWithGoogle: () => Promise<User>
  signOut: () => Promise<void>
  isMutating: boolean
}

export const SessionContext = createContext<SessionContextValue | null>(null)

export function useSession(): SessionContextValue {
  const context = useContext(SessionContext)
  if (!context) {
    throw new Error('useSession must be used inside SessionProvider')
  }
  return context
}
