import {
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query'
import { api } from '@/api'
import {
  SyncStatus,
  type CreateWalletRequest,
  type Wallet,
} from '@/api/types'
import { queryKeys } from '@/lib/query/queryKeys'
import { analytics } from '@/lib/analytics/analytics'
import { useSession } from '@/features/auth/sessionContext'
import { useProtectedQueryEnabled } from '@/features/auth/useProtectedQueryEnabled'

/**
 * Wallet server state.
 *
 * The list endpoint is cursor-paginated even though the MVP shows one wallet:
 * the contract is already plural, so multi-wallet support is a UI change only.
 */
export function useWallets(options?: { enabled?: boolean }) {
  const protectedEnabled = useProtectedQueryEnabled(options?.enabled ?? true)

  return useQuery({
    queryKey: queryKeys.wallets(),
    queryFn: ({ signal }) => api.wallets.list({ limit: 20 }, { signal }),
    enabled: protectedEnabled,
    select: (page) => page.items,
  })
}

/** True while the wallet has not finished its initial synchronisation. */
export function isSyncInProgress(wallet: Wallet | undefined | null): boolean {
  return (
    wallet?.sync.status === SyncStatus.PENDING ||
    wallet?.sync.status === SyncStatus.SYNCING
  )
}

/**
 * A single wallet.
 *
 * While the backend job is running the query polls, which is how the UI follows
 * real progress. When the backend later exposes push updates, only this hook
 * changes.
 */
export function useWallet(
  walletId: string | null,
  options?: { enabled?: boolean },
) {
  const extraEnabled = options?.enabled ?? true
  const protectedEnabled = useProtectedQueryEnabled(
    Boolean(walletId) && extraEnabled,
  )

  return useQuery({
    queryKey: queryKeys.wallet(walletId ?? 'none'),
    queryFn: ({ signal }) => api.wallets.get(walletId as string, { signal }),
    enabled: protectedEnabled,
    refetchInterval: (query) =>
      isSyncInProgress(query.state.data) ? 1_200 : false,
    staleTime: 0,
  })
}

/**
 * Creates a wallet.
 *
 * The mutation resolves as soon as the wallet row exists — the heavy
 * synchronisation runs in a backend job, so callers must follow `wallet.sync`
 * rather than treating this as "portfolio is ready".
 */
export function useCreateWallet() {
  const queryClient = useQueryClient()
  const { ensureSession } = useSession()

  return useMutation({
    mutationFn: async (request: CreateWalletRequest) => {
      // A wallet always belongs to a user, so a guest account is created first.
      await ensureSession()
      return api.wallets.create(request)
    },
    onSuccess: (wallet) => {
      queryClient.setQueryData(queryKeys.wallet(wallet.id), wallet)
      void queryClient.invalidateQueries({ queryKey: queryKeys.wallets() })
      analytics.track('analysis_started', { chain_id: wallet.chain_id })
    },
  })
}
