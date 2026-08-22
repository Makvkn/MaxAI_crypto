import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { PerformancePeriod } from '@/api/types'

/**
 * Client-only preferences, persisted across sessions.
 *
 * `selectedWalletId` is a *selection*, not wallet data: the wallet entity is
 * always read from TanStack Query. Keeping the selection here is what allows
 * multi-wallet support later without reworking the dashboard.
 */
interface PreferencesState {
  selectedWalletId: string | null
  performancePeriod: PerformancePeriod
  /** Set once the user has completed onboarding at least once. */
  hasCompletedOnboarding: boolean

  selectWallet: (walletId: string | null) => void
  setPerformancePeriod: (period: PerformancePeriod) => void
  markOnboardingComplete: () => void
}

export const usePreferencesStore = create<PreferencesState>()(
  persist(
    (set) => ({
      selectedWalletId: null,
      performancePeriod: PerformancePeriod.H24,
      hasCompletedOnboarding: false,

      selectWallet: (selectedWalletId) => set({ selectedWalletId }),
      setPerformancePeriod: (performancePeriod) => set({ performancePeriod }),
      markOnboardingComplete: () => set({ hasCompletedOnboarding: true }),
    }),
    { name: 'maxai.preferences.v1' },
  ),
)
