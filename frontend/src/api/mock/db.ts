import {
  ApiErrorCode,
  AuthMethod,
  DataFreshness,
  SubscriptionPlan,
  SubscriptionStatus,
  SyncStage,
  SyncStatus,
  UserKind,
  WalletStatus,
  type Conversation,
  type ConversationMessage,
  type User,
  type Wallet,
  type WalletSyncState,
} from '../types'
import { resolveVariant } from './variants'

/**
 * Mock persistence.
 *
 * State lives in localStorage so a page reload behaves like a returning user:
 * the wallet, its conversations and the daily AI usage all survive. Wallet sync
 * state is *derived* from `created_at` rather than stored, which is what makes
 * the asynchronous job observable without any client-side fake progress.
 */

const STORAGE_KEY = 'maxai.mock.db.v1'

export interface MockSession {
  access_token: string
  refresh_token: string
  expires_at: string
  user: User
}

interface MockState {
  session: MockSession | null
  /** Wallets as created. `sync` is recomputed on read. */
  wallets: Wallet[]
  conversations: Conversation[]
  messages: Record<string, ConversationMessage[]>
  usage: { date: string; used: number }
  sequence: number
}

const FREE_ENTITLEMENTS = {
  max_wallets: 1,
  ai_operations_per_day: 10,
  features: ['portfolio', 'performance', 'transactions', 'ai_insight'],
}

function emptyState(): MockState {
  return {
    session: null,
    wallets: [],
    conversations: [],
    messages: {},
    usage: { date: today(), used: 0 },
    sequence: 1,
  }
}

function today(): string {
  return new Date().toISOString().slice(0, 10)
}

export function createMockDb() {
  let state = load()

  function load(): MockState {
    try {
      const raw = globalThis.localStorage?.getItem(STORAGE_KEY)
      if (!raw) return emptyState()
      const parsed = JSON.parse(raw) as MockState
      return { ...emptyState(), ...parsed }
    } catch {
      return emptyState()
    }
  }

  function persist(): void {
    try {
      globalThis.localStorage?.setItem(STORAGE_KEY, JSON.stringify(state))
    } catch {
      // Non-fatal: the mock keeps working in memory.
    }
  }

  function nextId(prefix: string): string {
    state.sequence += 1
    persist()
    return `${prefix}_${state.sequence.toString(36)}${Date.now().toString(36).slice(-4)}`
  }

  return {
    get session(): MockSession | null {
      return state.session
    },

    setSession(session: MockSession | null): void {
      state.session = session
      persist()
    },

    createUser(kind: UserKind, email: string | null, methods: AuthMethod[]): User {
      return {
        id: nextId('usr'),
        kind,
        email,
        display_name: email ? (email.split('@')[0] ?? null) : null,
        auth_methods: methods,
        subscription: {
          plan: SubscriptionPlan.FREE,
          status: SubscriptionStatus.ACTIVE,
          current_period_end: null,
          entitlements: FREE_ENTITLEMENTS,
        },
        created_at: new Date().toISOString(),
      }
    },

    issueSession(user: User): MockSession {
      const session: MockSession = {
        access_token: `mock.access.${nextId('tok')}`,
        refresh_token: `mock.refresh.${nextId('tok')}`,
        expires_at: new Date(Date.now() + 15 * 60_000).toISOString(),
        user,
      }
      state.session = session
      persist()
      return session
    },

    listWallets(): Wallet[] {
      const now = Date.now()
      return state.wallets
        .filter((wallet) => wallet.status !== WalletStatus.DELETED)
        .map((wallet) => withSyncState(wallet, now))
    },

    findWallet(walletId: string): Wallet | null {
      const wallet = state.wallets.find((entry) => entry.id === walletId)
      return wallet ? withSyncState(wallet, Date.now()) : null
    },

    addWallet(input: { chain_id: string; address: string; label: string | null }): Wallet {
      const createdAt = new Date().toISOString()
      const wallet: Wallet = {
        id: nextId('wlt'),
        chain_id: input.chain_id,
        address: input.address,
        label: input.label,
        status: WalletStatus.SYNCING,
        sync: {
          status: SyncStatus.PENDING,
          stage: null,
          stages_completed: [],
          started_at: createdAt,
          completed_at: null,
          last_synced_at: null,
          data_freshness: null,
          error: null,
        },
        created_at: createdAt,
        updated_at: createdAt,
      }
      state.wallets.push(wallet)
      persist()
      return withSyncState(wallet, Date.now())
    },

    listConversations(walletId?: string): Conversation[] {
      return state.conversations
        .filter((conversation) => !walletId || conversation.wallet_id === walletId)
        .sort((a, b) => b.updated_at.localeCompare(a.updated_at))
    },

    findConversation(conversationId: string): Conversation | null {
      return (
        state.conversations.find((entry) => entry.id === conversationId) ?? null
      )
    },

    addConversation(walletId: string, title: string): Conversation {
      const timestamp = new Date().toISOString()
      const conversation: Conversation = {
        id: nextId('cnv'),
        wallet_id: walletId,
        title,
        message_count: 0,
        last_message_preview: null,
        created_at: timestamp,
        updated_at: timestamp,
      }
      state.conversations.unshift(conversation)
      state.messages[conversation.id] = []
      persist()
      return conversation
    },

    listMessages(conversationId: string): ConversationMessage[] {
      return state.messages[conversationId] ?? []
    },

    addMessage(message: ConversationMessage): ConversationMessage {
      const bucket = state.messages[message.conversation_id] ?? []
      bucket.push(message)
      state.messages[message.conversation_id] = bucket
      touchConversation(message)
      persist()
      return message
    },

    updateMessage(
      conversationId: string,
      messageId: string,
      patch: Partial<ConversationMessage>,
    ): ConversationMessage | null {
      const bucket = state.messages[conversationId]
      if (!bucket) return null
      const index = bucket.findIndex((message) => message.id === messageId)
      if (index === -1) return null

      const updated = { ...(bucket[index] as ConversationMessage), ...patch }
      bucket[index] = updated
      touchConversation(updated)
      persist()
      return updated
    },

    nextId,

    /* --- AI usage ------------------------------------------------------- */

    usage,
    consumeUsage,

    reset(): void {
      state = emptyState()
      persist()
    },
  }

  /** Daily budget, rolled over at UTC midnight. */
  function usage(): { date: string; used: number; limit: number } {
    if (state.usage.date !== today()) {
      state.usage = { date: today(), used: 0 }
      persist()
    }
    return {
      ...state.usage,
      limit:
        state.session?.user.subscription.entitlements.ai_operations_per_day ??
        FREE_ENTITLEMENTS.ai_operations_per_day,
    }
  }

  function consumeUsage(): { allowed: boolean; used: number; limit: number } {
    const current = usage()
    if (current.used >= current.limit) {
      return { allowed: false, used: current.used, limit: current.limit }
    }
    state.usage = { date: current.date, used: current.used + 1 }
    persist()
    return { allowed: true, used: state.usage.used, limit: current.limit }
  }

  function touchConversation(message: ConversationMessage): void {
    const conversation = state.conversations.find(
      (entry) => entry.id === message.conversation_id,
    )
    if (!conversation) return
    conversation.message_count = (state.messages[conversation.id] ?? []).length
    conversation.last_message_preview = message.content.slice(0, 140)
    conversation.updated_at = new Date().toISOString()
  }
}

export type MockDb = ReturnType<typeof createMockDb>

/* -------------------------------------------------------------------------- */
/* Synchronisation state machine                                              */
/* -------------------------------------------------------------------------- */

const STAGE_ORDER: SyncStage[] = [
  SyncStage.FETCHING_BALANCES,
  SyncStage.FETCHING_TRANSACTIONS,
  SyncStage.NORMALIZING_ASSETS,
  SyncStage.FETCHING_PRICES,
  SyncStage.CALCULATING_PORTFOLIO,
  SyncStage.PREPARING_ANALYSIS,
]

const QUEUE_MS = 900
const STAGE_MS = 1_150

/**
 * Derives `wallet.sync` from elapsed time, the way polling a real job queue
 * would: the client sees whatever stage the worker has actually reached.
 */
export function withSyncState(wallet: Wallet, now: number): Wallet {
  const variant = resolveVariant(wallet.address)
  const startedAt = new Date(wallet.created_at).getTime()
  const speed = variant.slowSync ? 3 : 1
  const elapsed = now - startedAt

  const queue = QUEUE_MS * speed
  const stageMs = STAGE_MS * speed
  const total = queue + STAGE_MS * STAGE_ORDER.length * speed

  if (elapsed < queue) {
    return applySync(wallet, {
      status: SyncStatus.PENDING,
      stage: null,
      stages_completed: [],
      started_at: wallet.created_at,
      completed_at: null,
      last_synced_at: null,
      data_freshness: null,
      error: null,
    })
  }

  if (elapsed < total) {
    const stageIndex = Math.min(
      Math.floor((elapsed - queue) / stageMs),
      STAGE_ORDER.length - 1,
    )
    return applySync(wallet, {
      status: SyncStatus.SYNCING,
      stage: STAGE_ORDER[stageIndex] as SyncStage,
      stages_completed: STAGE_ORDER.slice(0, stageIndex),
      started_at: wallet.created_at,
      completed_at: null,
      last_synced_at: null,
      data_freshness: null,
      error: null,
    })
  }

  const completedAt = new Date(startedAt + total).toISOString()

  if (variant.syncFails) {
    return applySync(
      {
        ...wallet,
        status: WalletStatus.ERROR,
      },
      {
        status: SyncStatus.FAILED,
        stage: null,
        stages_completed: STAGE_ORDER.slice(0, 2),
        started_at: wallet.created_at,
        completed_at: completedAt,
        last_synced_at: null,
        data_freshness: null,
        error: {
          code: ApiErrorCode.WALLET_SYNC_FAILED,
          message: 'Wallet synchronisation failed.',
        },
      },
    )
  }

  const freshness = variant.veryStale
    ? DataFreshness.VERY_STALE
    : variant.stale
      ? DataFreshness.STALE
      : DataFreshness.FRESH

  return applySync(
    { ...wallet, status: WalletStatus.ACTIVE },
    {
      status: variant.syncPartial ? SyncStatus.PARTIAL : SyncStatus.READY,
      stage: null,
      stages_completed: STAGE_ORDER,
      started_at: wallet.created_at,
      completed_at: completedAt,
      last_synced_at: completedAt,
      data_freshness: freshness,
      error: variant.syncPartial
        ? {
            code: ApiErrorCode.PROVIDER_ERROR,
            message: 'Some data could not be retrieved during synchronisation.',
          }
        : null,
    },
  )
}

function applySync(wallet: Wallet, sync: WalletSyncState): Wallet {
  return { ...wallet, sync }
}
