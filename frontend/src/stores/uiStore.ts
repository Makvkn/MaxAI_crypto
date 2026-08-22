import { create } from 'zustand'

/**
 * Ephemeral UI state.
 *
 * Zustand holds interaction state only: what is open, what is selected, what is
 * expanded. No server entity is ever stored here — wallets, portfolios,
 * transactions and conversations live in TanStack Query.
 */
export type DialogId = 'auth' | 'ai-limit' | 'wallet-switcher' | null

interface UiState {
  navOpen: boolean
  /** AI panel visibility on tablet/mobile, where it collapses. */
  aiPanelOpen: boolean
  activeDialog: DialogId
  /** Transaction whose detail panel is open. */
  selectedTransactionId: string | null
  hiddenAssetsExpanded: boolean
  scenarioOpen: boolean

  setNavOpen: (open: boolean) => void
  toggleAiPanel: () => void
  setAiPanelOpen: (open: boolean) => void
  openDialog: (dialog: NonNullable<DialogId>) => void
  closeDialog: () => void
  selectTransaction: (transactionId: string | null) => void
  toggleHiddenAssets: () => void
  setScenarioOpen: (open: boolean) => void
}

export const useUiStore = create<UiState>((set) => ({
  navOpen: false,
  aiPanelOpen: false,
  activeDialog: null,
  selectedTransactionId: null,
  hiddenAssetsExpanded: false,
  scenarioOpen: false,

  setNavOpen: (navOpen) => set({ navOpen }),
  toggleAiPanel: () => set((state) => ({ aiPanelOpen: !state.aiPanelOpen })),
  setAiPanelOpen: (aiPanelOpen) => set({ aiPanelOpen }),
  openDialog: (activeDialog) => set({ activeDialog }),
  closeDialog: () => set({ activeDialog: null }),
  selectTransaction: (selectedTransactionId) => set({ selectedTransactionId }),
  toggleHiddenAssets: () =>
    set((state) => ({ hiddenAssetsExpanded: !state.hiddenAssetsExpanded })),
  setScenarioOpen: (scenarioOpen) => set({ scenarioOpen }),
}))
