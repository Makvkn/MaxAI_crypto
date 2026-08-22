import { useEffect, useMemo, type ReactNode } from 'react'
import { SyncStatus } from '@/api/types'
import { usePreferencesStore } from '@/stores/preferencesStore'
import { isSyncInProgress, useWallet } from './hooks/useWallets'
import { WalletContext, type WalletContextValue } from './walletContext'

/** Resolves the wallet named by the route and shares it with the dashboard. */
export function WalletProvider({
  walletId,
  children,
}: {
  walletId: string
  children: ReactNode
}) {
  const selectWallet = usePreferencesStore((state) => state.selectWallet)
  const query = useWallet(walletId)

  const value = useMemo<WalletContextValue>(() => {
    const wallet = query.data
    return {
      walletId,
      wallet,
      isLoading: query.isLoading,
      error: query.error,
      isAnalysed:
        wallet?.sync.status === SyncStatus.READY ||
        wallet?.sync.status === SyncStatus.PARTIAL,
      isSyncing: isSyncInProgress(wallet),
      hasFailed: wallet?.sync.status === SyncStatus.FAILED,
      refetch: () => void query.refetch(),
    }
  }, [query, walletId])

  // Keep the persisted selection aligned with the route.
  useEffect(() => {
    if (usePreferencesStore.getState().selectedWalletId !== walletId) {
      selectWallet(walletId)
    }
  }, [selectWallet, walletId])

  return (
    <WalletContext.Provider value={value}>{children}</WalletContext.Provider>
  )
}
