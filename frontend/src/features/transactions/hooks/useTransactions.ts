import { useQuery } from '@tanstack/react-query'
import { apiGetTransaction, apiGetTransactions } from '@/api'
import type { Transaction, TransactionType } from '@/api/types'
import { queryKeys } from '@/lib/query/queryKeys'
import { useCursorInfiniteQuery } from '@/lib/query/useCursorInfiniteQuery'
import { useProtectedQueryEnabled } from '@/features/auth/useProtectedQueryEnabled'

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
  const protectedEnabled = useProtectedQueryEnabled(params?.enabled ?? true)

  return useCursorInfiniteQuery<Transaction>({
    queryKey: queryKeys.transactions(walletId, { type: params?.type }),
    enabled: protectedEnabled,
    fetchPage: ({ cursor, signal }) =>
      apiGetTransactions(
        { walletId },
        { cursor, limit, type: params?.type },
        { signal },
      ),
  })
}

export function useTransaction(
  walletId: string,
  transactionId: string | null,
) {
  const protectedEnabled = useProtectedQueryEnabled(Boolean(transactionId))

  return useQuery({
    queryKey: queryKeys.transaction(walletId, transactionId ?? 'none'),
    queryFn: ({ signal }) =>
      apiGetTransaction(
        { walletId, transactionId: transactionId as string },
        { signal },
      ),
    enabled: protectedEnabled,
    staleTime: 5 * 60_000,
  })
}
