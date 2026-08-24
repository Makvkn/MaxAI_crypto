import { describe, expect, it } from 'vitest'
import {
  initialAuthStatus,
  isAuthLoading,
  isAuthReady,
  isAuthenticatedStatus,
} from './authStatus'

describe('authStatus', () => {
  it('starts bootstrapping when tokens exist', () => {
    expect(initialAuthStatus(true)).toBe('bootstrapping')
  })

  it('starts unauthenticated when no tokens exist', () => {
    expect(initialAuthStatus(false)).toBe('unauthenticated')
  })

  it('derives authReady from status', () => {
    expect(isAuthReady('bootstrapping')).toBe(false)
    expect(isAuthReady('authenticated')).toBe(true)
    expect(isAuthReady('unauthenticated')).toBe(true)
  })

  it('derives isAuthenticated from status only', () => {
    expect(isAuthenticatedStatus('bootstrapping')).toBe(false)
    expect(isAuthenticatedStatus('authenticated')).toBe(true)
    expect(isAuthenticatedStatus('unauthenticated')).toBe(false)
  })

  it('derives isLoading from bootstrapping status', () => {
    expect(isAuthLoading('bootstrapping')).toBe(true)
    expect(isAuthLoading('authenticated')).toBe(false)
  })
})
