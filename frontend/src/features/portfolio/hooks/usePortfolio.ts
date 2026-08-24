import { useCallback } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '@/api'
import { queryKeys } from '@/lib/query/queryKeys'
import { useProtectedQueryEnabled } from '@/features/auth/useProtectedQueryEnabled'

/**
 * Portfolio server state.
 *
 * The frontend reads valuation, allocation and change figures; it never derives
 * them. `enabled` is driven by wallet sync state so the app does not ask for a
 * portfolio that cannot exist yet.
 */
export function usePortfolio(walletId: string, options?: { enabled?: boolean }) {
  const protectedEnabled = useProtectedQueryEnabled(options?.enabled ?? true)

  return useQuery({
    queryKey: queryKeys.portfolio(walletId),
    queryFn: ({ signal }) => api.portfolio.get(walletId, { signal }),
    enabled: protectedEnabled,
    staleTime: 60_000,
  })
}

/**
 * Manual refresh.
 *
 * The MVP has no realtime updates: background synchronisation plus an explicit
 * refresh is the contract. This invalidates the wallet and everything derived
 * from it rather than mutating anything locally.
 */
export function useRefreshWalletData(walletId: string) {
  const queryClient = useQueryClient()

  return useCallback(async () => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: queryKeys.wallet(walletId) }),
      queryClient.invalidateQueries({ queryKey: queryKeys.portfolio(walletId) }),
      queryClient.invalidateQueries({
        queryKey: queryKeys.performanceRoot(walletId),
      }),
      queryClient.invalidateQueries({
        queryKey: queryKeys.transactionsRoot(walletId),
      }),
    ])
  }, [queryClient, walletId])
}
