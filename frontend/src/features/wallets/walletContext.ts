import { createContext, useContext } from 'react'
import type { Wallet } from '@/api/types'

/**
 * Active wallet contract.
 *
 * The MVP analyses one wallet at a time, but consumers read the active wallet
 * from here rather than assuming a singleton. Adding a wallet switcher later
 * means changing the selection, not the dashboard.
 */
export interface WalletContextValue {
  walletId: string
  wallet: Wallet | undefined
  isLoading: boolean
  error: unknown
  /** Portfolio data can exist (READY or PARTIAL). */
  isAnalysed: boolean
  isSyncing: boolean
  hasFailed: boolean
  refetch: () => void
}

export const WalletContext = createContext<WalletContextValue | null>(null)

export function useActiveWallet(): WalletContextValue {
  const context = useContext(WalletContext)
  if (!context) {
    throw new Error('useActiveWallet must be used inside WalletProvider')
  }
  return context
}
