import { beforeEach, describe, expect, it } from 'vitest'
import {
  isAuthBootstrapping,
  isAuthGateOpen,
  resetAuthGate,
  runAuthBootstrap,
  syncAuthGate,
  waitForAuthReady,
} from './authGate'

describe('authGate', () => {
  beforeEach(() => {
    resetAuthGate()
  })

  it('keeps the gate closed while bootstrapping', () => {
    syncAuthGate('bootstrapping')
    expect(isAuthGateOpen()).toBe(false)
  })

  it('opens the gate for authenticated and unauthenticated states', () => {
    syncAuthGate('bootstrapping')
    syncAuthGate('authenticated')
    expect(isAuthGateOpen()).toBe(true)

    resetAuthGate()
    syncAuthGate('bootstrapping')
    syncAuthGate('unauthenticated')
    expect(isAuthGateOpen()).toBe(true)
  })

  it('blocks protected requests until bootstrap completes', async () => {
    syncAuthGate('bootstrapping')

    let released = false
    const waiter = waitForAuthReady().then(() => {
      released = true
    })

    await Promise.resolve()
    expect(released).toBe(false)

    syncAuthGate('authenticated')
    await waiter
    expect(released).toBe(true)
  })

  it('runs bootstrap without blocking internal authenticated calls', async () => {
    syncAuthGate('bootstrapping')

    const result = await runAuthBootstrap(async () => {
      expect(isAuthBootstrapping()).toBe(true)
      expect(isAuthGateOpen()).toBe(false)
      return 'ok'
    })

    expect(result).toBe('ok')
    expect(isAuthBootstrapping()).toBe(false)
  })
})
