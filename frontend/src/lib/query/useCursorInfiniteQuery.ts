import { useInfiniteQuery, type QueryKey } from '@tanstack/react-query'
import { useMemo } from 'react'
import type { Cursor, CursorPage } from '@/api/types'

/**
 * Reusable cursor pagination.
 *
 * The backend paginates by opaque cursor only — there is no page or offset
 * parameter anywhere in the API. This hook wraps TanStack Query's infinite
 * query so every paginated list (transactions, conversations, messages) behaves
 * identically and never invents page numbers.
 */
export interface CursorQueryResult<T> {
  items: T[]
  pageCount: number
  hasNextPage: boolean
  isLoading: boolean
  isFetching: boolean
  isFetchingNextPage: boolean
  error: unknown
  fetchNextPage: () => void
  refetch: () => void
}

export function useCursorInfiniteQuery<T>(params: {
  queryKey: QueryKey
  fetchPage: (args: {
    cursor: Cursor | null
    signal: AbortSignal
  }) => Promise<CursorPage<T>>
  limit?: number
  enabled?: boolean
  staleTime?: number
}): CursorQueryResult<T> {
  const query = useInfiniteQuery({
    queryKey: params.queryKey,
    queryFn: ({ pageParam, signal }) =>
      params.fetchPage({ cursor: pageParam, signal }),
    initialPageParam: null as Cursor | null,
    getNextPageParam: (lastPage) =>
      lastPage.has_more ? lastPage.next_cursor : null,
    enabled: params.enabled ?? true,
    staleTime: params.staleTime ?? 30_000,
  })

  const items = useMemo(
    () => query.data?.pages.flatMap((page) => page.items) ?? [],
    [query.data],
  )

  return {
    items,
    pageCount: query.data?.pages.length ?? 0,
    hasNextPage: query.hasNextPage,
    isLoading: query.isLoading,
    isFetching: query.isFetching,
    isFetchingNextPage: query.isFetchingNextPage,
    error: query.error,
    fetchNextPage: () => {
      if (query.hasNextPage && !query.isFetchingNextPage) {
        void query.fetchNextPage()
      }
    },
    refetch: () => void query.refetch(),
  }
}
