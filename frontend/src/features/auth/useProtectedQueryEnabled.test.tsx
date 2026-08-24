import { describe, expect, it } from 'vitest'
import { renderHook } from '@testing-library/react'
import type { ReactNode } from 'react'
import { SessionContext, type SessionContextValue } from './sessionContext'
import { useProtectedQueryEnabled } from './useProtectedQueryEnabled'

function wrapper(value: SessionContextValue) {
  return function Provider({ children }: { children: ReactNode }) {
    return (
      <SessionContext.Provider value={value}>{children}</SessionContext.Provider>
    )
  }
}

function baseSession(
  overrides: Partial<SessionContextValue> = {},
): SessionContextValue {
  return {
    user: null,
    authReady: false,
    isAuthenticated: false,
    isGuest: false,
    isLoading: true,
    ensureSession: async () => {
      throw new Error('not implemented')
    },
    signInWithEmail: async () => {
      throw new Error('not implemented')
    },
    registerWithEmail: async () => {
      throw new Error('not implemented')
    },
    signInWithGoogle: async () => {
      throw new Error('not implemented')
    },
    upgradeWithEmail: async () => {
      throw new Error('not implemented')
    },
    upgradeWithGoogle: async () => {
      throw new Error('not implemented')
    },
    signOut: async () => {},
    isMutating: false,
    ...overrides,
  }
}

describe('useProtectedQueryEnabled', () => {
  it('is disabled while auth bootstrap is in progress', () => {
    const { result } = renderHook(() => useProtectedQueryEnabled(true), {
      wrapper: wrapper(
        baseSession({
          authReady: false,
          isAuthenticated: false,
          isLoading: true,
        }),
      ),
    })

    expect(result.current).toBe(false)
  })

  it('is disabled when bootstrap finished without a session', () => {
    const { result } = renderHook(() => useProtectedQueryEnabled(true), {
      wrapper: wrapper(
        baseSession({
          authReady: true,
          isAuthenticated: false,
          isLoading: false,
        }),
      ),
    })

    expect(result.current).toBe(false)
  })

  it('is enabled after bootstrap succeeds', () => {
    const { result } = renderHook(() => useProtectedQueryEnabled(true), {
      wrapper: wrapper(
        baseSession({
          authReady: true,
          isAuthenticated: true,
          isLoading: false,
        }),
      ),
    })

    expect(result.current).toBe(true)
  })

  it('stays disabled with cached user during bootstrap', () => {
    const { result } = renderHook(() => useProtectedQueryEnabled(true), {
      wrapper: wrapper(
        baseSession({
          user: { id: 'cached-user' } as SessionContextValue['user'],
          authReady: false,
          isAuthenticated: false,
          isLoading: true,
        }),
      ),
    })

    expect(result.current).toBe(false)
  })

  it('respects an additional enabled=false condition', () => {
    const { result } = renderHook(() => useProtectedQueryEnabled(false), {
      wrapper: wrapper(
        baseSession({
          authReady: true,
          isAuthenticated: true,
          isLoading: false,
        }),
      ),
    })

    expect(result.current).toBe(false)
  })
})
