import { describe, expect, it } from 'vitest'
import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { AssetVisibility, ValuationStatus } from '@/api/types'
import { renderWithProviders } from '@/test/utils'
import {
  makePortfolio,
  makePosition,
  makeUnpricedPosition,
} from '@/test/fixtures'
import { useUiStore } from '@/stores/uiStore'
import { AssetsTable } from './AssetsTable'

describe('AssetsTable', () => {
  it('shows a balance with no value when the price is unknown', () => {
    const portfolio = makePortfolio({
      positions: [makePosition(), makeUnpricedPosition()],
      valuation_status: ValuationStatus.PARTIAL,
    })

    renderWithProviders(<AssetsTable portfolio={portfolio} />)

    const row = screen.getByText('UNKNOWN').closest('tr')
    expect(row).not.toBeNull()
    expect(row).toHaveTextContent('500,000')
    // Neither the price nor the value may be rendered as a fabricated zero.
    expect(row).not.toHaveTextContent('$0')
    expect(row?.textContent).toContain('—')
  })

  it('keeps hidden assets collapsed behind their backend classification', async () => {
    const spam = makePosition({
      asset: makePosition().asset,
      visibility: AssetVisibility.HIDDEN_SPAM,
    })
    const portfolio = makePortfolio({
      positions: [
        makePosition(),
        {
          ...spam,
          asset: {
            ...spam.asset,
            id: 'ethereum:0xspam',
            symbol: 'SPAMCOIN',
            name: 'Airdropped token',
          },
        },
      ],
    })

    renderWithProviders(<AssetsTable portfolio={portfolio} />)

    expect(screen.queryByText('SPAMCOIN')).not.toBeInTheDocument()

    const toggle = screen.getByRole('button', { name: /hidden assets/i })
    await userEvent.click(toggle)

    expect(screen.getByText('SPAMCOIN')).toBeInTheDocument()
    expect(screen.getByText('Spam')).toBeInTheDocument()
    useUiStore.setState({ hiddenAssetsExpanded: false })
  })

  it('does not offer a simulation for a position that has no value', () => {
    const portfolio = makePortfolio({
      positions: [makeUnpricedPosition()],
      valuation_status: ValuationStatus.PARTIAL,
    })

    renderWithProviders(
      <AssetsTable portfolio={portfolio} onSimulate={() => {}} />,
    )

    expect(screen.getByRole('button', { name: /simulate/i })).toBeDisabled()
  })
})
