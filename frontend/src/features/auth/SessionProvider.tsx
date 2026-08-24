import {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  apiCreateGuestSession,
  apiLoginWithEmail,
  apiLoginWithGoogle,
  apiLogout,
  apiRegisterWithEmail,
  apiUpgradeAccount,
  tokenStore,
} from '@/api'
import { resetAuthGate, syncAuthGate, waitForAuthReady } from '@/api/authGate'
import type {
  AuthSession,
  EmailCredentials,
  UpgradeAccountRequest,
  User,
} from '@/api/types'
import { UserKind } from '@/api/types'
import { analytics } from '@/lib/analytics/analytics'
import {
  initialAuthStatus,
  isAuthLoading,
  isAuthReady,
  isAuthenticatedStatus,
  type AuthStatus,
} from './authStatus'
import { SessionContext, type SessionContextValue } from './sessionContext'
import { createBootstrapEpoch } from './bootstrapEpoch'
import {
  runSessionBootstrap,
  sessionQueryOptions,
} from './sessionQueryOptions'
import { useTokens } from './useTokens'

/**
 * The Google ID token the backend verifies.
 *
 * A real integration obtains this from Google Identity Services in the browser.
 * The exchange stays a backend concern: no provider secret exists in this app.
 */
function googleIdToken(): string {
  return 'google-id-token'
}

const toUser = (session: AuthSession): User => session.user

/**
 * Session lifecycle.
 *
 * AuthStatus drives authentication readiness. TanStack Query caches the
 * current user but never defines whether bootstrap has finished.
 */
export function SessionProvider({ children }: { children: ReactNode }) {
  const tokens = useTokens()
  const queryClient = useQueryClient()

  const bootstrapEpoch = useRef(createBootstrapEpoch())
  const [internalStatus, setInternalStatus] = useState<AuthStatus>(() =>
    initialAuthStatus(Boolean(tokenStore.get())),
  )
  const authStatus: AuthStatus = !tokens ? 'unauthenticated' : internalStatus

  const { data: user = null } = useQuery({
    ...sessionQueryOptions(),
    enabled: false,
  })

  useEffect(() => {
    syncAuthGate(authStatus)
  }, [authStatus])

  useEffect(() => {
    if (!tokens) return
    if (internalStatus === 'authenticated') return

    if (internalStatus === 'unauthenticated') {
      setInternalStatus('bootstrapping')
      return
    }

    if (internalStatus !== 'bootstrapping') return

    const attempt = bootstrapEpoch.current.start()

    void runSessionBootstrap(queryClient)
      .then(() => {
        if (bootstrapEpoch.current.isCurrent(attempt)) {
          setInternalStatus('authenticated')
        }
      })
      .catch(() => {
        if (bootstrapEpoch.current.isCurrent(attempt)) {
          setInternalStatus('unauthenticated')
        }
      })
  }, [tokens, internalStatus, queryClient])

  const adopt = useCallback(
    (session: AuthSession): User => {
      bootstrapEpoch.current.invalidate()
      tokenStore.set({
        access_token: session.access_token,
        refresh_token: session.refresh_token,
        expires_at: session.expires_at,
      })
      queryClient.setQueryData(sessionQueryOptions().queryKey, session.user)
      setInternalStatus('authenticated')
      analytics.identify(session.user.id, {
        kind: session.user.kind,
        plan: session.user.subscription.plan,
      })
      return session.user
    },
    [queryClient],
  )

  const guestMutation = useMutation({
    mutationFn: () => apiCreateGuestSession(),
    onSuccess: adopt,
  })

  const emailLoginMutation = useMutation({
    mutationFn: (credentials: EmailCredentials) =>
      apiLoginWithEmail(credentials),
    onSuccess: (session) => {
      adopt(session)
      analytics.track('signed_in', { method: 'EMAIL' })
    },
  })

  const emailRegisterMutation = useMutation({
    mutationFn: (credentials: EmailCredentials) =>
      apiRegisterWithEmail(credentials),
    onSuccess: (session) => {
      adopt(session)
      analytics.track('signed_in', { method: 'EMAIL' })
    },
  })

  const googleMutation = useMutation({
    mutationFn: () => apiLoginWithGoogle({ id_token: googleIdToken() }),
    onSuccess: (session) => {
      adopt(session)
      analytics.track('signed_in', { method: 'GOOGLE' })
    },
  })

  const upgradeMutation = useMutation({
    mutationFn: (request: UpgradeAccountRequest) =>
      apiUpgradeAccount(request),
    onSuccess: (session, request) => {
      adopt(session)
      analytics.track('account_upgraded', { method: request.method })
    },
  })

  const logoutMutation = useMutation({
    mutationFn: () => apiLogout(),
    onSettled: () => {
      bootstrapEpoch.current.invalidate()
      tokenStore.clear()
      queryClient.clear()
      resetAuthGate()
      setInternalStatus('unauthenticated')
    },
  })

  const authReady = isAuthReady(authStatus)
  const isAuthenticated = isAuthenticatedStatus(authStatus)

  const ensureSession = useCallback(async () => {
    if (user) return user
    if (tokens) {
      await waitForAuthReady()
      const cached = queryClient.getQueryData<User>(
        sessionQueryOptions().queryKey,
      )
      if (cached) return cached
      return runSessionBootstrap(queryClient)
    }
    const session = await guestMutation.mutateAsync()
    return session.user
  }, [guestMutation, queryClient, tokens, user])

  const value = useMemo<SessionContextValue>(
    () => ({
      user,
      authReady,
      isAuthenticated,
      isGuest: user?.kind === UserKind.GUEST,
      isLoading: isAuthLoading(authStatus),
      ensureSession,
      signInWithEmail: (credentials) =>
        emailLoginMutation.mutateAsync(credentials).then(toUser),
      registerWithEmail: (credentials) =>
        emailRegisterMutation.mutateAsync(credentials).then(toUser),
      signInWithGoogle: () => googleMutation.mutateAsync().then(toUser),
      upgradeWithEmail: (credentials) =>
        upgradeMutation
          .mutateAsync({ method: 'EMAIL', ...credentials })
          .then(toUser),
      upgradeWithGoogle: () =>
        upgradeMutation
          .mutateAsync({ method: 'GOOGLE', id_token: googleIdToken() })
          .then(toUser),
      signOut: async () => {
        await logoutMutation.mutateAsync()
      },
      isMutating:
        guestMutation.isPending ||
        emailLoginMutation.isPending ||
        emailRegisterMutation.isPending ||
        googleMutation.isPending ||
        upgradeMutation.isPending,
    }),
    [
      authReady,
      authStatus,
      ensureSession,
      emailLoginMutation,
      emailRegisterMutation,
      googleMutation,
      guestMutation,
      isAuthenticated,
      logoutMutation,
      upgradeMutation,
      user,
    ],
  )

  return (
    <SessionContext.Provider value={value}>{children}</SessionContext.Provider>
  )
}
