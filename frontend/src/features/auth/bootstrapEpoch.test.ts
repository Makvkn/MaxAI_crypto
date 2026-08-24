import { describe, expect, it } from 'vitest'
import { createBootstrapEpoch } from './bootstrapEpoch'

describe('createBootstrapEpoch', () => {
  it('marks earlier attempts stale after invalidate', () => {
    const epoch = createBootstrapEpoch()

    const attempt = epoch.start()
    expect(epoch.isCurrent(attempt)).toBe(true)

    epoch.invalidate()
    expect(epoch.isCurrent(attempt)).toBe(false)
  })

  it('isolates a new attempt after invalidate', () => {
    const epoch = createBootstrapEpoch()

    const first = epoch.start()
    epoch.invalidate()
    const second = epoch.start()

    expect(epoch.isCurrent(first)).toBe(false)
    expect(epoch.isCurrent(second)).toBe(true)
  })
})
