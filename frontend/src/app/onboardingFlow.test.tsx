import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import App from '@/App'
import { useOnboardingStore } from '@/stores/onboardingStore'
import { usePreferencesStore } from '@/stores/preferencesStore'
import { useUiStore } from '@/stores/uiStore'

const ADDRESS = '0x71C7656EC7ab88b098defB751B7401B5f6d8976F'

/**
 * The MVP journey end to end against the mock backend:
 *
 *   landing -> select network -> enter address -> analyse -> synchronise ->
 *   portfolio -> ask AI -> explain a transaction -> simulate a scenario
 *
 * It runs as one continuous session because the free plan allows a single
 * wallet, which is exactly the constraint a real user hits. Nothing is stubbed
 * beyond time, so the router, API layer, query cache, sync state machine and
 * streaming all take part.
 */
describe('MVP journey', () => {
  beforeEach(() => {
    vi.useFakeTimers({ shouldAdvanceTime: true })
    window.history.pushState({}, '', '/')
    useOnboardingStore.getState().reset()
    usePreferencesStore.setState({
      selectedWalletId: null,
      hasCompletedOnboarding: false,
    })
    useUiStore.setState({ aiPanelOpen: false, selectedTransactionId: null })
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('goes from landing page to portfolio, answers, explanation and scenario', async () => {
    const user = userEvent.setup({
      advanceTimers: (ms) => vi.advanceTimersByTime(ms),
    })

    render(<App />)

    expect(
      screen.getByRole('heading', { name: /understand your crypto/i }),
    ).toBeInTheDocument()

    await user.click(
      screen.getAllByRole('link', {
        name: /analyze your wallet/i,
      })[0] as HTMLElement,
    )

    // The network is chosen explicitly; nothing is detected from the address.
    expect(
      await screen.findByRole('heading', {
        name: /which network is your wallet on/i,
      }),
    ).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: /Ethereum/ }))

    const input = await screen.findByLabelText(/wallet address/i)
    await user.type(input, ADDRESS)
    await user.click(screen.getByRole('button', { name: /analyze wallet/i }))

    // The wallet exists long before the portfolio does.
    await waitFor(
      () => {
        expect(
          screen.getByText(/analysis queued|analysing wallet/i),
        ).toBeInTheDocument()
      },
      { timeout: 5_000 },
    )

    // Let the backend job work through its stages.
    await vi.advanceTimersByTimeAsync(14_000)

    await waitFor(
      () => {
        expect(screen.getByText('Portfolio')).toBeInTheDocument()
      },
      { timeout: 5_000 },
    )
    expect(await screen.findByText('Assets')).toBeInTheDocument()

    // Ask AI: a suggested question, streamed, answered with claims attached.
    await user.click(
      await screen.findByRole('button', {
        name: /why is my portfolio (down|up)\?|what is my portfolio doing\?/i,
      }),
    )
    await vi.advanceTimersByTimeAsync(20_000)

    await waitFor(
      () => {
        expect(screen.getByText(/claims? backed by backend data/i)).toBeVisible()
      },
      { timeout: 5_000 },
    )
    // The meter reflects the backend's count, not a local counter.
    expect(await screen.findByText('1 / 10')).toBeInTheDocument()

    // Transaction explainer.
    const rows = await screen.findAllByRole('button', {
      name: /open .* details/i,
    })
    await user.click(rows[0] as HTMLElement)

    const transactionDialog = await screen.findByRole('dialog')
    await vi.advanceTimersByTimeAsync(2_000)
    await user.click(
      await within(transactionDialog).findByRole('button', {
        name: /explain with ai/i,
      }),
    )
    await vi.advanceTimersByTimeAsync(20_000)

    await waitFor(() => {
      expect(transactionDialog).toHaveTextContent(/explanation/i)
    })

    await user.click(
      within(transactionDialog).getByRole('button', { name: /close dialog/i }),
    )

    // Scenario simulation: the projection comes back from the backend.
    await user.click(screen.getByRole('button', { name: /run a scenario/i }))
    const scenarioDialog = await screen.findByRole('dialog')
    await user.click(
      within(scenarioDialog).getByRole('button', { name: 'Simulate' }),
    )
    await vi.advanceTimersByTimeAsync(20_000)

    await waitFor(
      () => {
        expect(scenarioDialog).toHaveTextContent(/portfolio now/i)
        expect(scenarioDialog).toHaveTextContent(/difference/i)
      },
      { timeout: 5_000 },
    )
  }, 60_000)
})
