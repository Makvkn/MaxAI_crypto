import { QueryClient } from '@tanstack/react-query'
import { isRetryableError } from '@/api/errors'

/**
 * TanStack Query owns all server state.
 *
 * Defaults reflect a financial dashboard: data is refetched deliberately
 * rather than aggressively, and only transport/internal failures are retried —
 * a `DATA_UNAVAILABLE` or `RATE_LIMIT` answer is information, not a fluke.
 */
export function createQueryClient(): QueryClient {
  return new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: 30_000,
        gcTime: 5 * 60_000,
        retry: (failureCount, error) =>
          failureCount < 2 && isRetryableError(error),
        retryDelay: (attempt) => Math.min(1000 * 2 ** attempt, 8_000),
        refetchOnWindowFocus: false,
        refetchOnReconnect: true,
      },
      mutations: {
        retry: false,
      },
    },
  })
}
