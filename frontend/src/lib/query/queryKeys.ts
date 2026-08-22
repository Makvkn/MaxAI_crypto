import type { PerformancePeriod, TransactionType } from '@/api/types'

/**
 * Every query key in the application.
 *
 * Centralised so invalidation is precise and no key string is duplicated.
 * Keys are hierarchical: invalidating `queryKeys.wallet(id)` also invalidates
 * that wallet's portfolio, performance and transactions.
 */
export const queryKeys = {
  session: () => ['session'] as const,

  aiUsage: () => ['ai', 'usage'] as const,

  wallets: () => ['wallets'] as const,
  wallet: (walletId: string) => ['wallets', walletId] as const,

  portfolio: (walletId: string) => ['wallets', walletId, 'portfolio'] as const,

  /** Prefix covering every period, for invalidation. */
  performanceRoot: (walletId: string) =>
    ['wallets', walletId, 'performance'] as const,

  performance: (walletId: string, period: PerformancePeriod) =>
    ['wallets', walletId, 'performance', period] as const,

  /** Prefix covering every type filter, for invalidation. */
  transactionsRoot: (walletId: string) =>
    ['wallets', walletId, 'transactions'] as const,

  transactions: (walletId: string, filters?: { type?: TransactionType }) =>
    ['wallets', walletId, 'transactions', filters?.type ?? 'all'] as const,

  transaction: (walletId: string, transactionId: string) =>
    ['wallets', walletId, 'transactions', 'detail', transactionId] as const,

  conversations: (walletId?: string) =>
    ['ai', 'conversations', walletId ?? 'all'] as const,

  messages: (conversationId: string) =>
    ['ai', 'conversations', conversationId, 'messages'] as const,
} as const

/** Keys whose data becomes invalid after a wallet finishes synchronising. */
export function walletScopedKeys(walletId: string) {
  return [queryKeys.wallet(walletId), queryKeys.portfolio(walletId)]
}
