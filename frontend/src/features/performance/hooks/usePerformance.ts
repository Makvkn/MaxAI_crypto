import { useQuery } from '@tanstack/react-query'
import { api } from '@/api'
import type { PerformancePeriod } from '@/api/types'
import { queryKeys } from '@/lib/query/queryKeys'
import { useProtectedQueryEnabled } from '@/features/auth/useProtectedQueryEnabled'

/**
 * Snapshot-based performance for a period.
 *
 * The response carries both the period result and the snapshot series used by
 * the chart, so no history is reconstructed client-side.
 */
export function usePerformance(
  walletId: string,
  period: PerformancePeriod,
  options?: { enabled?: boolean },
) {
  const protectedEnabled = useProtectedQueryEnabled(options?.enabled ?? true)

  return useQuery({
    queryKey: queryKeys.performance(walletId, period),
    queryFn: ({ signal }) => api.performance.get(walletId, period, { signal }),
    enabled: protectedEnabled,
    staleTime: 60_000,
  })
}
