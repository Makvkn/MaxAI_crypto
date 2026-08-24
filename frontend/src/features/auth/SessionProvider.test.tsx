import { act, renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import type { ReactNode } from 'react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { tokenStore } from '@/api'
import type { AuthSession, User } from '@/api/types'
import { SessionProvider } from './SessionProvider'
import { useSession } from './sessionContext'
import * as sessionQueryOptionsModule from './sessionQueryOptions'

const storedTokens = {
  access_token: 'access',
  refresh_token: 'refresh',
  expires_at: '2099-01-01T00:00:00.000Z',
}

const { signedInSession } = vi.hoisted(() => ({
  signedInSession: {
    access_token: 'new-access',
    refresh_token: 'refresh',
    expires_at: '2099-01-01T00:00:00.000Z',
    user: {
      id: 'signed-in-user',
      kind: 'REGISTERED',
      subscription: { plan: 'FREE' },
    },
  } as AuthSession,
}))

let resolveBootstrap!: (user: User) => void
let rejectBootstrap!: (error: unknown) => void

vi.mock('./sessionQueryOptions', async (importOriginal) => {
  const actual =
    await importOriginal<typeof sessionQueryOptionsModule>()
  return {
    ...actual,
    runSessionBootstrap: vi.fn(
      () =>
        new Promise<User>((resolve, reject) => {
          resolveBootstrap = resolve
          rejectBootstrap = reject
        }),
    ),
  }
})

vi.mock('@/api', async () => {
  const actual = await vi.importActual<typeof import('@/api')>('@/api')
  return {
    ...actual,
    apiLogout: vi.fn().mockResolvedValue(undefined),
    apiLoginWithEmail: vi.fn().mockResolvedValue(signedInSession),
  }
})

vi.mock('@/lib/analytics/analytics', () => ({
  analytics: {
    identify: vi.fn(),
    track: vi.fn(),
  },
}))

function createWrapper(queryClient: QueryClient) {
  return function Provider({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        <SessionProvider>{children}</SessionProvider>
      </QueryClientProvider>
    )
  }
}

describe('SessionProvider bootstrap epoch', () => {
  beforeEach(() => {
    tokenStore.clear()
    vi.clearAllMocks()
  })

  it('ignores stale bootstrap success after logout', async () => {
    tokenStore.set(storedTokens)
    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    })

    const { result } = renderHook(() => useSession(), {
      wrapper: createWrapper(queryClient),
    })

    await waitFor(() => {
      expect(result.current.isLoading).toBe(true)
    })

    await act(async () => {
      await result.current.signOut()
    })

    expect(result.current.authReady).toBe(true)
    expect(result.current.isAuthenticated).toBe(false)

    await act(async () => {
      resolveBootstrap({ id: 'stale-user' } as User)
    })

    expect(result.current.isAuthenticated).toBe(false)
  })

  it('ignores stale bootstrap failure after adopt', async () => {
    tokenStore.set(storedTokens)
    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    })

    const { result } = renderHook(() => useSession(), {
      wrapper: createWrapper(queryClient),
    })

    await waitFor(() => {
      expect(result.current.isLoading).toBe(true)
    })

    await act(async () => {
      await result.current.signInWithEmail({
        email: 'user@example.com',
        password: 'secret',
      })
    })

    expect(result.current.isAuthenticated).toBe(true)
    expect(result.current.user).toEqual(signedInSession.user)

    await act(async () => {
      rejectBootstrap(new Error('stale bootstrap failed'))
    })

    expect(result.current.isAuthenticated).toBe(true)
    expect(result.current.user).toEqual(signedInSession.user)
  })
})
