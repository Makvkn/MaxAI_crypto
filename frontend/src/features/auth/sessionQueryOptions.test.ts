import { describe, expect, it, vi } from 'vitest'
import { QueryClient } from '@tanstack/react-query'
import * as authGate from '@/api/authGate'
import type { User } from '@/api/types'
import { queryKeys } from '@/lib/query/queryKeys'
import { runSessionBootstrap } from './sessionQueryOptions'

describe('runSessionBootstrap', () => {
  it('fetches with staleTime 0 even when cached user exists', async () => {
    const staleUser = { id: 'stale-user' } as User
    const freshUser = { id: 'fresh-user' } as User

    const queryClient = new QueryClient()
    queryClient.setQueryData(queryKeys.session(), staleUser)

    vi.spyOn(authGate, 'runAuthBootstrap').mockImplementation(async (task) =>
      task(),
    )
    const fetchSpy = vi
      .spyOn(queryClient, 'fetchQuery')
      .mockResolvedValue(freshUser)

    await expect(runSessionBootstrap(queryClient)).resolves.toEqual(freshUser)

    expect(fetchSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        staleTime: 0,
      }),
    )
  })
})
