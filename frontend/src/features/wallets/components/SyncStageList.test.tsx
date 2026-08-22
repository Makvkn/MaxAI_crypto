import { describe, expect, it } from 'vitest'
import { screen } from '@testing-library/react'
import { SyncStage, SyncStatus, type WalletSyncState } from '@/api/types'
import { renderWithProviders } from '@/test/utils'
import { makeWallet } from '@/test/fixtures'
import { SyncStageList } from './SyncStageList'
import { InitialSyncScreen } from './InitialSyncScreen'

function syncState(overrides: Partial<WalletSyncState>): WalletSyncState {
  return {
    status: SyncStatus.SYNCING,
    stage: SyncStage.FETCHING_PRICES,
    stages_completed: [
      SyncStage.FETCHING_BALANCES,
      SyncStage.FETCHING_TRANSACTIONS,
      SyncStage.NORMALIZING_ASSETS,
    ],
    started_at: '2026-08-21T12:00:00.000Z',
    completed_at: null,
    last_synced_at: null,
    data_freshness: null,
    error: null,
    ...overrides,
  }
}

describe('synchronisation UX', () => {
  it('marks only the stages the backend reported', () => {
    renderWithProviders(<SyncStageList sync={syncState({})} />)

    const items = screen.getAllByRole('listitem')
    expect(items).toHaveLength(6)

    expect(screen.getByText('Fetching balances').closest('li')).toHaveTextContent(
      'completed',
    )
    expect(
      screen.getByText('Fetching market prices').closest('li'),
    ).toHaveTextContent('in progress')
    // A stage the backend has not reached is neither running nor complete.
    const notStarted = screen.getByText('Preparing analysis').closest('li')
    expect(notStarted).not.toHaveTextContent('completed')
    expect(notStarted).not.toHaveTextContent('in progress')
  })

  it('describes a queued job without inventing progress', () => {
    const wallet = makeWallet({
      sync: syncState({
        status: SyncStatus.PENDING,
        stage: null,
        stages_completed: [],
      }),
    })

    renderWithProviders(
      <InitialSyncScreen wallet={wallet} onCheckAgain={() => {}} />,
    )

    expect(screen.getByText('Analysis queued')).toBeInTheDocument()
    expect(screen.queryByText(/completed/)).not.toBeInTheDocument()
  })

  it('explains a failed sync in domain language and offers a retry', () => {
    const wallet = makeWallet({
      sync: syncState({
        status: SyncStatus.FAILED,
        stage: null,
        error: {
          code: 'WALLET_SYNC_FAILED',
          message: 'Tatum 429 Too Many Requests',
        },
      }),
    })

    renderWithProviders(
      <InitialSyncScreen wallet={wallet} onCheckAgain={() => {}} />,
    )

    expect(screen.getByText('Unable to analyse this wallet')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /check again/i })).toBeVisible()
    // Provider details must never reach the user.
    expect(screen.queryByText(/tatum/i)).not.toBeInTheDocument()
  })
})
