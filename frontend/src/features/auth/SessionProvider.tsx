import { useCallback, useMemo, type ReactNode } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api, tokenStore } from '@/api'
import type {
  AuthSession,
  EmailCredentials,
  UpgradeAccountRequest,
  User,
} from '@/api/types'
import { UserKind } from '@/api/types'
import { queryKeys } from '@/lib/query/queryKeys'
import { analytics } from '@/lib/analytics/analytics'
import { SessionContext, type SessionContextValue } from './sessionContext'
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
 * Tracks whether a session exists, resolves the current user through TanStack
 * Query and performs the transitions the product needs — including guest ->
 * registered upgrade, which preserves `user.id` server-side.
 */
export function SessionProvider({ children }: { children: ReactNode }) {
  const tokens = useTokens()
  const queryClient = useQueryClient()

  const sessionQuery = useQuery({
    queryKey: queryKeys.session(),
    queryFn: ({ signal }) => api.auth.getCurrentUser({ signal }),
    enabled: Boolean(tokens),
    staleTime: 5 * 60_000,
    retry: false,
  })

  const adopt = useCallback(
    (session: AuthSession): User => {
      tokenStore.set({
        access_token: session.access_token,
        refresh_token: session.refresh_token,
        expires_at: session.expires_at,
      })
      queryClient.setQueryData(queryKeys.session(), session.user)
      analytics.identify(session.user.id, {
        kind: session.user.kind,
        plan: session.user.subscription.plan,
      })
      return session.user
    },
    [queryClient],
  )

  const guestMutation = useMutation({
    mutationFn: () => api.auth.createGuestSession(),
    onSuccess: adopt,
  })

  const emailLoginMutation = useMutation({
    mutationFn: (credentials: EmailCredentials) =>
      api.auth.loginWithEmail(credentials),
    onSuccess: (session) => {
      adopt(session)
      analytics.track('signed_in', { method: 'EMAIL' })
    },
  })

  const emailRegisterMutation = useMutation({
    mutationFn: (credentials: EmailCredentials) =>
      api.auth.registerWithEmail(credentials),
    onSuccess: (session) => {
      adopt(session)
      analytics.track('signed_in', { method: 'EMAIL' })
    },
  })

  const googleMutation = useMutation({
    mutationFn: () => api.auth.loginWithGoogle({ id_token: googleIdToken() }),
    onSuccess: (session) => {
      adopt(session)
      analytics.track('signed_in', { method: 'GOOGLE' })
    },
  })

  const upgradeMutation = useMutation({
    mutationFn: (request: UpgradeAccountRequest) =>
      api.auth.upgradeAccount(request),
    onSuccess: (session, request) => {
      adopt(session)
      analytics.track('account_upgraded', { method: request.method })
    },
  })

  const logoutMutation = useMutation({
    mutationFn: () => api.auth.logout(),
    onSettled: () => {
      tokenStore.clear()
      queryClient.clear()
    },
  })

  const user = sessionQuery.data ?? null

  const ensureSession = useCallback(async () => {
    if (user) return user
    if (tokens) {
      const existing = await queryClient.fetchQuery({
        queryKey: queryKeys.session(),
        queryFn: ({ signal }) => api.auth.getCurrentUser({ signal }),
      })
      return existing
    }
    const session = await guestMutation.mutateAsync()
    return session.user
  }, [guestMutation, queryClient, tokens, user])

  const value = useMemo<SessionContextValue>(
    () => ({
      user,
      isAuthenticated: Boolean(tokens),
      isGuest: user?.kind === UserKind.GUEST,
      isLoading: Boolean(tokens) && sessionQuery.isLoading,
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
      ensureSession,
      emailLoginMutation,
      emailRegisterMutation,
      googleMutation,
      guestMutation,
      logoutMutation,
      sessionQuery.isLoading,
      tokens,
      upgradeMutation,
      user,
    ],
  )

  return (
    <SessionContext.Provider value={value}>{children}</SessionContext.Provider>
  )
}
