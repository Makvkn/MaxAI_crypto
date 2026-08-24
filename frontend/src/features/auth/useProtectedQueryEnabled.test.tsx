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
    isAuthInitialized: false,
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
  it('is disabled while auth initialization is in progress', () => {
    const { result } = renderHook(() => useProtectedQueryEnabled(true), {
      wrapper: wrapper(
        baseSession({
          isAuthInitialized: false,
          isAuthenticated: false,
          isLoading: true,
        }),
      ),
    })

    expect(result.current).toBe(false)
  })

  it('is disabled for guests without a validated session', () => {
    const { result } = renderHook(() => useProtectedQueryEnabled(true), {
      wrapper: wrapper(
        baseSession({
          isAuthInitialized: true,
          isAuthenticated: false,
          isLoading: false,
        }),
      ),
    })

    expect(result.current).toBe(false)
  })

  it('is enabled after auth initialization succeeds', () => {
    const { result } = renderHook(() => useProtectedQueryEnabled(true), {
      wrapper: wrapper(
        baseSession({
          isAuthInitialized: true,
          isAuthenticated: true,
          isLoading: false,
        }),
      ),
    })

    expect(result.current).toBe(true)
  })

  it('respects an additional enabled=false condition', () => {
    const { result } = renderHook(() => useProtectedQueryEnabled(false), {
      wrapper: wrapper(
        baseSession({
          isAuthInitialized: true,
          isAuthenticated: true,
          isLoading: false,
        }),
      ),
    })

    expect(result.current).toBe(false)
  })
})
