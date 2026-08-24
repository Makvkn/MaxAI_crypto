import { beforeEach, describe, expect, it } from 'vitest'
import {
  isAuthBootstrapping,
  isAuthReady,
  markAuthPending,
  markAuthReady,
  resetAuthGate,
  runAuthBootstrap,
  waitForAuthReady,
} from './authGate'

describe('authGate', () => {
  beforeEach(() => {
    resetAuthGate()
  })

  it('blocks protected requests until bootstrap completes', async () => {
    markAuthPending()

    let released = false
    const waiter = waitForAuthReady().then(() => {
      released = true
    })

    await Promise.resolve()
    expect(released).toBe(false)

    markAuthReady()
    await waiter
    expect(released).toBe(true)
  })

  it('runs bootstrap without blocking internal authenticated calls', async () => {
    const result = await runAuthBootstrap(async () => {
      expect(isAuthBootstrapping()).toBe(true)
      expect(isAuthReady()).toBe(false)
      return 'ok'
    })

    expect(result).toBe('ok')
    expect(isAuthReady()).toBe(true)
    expect(isAuthBootstrapping()).toBe(false)
  })
})
