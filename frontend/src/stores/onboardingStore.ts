import { create } from 'zustand'
import type { ChainId } from '@/api/types'

/**
 * Onboarding draft state.
 *
 * Temporary form state spanning two steps (network, then address). Cleared once
 * the wallet is created — the created wallet itself is server state.
 */
interface OnboardingState {
  chainId: ChainId | null
  address: string
  setChain: (chainId: ChainId) => void
  setAddress: (address: string) => void
  reset: () => void
}

export const useOnboardingStore = create<OnboardingState>((set) => ({
  chainId: null,
  address: '',
  setChain: (chainId) => set({ chainId }),
  setAddress: (address) => set({ address }),
  reset: () => set({ chainId: null, address: '' }),
}))
