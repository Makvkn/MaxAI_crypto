import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { ApiError } from '../errors'
import {
  ApiErrorCode,
  AssetVisibility,
  DataNoticeCode,
  PerformancePeriod,
  SyncStatus,
  ValuationStatus,
  type AIStreamEvent,
} from '../types'
import { createMockApi } from './index'

/**
 * The mock adapter is the backend the UI develops against, so its behaviour is
 * tested as a contract: asynchronous sync, unknown prices, cursor pagination,
 * streaming and usage limits must all behave the way the real API is specified
 * to behave.
 */

const HEALTHY = '0x71C7656EC7ab88b098defB751B7401B5f6d8976F'
const PARTIAL = '0xpartial000000000000000000000000000000001'
const FAILING = '0xfail0000000000000000000000000000000000001'

/** Long enough for the derived sync state machine to finish. */
const SYNC_MS = 12_000

describe('mock API', () => {
  let api: ReturnType<typeof createMockApi>

  beforeEach(() => {
    vi.useFakeTimers()
    localStorage.clear()
    api = createMockApi()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  /** Advances the fake clock so pending latency and the sync job resolve. */
  async function settle<T>(promise: Promise<T>, ms = 2_000): Promise<T> {
    const result = promise.then(
      (value) => ({ ok: true as const, value }),
      (error: unknown) => ({ ok: false as const, error }),
    )
    await vi.advanceTimersByTimeAsync(ms)
    const settled = await result
    if (!settled.ok) throw settled.error
    return settled.value
  }

  async function analysedWallet(address = HEALTHY) {
    await settle(api.auth.createGuestSession())
    const created = await settle(
      api.wallets.create({ chain_id: 'ethereum', address }),
    )
    await vi.advanceTimersByTimeAsync(SYNC_MS)
    return settle(api.wallets.get(created.id))
  }

  it('returns a wallet that is only queued, not analysed', async () => {
    await settle(api.auth.createGuestSession())

    const wallet = await settle(
      api.wallets.create({ chain_id: 'ethereum', address: HEALTHY }),
    )

    expect(wallet.sync.status).toBe(SyncStatus.PENDING)
    expect(wallet.sync.stages_completed).toEqual([])
  })

  it('refuses portfolio data until the initial sync has finished', async () => {
    await settle(api.auth.createGuestSession())
    const wallet = await settle(
      api.wallets.create({ chain_id: 'ethereum', address: HEALTHY }),
    )

    await expect(settle(api.portfolio.get(wallet.id))).rejects.toMatchObject({
      code: ApiErrorCode.WALLET_NOT_READY,
    })
  })

  it('reports real sync stages while the job runs', async () => {
    await settle(api.auth.createGuestSession())
    const created = await settle(
      api.wallets.create({ chain_id: 'ethereum', address: HEALTHY }),
    )

    await vi.advanceTimersByTimeAsync(3_000)
    const midway = await settle(api.wallets.get(created.id))

    expect(midway.sync.status).toBe(SyncStatus.SYNCING)
    expect(midway.sync.stage).not.toBeNull()
    expect(midway.sync.stages_completed.length).toBeGreaterThan(0)
    expect(midway.sync.stages_completed).not.toContain(midway.sync.stage)
  })

  it('exposes a fully valued portfolio once ready', async () => {
    const wallet = await analysedWallet()
    expect(wallet.sync.status).toBe(SyncStatus.READY)

    const portfolio = await settle(api.portfolio.get(wallet.id))

    expect(portfolio.valuation_status).toBe(ValuationStatus.COMPLETE)
    expect(portfolio.total_value_usd).not.toBeNull()
    expect(portfolio.positions.length).toBeGreaterThan(0)
  })

  it('leaves an unpriced position without a value instead of zero', async () => {
    const wallet = await analysedWallet(PARTIAL)
    const portfolio = await settle(api.portfolio.get(wallet.id))

    expect(portfolio.valuation_status).toBe(ValuationStatus.PARTIAL)
    expect(portfolio.exclusions.unpriced_positions).toBeGreaterThan(0)
    expect(portfolio.notices.map((notice) => notice.code)).toContain(
      DataNoticeCode.UNPRICED_ASSETS_EXCLUDED,
    )

    const unpriced = portfolio.positions.find(
      (position) =>
        position.visibility === AssetVisibility.VISIBLE &&
        position.valuation_status === ValuationStatus.UNAVAILABLE,
    )

    expect(unpriced).toBeDefined()
    expect(unpriced?.value_usd).toBeNull()
    expect(unpriced?.price?.value_usd ?? null).toBeNull()
    // The balance itself is still a fact and remains displayable.
    expect(Number(unpriced?.balance)).toBeGreaterThan(0)
  })

  it('surfaces a failed synchronisation as a domain error state', async () => {
    const wallet = await analysedWallet(FAILING)

    expect(wallet.sync.status).toBe(SyncStatus.FAILED)
    expect(wallet.sync.error?.code).toBe(ApiErrorCode.WALLET_SYNC_FAILED)

    await expect(settle(api.portfolio.get(wallet.id))).rejects.toBeInstanceOf(
      ApiError,
    )
  })

  it('paginates transactions by opaque cursor', async () => {
    const wallet = await analysedWallet()

    const first = await settle(api.transactions.list(wallet.id, { limit: 5 }))
    expect(first.items).toHaveLength(5)
    expect(first.has_more).toBe(true)
    expect(first.next_cursor).toBeTruthy()

    const second = await settle(
      api.transactions.list(wallet.id, { limit: 5, cursor: first.next_cursor }),
    )

    const firstIds = first.items.map((item) => item.id)
    expect(second.items.some((item) => firstIds.includes(item.id))).toBe(false)
  })

  it('returns a snapshot series for performance rather than a single number', async () => {
    const wallet = await analysedWallet()

    const performance = await settle(
      api.performance.get(wallet.id, PerformancePeriod.D7),
    )

    expect(performance.period).toBe(PerformancePeriod.D7)
    expect(performance.series.length).toBeGreaterThan(1)
    expect(performance.drivers.length).toBeGreaterThan(0)
  })

  it('streams a structured answer with tool activity', async () => {
    const wallet = await analysedWallet()
    const conversation = await settle(
      api.conversations.create({ wallet_id: wallet.id }),
    )

    const events = await collect(
      api.conversations.streamMessage(conversation.id, {
        content: 'Why is my portfolio down?',
      }),
    )

    const types = events.map((event) => event.type)
    expect(types).toContain('tool_started')
    expect(types).toContain('tool_completed')
    expect(types).toContain('text_delta')
    expect(types.at(-1)).toBe('completed')

    const completed = events.at(-1)
    if (completed?.type !== 'completed') throw new Error('expected completion')

    expect(completed.message.response?.answer).toBeTruthy()
    expect(completed.message.response?.claims.length).toBeGreaterThan(0)
    expect(completed.usage?.used).toBe(1)
  })

  it('refuses AI work once the daily limit is spent', async () => {
    const wallet = await analysedWallet()
    const conversation = await settle(
      api.conversations.create({ wallet_id: wallet.id }),
    )

    const usage = await settle(api.ai.getUsage())

    for (let index = 0; index < usage.limit; index += 1) {
      await collect(
        api.conversations.streamMessage(conversation.id, {
          content: 'Summarise my portfolio',
        }),
      )
    }

    const events = await collect(
      api.conversations.streamMessage(conversation.id, {
        content: 'Summarise my portfolio',
      }),
    )

    const last = events.at(-1)
    expect(last?.type).toBe('error')
    if (last?.type === 'error') {
      expect(last.error.code).toBe(ApiErrorCode.AI_DAILY_LIMIT_REACHED)
    }
  })

  /** Drains a stream, advancing the fake clock between events. */
  async function collect(
    stream: AsyncIterable<AIStreamEvent>,
  ): Promise<AIStreamEvent[]> {
    const events: AIStreamEvent[] = []
    const iterator = stream[Symbol.asyncIterator]()

    for (;;) {
      const next = iterator.next()
      await vi.advanceTimersByTimeAsync(500)
      const result = await next
      if (result.done) break
      events.push(result.value)
    }

    return events
  }
})
