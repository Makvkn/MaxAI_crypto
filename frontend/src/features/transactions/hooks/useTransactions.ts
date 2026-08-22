import { useQuery } from '@tanstack/react-query'
import { api } from '@/api'
import type { Transaction, TransactionType } from '@/api/types'
import { queryKeys } from '@/lib/query/queryKeys'
import { useCursorInfiniteQuery } from '@/lib/query/useCursorInfiniteQuery'

/**
 * Transaction history.
 *
 * Cursor-paginated, newest first. Filtering by type is a server-side parameter:
 * the backend decides what a type means, so the client does not re-classify.
 */
export function useTransactions(
  walletId: string,
  params?: { type?: TransactionType; limit?: number; enabled?: boolean },
) {
  const limit = params?.limit ?? 25

  return useCursorInfiniteQuery<Transaction>({
    queryKey: queryKeys.transactions(walletId, { type: params?.type }),
    enabled: params?.enabled ?? true,
    fetchPage: ({ cursor, signal }) =>
      api.transactions.list(
        walletId,
        { cursor, limit, type: params?.type },
        { signal },
      ),
  })
}

export function useTransaction(
  walletId: string,
  transactionId: string | null,
) {
  return useQuery({
    queryKey: queryKeys.transaction(walletId, transactionId ?? 'none'),
    queryFn: ({ signal }) =>
      api.transactions.get(walletId, transactionId as string, { signal }),
    enabled: Boolean(transactionId),
    staleTime: 5 * 60_000,
  })
}
