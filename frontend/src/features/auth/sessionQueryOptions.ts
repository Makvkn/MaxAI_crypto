import type { QueryClient } from '@tanstack/react-query'
import { api } from '@/api'
import { runAuthBootstrap } from '@/api/authGate'
import type { User } from '@/api/types'
import { queryKeys } from '@/lib/query/queryKeys'

const SESSION_STALE_TIME_MS = 5 * 60_000

/** Shared TanStack Query config for `/auth/session`. */
export function sessionQueryOptions() {
  return {
    queryKey: queryKeys.session(),
    queryFn: ({ signal }: { signal: AbortSignal }) =>
      api.auth.initializeSession({ signal }),
    staleTime: SESSION_STALE_TIME_MS,
    retry: false as const,
  }
}

/**
 * Validates the stored session as a single bootstrap attempt.
 *
 * staleTime: 0 forces a network round-trip even when cached user data exists.
 * AuthStatus transitions only when this fetchQuery promise settles.
 */
export async function runSessionBootstrap(
  queryClient: QueryClient,
): Promise<User> {
  return runAuthBootstrap(() =>
    queryClient.fetchQuery({
      ...sessionQueryOptions(),
      staleTime: 0,
    }),
  )
}
