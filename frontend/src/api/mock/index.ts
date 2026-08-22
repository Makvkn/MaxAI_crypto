import type { MaxAIApi } from '../contract'
import { ApiError } from '../errors'
import {
  AIIntent,
  AIToolCallStatus,
  ApiErrorCode,
  AuthMethod,
  DataQuality,
  MessageRole,
  MessageStatus,
  PerformancePeriod,
  SyncStatus,
  UserKind,
  type AIResponse,
  type AIStreamEvent,
  type AIUsage,
  type AuthSession,
  type ConversationMessage,
  type CursorPage,
  type Portfolio,
  type ScenarioResult,
  type Transaction,
  type Wallet,
} from '../types'
import { buildAiResponse, resolveIntent, toolsForIntent } from './data/ai'
import { buildPerformance } from './data/history'
import { buildPortfolio } from './data/portfolio'
import { buildTransactions } from './data/transactions'
import { createMockDb, type MockDb } from './db'
import * as d from './support/decimal'
import { delay } from './support/latency'
import { resolveVariant } from './variants'

/**
 * In-memory MaxAI backend.
 *
 * This adapter implements the exact same `MaxAIApi` contract as the HTTP
 * adapter, using the exact same DTO types. Nothing above `src/api` can tell
 * which one is active, which is what makes `VITE_API_MODE=real` a
 * configuration change rather than a refactor.
 *
 * It deliberately reproduces the awkward parts of the real system:
 * asynchronous wallet sync, cursor pagination, partial and stale data,
 * unpriced assets, domain errors, AI usage limits and SSE-shaped streaming.
 */

const PAGE_SIZE = 50
const AI_STREAM_CHUNK_MS = 26
const AI_TOOL_MS = 320

export function createMockApi(): MaxAIApi {
  const db = createMockDb()

  return {
    auth: createAuth(db),
    wallets: {
      async list(params, options) {
        await delay(160, options?.signal)
        return paginate(db.listWallets(), params?.cursor ?? null, params?.limit)
      },

      async get(walletId, options) {
        await delay(140, options?.signal)
        return requireWallet(db, walletId)
      },

      async create(request, options) {
        await delay(420, options?.signal)

        const address = request.address.trim()
        if (address.length < 20) {
          throw new ApiError({
            code: ApiErrorCode.INVALID_WALLET_ADDRESS,
            message: 'The wallet address is not valid for this chain.',
            status: 422,
            details: { fields: { address: 'INVALID_FORMAT' } },
          })
        }

        const entitlement =
          db.session?.user.subscription.entitlements.max_wallets ?? 1
        if (db.listWallets().length >= entitlement) {
          throw new ApiError({
            code: ApiErrorCode.WALLET_LIMIT_REACHED,
            message: 'Wallet limit reached for this plan.',
            status: 429,
            details: { limit: entitlement },
          })
        }

        // Returns immediately: the sync job has only been enqueued.
        return db.addWallet({
          chain_id: request.chain_id,
          address,
          label: request.label ?? null,
        })
      },
    },

    portfolio: {
      async get(walletId, options) {
        await delay(260, options?.signal)
        return resolvePortfolio(db, walletId)
      },
    },

    performance: {
      async get(walletId, period, options) {
        await delay(300, options?.signal)
        const wallet = requireReadyWallet(db, walletId)
        const variant = resolveVariant(wallet.address)
        const portfolio = resolvePortfolio(db, walletId)
        return buildPerformance(wallet, portfolio, period, variant, new Date())
      },
    },

    transactions: {
      async list(walletId, params, options) {
        await delay(280, options?.signal)
        const all = resolveTransactions(db, walletId).filter(
          (transaction) => !params?.type || transaction.type === params.type,
        )
        return paginate(all, params?.cursor ?? null, params?.limit)
      },

      async get(walletId, transactionId, options) {
        await delay(180, options?.signal)
        const found = resolveTransactions(db, walletId).find(
          (transaction) => transaction.id === transactionId,
        )
        if (!found) {
          throw new ApiError({
            code: ApiErrorCode.TRANSACTION_NOT_FOUND,
            message: 'Transaction not found.',
            status: 404,
          })
        }
        return found
      },
    },

    conversations: {
      async list(params, options) {
        await delay(150, options?.signal)
        return paginate(
          db.listConversations(params?.wallet_id),
          params?.cursor ?? null,
          params?.limit,
        )
      },

      async create(request, options) {
        await delay(200, options?.signal)
        requireWallet(db, request.wallet_id)
        return db.addConversation(
          request.wallet_id,
          request.title?.trim() || 'Portfolio analysis',
        )
      },

      async listMessages(conversationId, params, options) {
        await delay(180, options?.signal)
        if (!db.findConversation(conversationId)) {
          throw new ApiError({
            code: ApiErrorCode.CONVERSATION_NOT_FOUND,
            message: 'Conversation not found.',
            status: 404,
          })
        }
        // Newest first, so the cursor pages backwards through history.
        const ordered = [...db.listMessages(conversationId)].reverse()
        return paginate(ordered, params?.cursor ?? null, params?.limit)
      },

      streamMessage(conversationId, request, options) {
        return streamMessage(db, conversationId, request, options?.signal)
      },
    },

    ai: {
      async getUsage(options) {
        await delay(120, options?.signal)
        return usagePayload(db)
      },
    },

    scenarios: {
      async simulate(request, options) {
        await delay(520, options?.signal)
        return simulateScenario(db, request)
      },
    },
  }
}

/* -------------------------------------------------------------------------- */
/* Auth                                                                       */
/* -------------------------------------------------------------------------- */

function createAuth(db: MockDb): MaxAIApi['auth'] {
  const toSession = (session: ReturnType<MockDb['issueSession']>): AuthSession => ({
    access_token: session.access_token,
    refresh_token: session.refresh_token,
    expires_at: session.expires_at,
    user: session.user,
  })

  return {
    async createGuestSession(options) {
      await delay(260, options?.signal)
      const user = db.createUser(UserKind.GUEST, null, [AuthMethod.GUEST])
      return toSession(db.issueSession(user))
    },

    async registerWithEmail(credentials, options) {
      await delay(420, options?.signal)
      const user = upgradeOrCreate(db, credentials.email, AuthMethod.EMAIL)
      return toSession(db.issueSession(user))
    },

    async loginWithEmail(credentials, options) {
      await delay(420, options?.signal)
      if (credentials.password.length < 8) {
        throw new ApiError({
          code: ApiErrorCode.INVALID_CREDENTIALS,
          message: 'Invalid email or password.',
          status: 401,
        })
      }
      const user = upgradeOrCreate(db, credentials.email, AuthMethod.EMAIL)
      return toSession(db.issueSession(user))
    },

    async loginWithGoogle(_request, options) {
      await delay(380, options?.signal)
      const user = upgradeOrCreate(db, 'demo@google.com', AuthMethod.GOOGLE)
      return toSession(db.issueSession(user))
    },

    async upgradeAccount(request, options) {
      await delay(420, options?.signal)
      const email = request.method === 'EMAIL' ? request.email : 'demo@google.com'
      const method =
        request.method === 'EMAIL' ? AuthMethod.EMAIL : AuthMethod.GOOGLE
      const user = upgradeOrCreate(db, email, method)
      return toSession(db.issueSession(user))
    },

    async getCurrentUser(options) {
      await delay(120, options?.signal)
      const session = db.session
      if (!session) {
        throw new ApiError({
          code: ApiErrorCode.AUTHENTICATION_ERROR,
          message: 'No active session.',
          status: 401,
        })
      }
      return session.user
    },

    async logout(options) {
      await delay(120, options?.signal)
      db.setSession(null)
    },
  }
}

/** Keeps `user.id` stable so guest data survives the upgrade. */
function upgradeOrCreate(db: MockDb, email: string, method: AuthMethod) {
  const existing = db.session?.user
  if (existing) {
    return {
      ...existing,
      kind: UserKind.REGISTERED,
      email,
      display_name: email.split('@')[0] ?? null,
      auth_methods: Array.from(new Set([...existing.auth_methods, method])),
    }
  }
  return db.createUser(UserKind.REGISTERED, email, [method])
}

/* -------------------------------------------------------------------------- */
/* Wallet-derived data                                                        */
/* -------------------------------------------------------------------------- */

function requireWallet(db: MockDb, walletId: string): Wallet {
  const wallet = db.findWallet(walletId)
  if (!wallet) {
    throw new ApiError({
      code: ApiErrorCode.WALLET_NOT_FOUND,
      message: 'Wallet not found.',
      status: 404,
    })
  }
  return wallet
}

/** Portfolio data only exists once the initial sync has produced a snapshot. */
function requireReadyWallet(db: MockDb, walletId: string): Wallet {
  const wallet = requireWallet(db, walletId)

  if (
    wallet.sync.status === SyncStatus.PENDING ||
    wallet.sync.status === SyncStatus.SYNCING
  ) {
    throw new ApiError({
      code: ApiErrorCode.WALLET_NOT_READY,
      message: 'The wallet is still synchronising.',
      status: 409,
      details: { sync_status: wallet.sync.status },
    })
  }

  if (wallet.sync.status === SyncStatus.FAILED) {
    throw new ApiError({
      code: ApiErrorCode.WALLET_SYNC_FAILED,
      message: 'Wallet synchronisation failed.',
      status: 503,
    })
  }

  return wallet
}

function resolvePortfolio(db: MockDb, walletId: string): Portfolio {
  const wallet = requireReadyWallet(db, walletId)
  const variant = resolveVariant(wallet.address)

  if (variant.portfolioUnavailable) {
    throw new ApiError({
      code: ApiErrorCode.PORTFOLIO_DATA_TEMPORARILY_UNAVAILABLE,
      message: 'Portfolio data is temporarily unavailable.',
      status: 503,
    })
  }

  return buildPortfolio(wallet, variant, new Date()).portfolio
}

function resolveTransactions(db: MockDb, walletId: string): Transaction[] {
  const wallet = requireReadyWallet(db, walletId)
  const variant = resolveVariant(wallet.address)
  const { portfolio } = buildPortfolio(wallet, variant, new Date())
  return buildTransactions(wallet, portfolio.positions, new Date())
}

/* -------------------------------------------------------------------------- */
/* AI                                                                         */
/* -------------------------------------------------------------------------- */

function usagePayload(db: MockDb): AIUsage {
  const usage = db.usage()
  const resets = new Date()
  resets.setUTCHours(24, 0, 0, 0)

  return {
    date: usage.date,
    used: usage.used,
    limit: usage.limit,
    remaining: Math.max(0, usage.limit - usage.used),
    resets_at: resets.toISOString(),
    plan: db.session?.user.subscription.plan ?? 'FREE',
  }
}

function aiLimitError(db: MockDb): ApiError {
  const usage = usagePayload(db)
  return new ApiError({
    code: ApiErrorCode.AI_DAILY_LIMIT_REACHED,
    message: 'Daily AI operation limit reached.',
    status: 429,
    details: {
      limit: usage.limit,
      used: usage.used,
      remaining: 0,
      resets_at: usage.resets_at,
    },
  })
}

async function* streamMessage(
  db: MockDb,
  conversationId: string,
  request: { content: string; context?: { transaction_id?: string } },
  signal?: AbortSignal,
): AsyncGenerator<AIStreamEvent> {
  const conversation = db.findConversation(conversationId)
  if (!conversation) {
    yield {
      type: 'error',
      error: {
        code: ApiErrorCode.CONVERSATION_NOT_FOUND,
        message: 'Conversation not found.',
      },
    }
    return
  }

  const now = new Date()
  db.addMessage({
    id: db.nextId('msg'),
    conversation_id: conversationId,
    role: MessageRole.USER,
    status: MessageStatus.COMPLETED,
    content: request.content,
    response: null,
    tool_calls: [],
    error: null,
    created_at: now.toISOString(),
  })

  const budget = db.consumeUsage()
  if (!budget.allowed) {
    const error = aiLimitError(db)
    db.addMessage({
      id: db.nextId('msg'),
      conversation_id: conversationId,
      role: MessageRole.ASSISTANT,
      status: MessageStatus.FAILED,
      content: '',
      response: null,
      tool_calls: [],
      error: { code: error.code, message: error.message, details: error.details },
      created_at: new Date().toISOString(),
    })
    yield {
      type: 'error',
      error: { code: error.code, message: error.message, details: error.details },
    }
    return
  }

  const assistant = db.addMessage({
    id: db.nextId('msg'),
    conversation_id: conversationId,
    role: MessageRole.ASSISTANT,
    status: MessageStatus.STREAMING,
    content: '',
    response: null,
    tool_calls: [],
    error: null,
    created_at: new Date().toISOString(),
  })

  let response: AIResponse
  try {
    response = composeAnswer(db, conversation.wallet_id, request)
  } catch (error) {
    const body =
      error instanceof ApiError
        ? { code: error.code, message: error.message, details: error.details }
        : { code: ApiErrorCode.AI_UNAVAILABLE, message: 'AI is unavailable.' }
    db.updateMessage(conversationId, assistant.id, {
      status: MessageStatus.FAILED,
      error: body,
    })
    yield { type: 'error', error: body }
    return
  }

  const intent = response.intent
  const toolCalls = toolsForIntent(intent)

  for (const tool of toolCalls) {
    if (signal?.aborted) return
    const toolCallId = db.nextId('tc')
    yield { type: 'tool_started', tool_call_id: toolCallId, tool }
    await delay(AI_TOOL_MS, signal).catch(() => undefined)
    if (signal?.aborted) return
    yield { type: 'tool_completed', tool_call_id: toolCallId, tool, ok: true }
  }

  for (const chunk of chunkText(response.answer)) {
    if (signal?.aborted) return
    yield { type: 'text_delta', text: chunk }
    await delay(AI_STREAM_CHUNK_MS, signal).catch(() => undefined)
  }

  if (signal?.aborted) return

  const completed = db.updateMessage(conversationId, assistant.id, {
    status: MessageStatus.COMPLETED,
    content: response.answer,
    response,
    tool_calls: toolCalls.map((tool) => ({
      id: db.nextId('tc'),
      tool,
      status: AIToolCallStatus.COMPLETED,
      started_at: new Date().toISOString(),
      completed_at: new Date().toISOString(),
    })),
  })

  yield {
    type: 'completed',
    message: completed as ConversationMessage,
    usage: usagePayload(db),
  }
}

function composeAnswer(
  db: MockDb,
  walletId: string,
  request: { content: string; context?: { transaction_id?: string } },
): AIResponse {
  const wallet = requireWallet(db, walletId)
  const variant = resolveVariant(wallet.address)

  // An unsupported question needs no portfolio context at all.
  const preliminaryIntent = resolveIntent(request.content, {
    hasTransaction: Boolean(request.context?.transaction_id),
    hasScenario: false,
  })

  const portfolio = buildPortfolio(wallet, variant, new Date()).portfolio
  const performance = buildPerformance(
    wallet,
    portfolio,
    PerformancePeriod.H24,
    variant,
    new Date(),
  )

  if (preliminaryIntent === AIIntent.UNSUPPORTED) {
    return buildAiResponse({
      question: request.content,
      portfolio,
      performance,
      variant,
    })
  }

  const transaction = request.context?.transaction_id
    ? (buildTransactions(wallet, portfolio.positions, new Date()).find(
        (entry) => entry.id === request.context?.transaction_id,
      ) ?? null)
    : null

  return buildAiResponse({
    question: request.content,
    portfolio,
    performance,
    transaction,
    variant,
  })
}

/** Splits an answer into token-sized chunks, as a model would emit them. */
function chunkText(text: string): string[] {
  const words = text.split(/(\s+)/)
  const chunks: string[] = []
  let current = ''

  for (const word of words) {
    current += word
    if (current.length >= 12) {
      chunks.push(current)
      current = ''
    }
  }
  if (current) chunks.push(current)
  return chunks
}

/* -------------------------------------------------------------------------- */
/* Scenarios                                                                  */
/* -------------------------------------------------------------------------- */

function simulateScenario(
  db: MockDb,
  request: { wallet_id: string; asset_id: string; change_pct: string; type: string },
): ScenarioResult {
  const wallet = requireReadyWallet(db, request.wallet_id)
  const variant = resolveVariant(wallet.address)
  const { portfolio } = buildPortfolio(wallet, variant, new Date())

  const position = portfolio.positions.find(
    (entry) => entry.asset.id === request.asset_id,
  )

  if (!position) {
    throw new ApiError({
      code: ApiErrorCode.VALIDATION_ERROR,
      message: 'The wallet does not hold this asset.',
      status: 422,
      details: { fields: { asset_id: 'NOT_HELD' } },
    })
  }

  if (position.value_usd === null || portfolio.total_value_usd === null) {
    throw new ApiError({
      code: ApiErrorCode.PRICE_DATA_UNAVAILABLE,
      message: 'This asset cannot be simulated because its price is unknown.',
      status: 503,
    })
  }

  const budget = db.consumeUsage()
  if (!budget.allowed) throw aiLimitError(db)

  const projectedAssetValue = d.applyPercent(
    position.value_usd,
    request.change_pct,
    2,
  )
  const assetImpact = d.subtract(projectedAssetValue, position.value_usd, 2)
  const projectedPortfolio = d.add([portfolio.total_value_usd, assetImpact], 2)

  const result: ScenarioResult = {
    id: db.nextId('scn'),
    wallet_id: wallet.id,
    type: 'ASSET_PRICE_CHANGE',
    currency: 'USD',
    asset: position.asset,
    change_pct: request.change_pct,
    baseline: {
      portfolio_value_usd: portfolio.total_value_usd,
      asset_value_usd: position.value_usd,
      asset_allocation_pct: position.allocation_pct,
    },
    projection: {
      portfolio_value_usd: projectedPortfolio,
      asset_value_usd: projectedAssetValue,
      asset_impact_usd: assetImpact,
      portfolio_change_usd: assetImpact,
      portfolio_change_pct: d.percentOf(
        assetImpact,
        portfolio.total_value_usd,
        2,
      ),
    },
    data_quality:
      portfolio.data_quality === DataQuality.COMPLETE
        ? DataQuality.COMPLETE
        : portfolio.data_quality,
    calculation_id: `calc_scenario_${wallet.id}_${position.asset.symbol}`,
    calculation_version: portfolio.calculation_version,
    created_at: new Date().toISOString(),
    explanation: null,
  }

  const performance = buildPerformance(
    wallet,
    portfolio,
    PerformancePeriod.H24,
    variant,
    new Date(),
  )

  return {
    ...result,
    explanation: buildAiResponse({
      question: `What if ${position.asset.symbol} moves ${request.change_pct}%?`,
      portfolio,
      performance,
      scenario: result,
      variant,
    }),
  }
}

/* -------------------------------------------------------------------------- */
/* Cursor pagination                                                          */
/* -------------------------------------------------------------------------- */

/**
 * Opaque cursors. The frontend must never decode or construct these — the
 * encoding exists only so the mock can resume a page.
 */
function encodeCursor(offset: number): string {
  return btoa(`offset:${offset}`)
}

function decodeCursor(cursor: string | null): number {
  if (!cursor) return 0
  try {
    const decoded = atob(cursor)
    const value = Number(decoded.replace('offset:', ''))
    return Number.isFinite(value) && value >= 0 ? value : 0
  } catch {
    return 0
  }
}

function paginate<T>(
  items: T[],
  cursor: string | null,
  limit = PAGE_SIZE,
): CursorPage<T> {
  const offset = decodeCursor(cursor)
  const size = Math.min(Math.max(limit, 1), 100)
  const slice = items.slice(offset, offset + size)
  const nextOffset = offset + slice.length
  const hasMore = nextOffset < items.length

  return {
    items: slice,
    next_cursor: hasMore ? encodeCursor(nextOffset) : null,
    has_more: hasMore,
  }
}
