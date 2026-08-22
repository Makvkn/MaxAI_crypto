import { describe, expect, it } from 'vitest'
import {
  UNKNOWN_VALUE,
  formatBalance,
  formatPrice,
  formatUsd,
  formatUsdDelta,
} from './money'
import { deltaDirection, formatPercent, formatPercentDelta } from './percent'
import { truncateMiddle } from './address'

describe('money formatting', () => {
  it('formats a backend decimal without changing its meaning', () => {
    expect(formatUsd('24850.123456')).toBe('$24,850.12')
    expect(formatUsd('24850.123456', { compact: true })).toBe('$24.9K')
  })

  it('renders unknown values as an em dash, never as zero', () => {
    expect(formatUsd(null)).toBe(UNKNOWN_VALUE)
    expect(formatUsd(undefined)).toBe(UNKNOWN_VALUE)
    expect(formatUsd('')).toBe(UNKNOWN_VALUE)
    expect(formatPrice(null)).toBe(UNKNOWN_VALUE)
    expect(formatBalance(null)).toBe(UNKNOWN_VALUE)
  })

  it('still formats an explicit zero the backend reports', () => {
    expect(formatUsd('0')).toBe('$0.00')
  })

  it('keeps precision for very small prices', () => {
    expect(formatPrice('0.0000241')).toBe('$0.0000241')
  })

  it('signs deltas', () => {
    expect(formatUsdDelta('412.3')).toBe('+$412.30')
    expect(formatUsdDelta('-1092.44')).toBe('-$1,092.44')
    expect(formatUsdDelta(null)).toBe(UNKNOWN_VALUE)
  })

  it('keeps large balances readable and small ones exact', () => {
    expect(formatBalance('500000')).toBe('500,000')
    expect(formatBalance('0.00004215')).toBe('0.00004215')
    expect(formatBalance('1.5', { symbol: 'ETH' })).toBe('1.5 ETH')
  })
})

describe('percent formatting', () => {
  it('formats percentages that already arrive as percent', () => {
    expect(formatPercent('52.4')).toBe('52.40%')
    expect(formatPercentDelta('-4.21')).toBe('-4.21%')
    expect(formatPercentDelta('6.4')).toBe('+6.40%')
    expect(formatPercent(null)).toBe(UNKNOWN_VALUE)
  })

  it('derives direction only from the sign of a backend value', () => {
    expect(deltaDirection('-1')).toBe('down')
    expect(deltaDirection('1')).toBe('up')
    expect(deltaDirection('0')).toBe('flat')
    expect(deltaDirection(null)).toBe('unknown')
  })
})

describe('address truncation', () => {
  it('shortens long values and leaves short ones alone', () => {
    expect(truncateMiddle('0x71C7656EC7ab88b098defB751B7401B5f6d8976F')).toBe(
      '0x71C7…976F',
    )
    expect(truncateMiddle('short')).toBe('short')
  })
})
