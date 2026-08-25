import { describe, expect, it } from 'vitest'
import { screen } from '@testing-library/react'
import { DataFreshness, DataQuality, ValuationStatus } from '@/api/types'
import { renderWithProviders } from '@/test/utils'
import { makePortfolio, makeWallet } from '@/test/fixtures'
import { PortfolioHeader } from './PortfolioHeader'

const noop = () => {}

describe('PortfolioHeader', () => {
  it('renders the backend total and its 24h change', () => {
    renderWithProviders(
      <PortfolioHeader
        wallet={makeWallet()}
        portfolio={makePortfolio()}
        onRefresh={noop}
        isRefreshing={false}
      />,
    )

    expect(screen.getByText('$24,850.12')).toBeInTheDocument()
    expect(screen.getByText('-4.21%')).toBeInTheDocument()
    expect(screen.getByText('(-$1,092.44)')).toBeInTheDocument()
  })

  it('flags a partial valuation next to the figure', () => {
    renderWithProviders(
      <PortfolioHeader
        wallet={makeWallet()}
        portfolio={makePortfolio({
          valuation_status: ValuationStatus.PARTIAL,
          data_quality: DataQuality.PARTIAL,
          exclusions: {
            unpriced_positions: 1,
            nfts_excluded: true,
            defi_positions_excluded: true,
          },
        })}
        onRefresh={noop}
        isRefreshing={false}
      />,
    )

    expect(screen.getByText('Partial')).toBeInTheDocument()
    expect(screen.getByText('$24,850.12')).toBeInTheDocument()
  })

  it('shows no value at all when the valuation is unavailable', () => {
    renderWithProviders(
      <PortfolioHeader
        wallet={makeWallet()}
        portfolio={makePortfolio({
          total_value_usd: null,
          valuation_status: ValuationStatus.UNAVAILABLE,
          data_quality: DataQuality.UNAVAILABLE,
          change_24h_pct: null,
          change_24h_usd: null,
        })}
        onRefresh={noop}
        isRefreshing={false}
      />,
    )

    expect(
      screen.getByLabelText('Portfolio valuation is unavailable'),
    ).toHaveTextContent('—')
    expect(screen.queryByText('$0.00')).not.toBeInTheDocument()
  })

  it('shows $0 for a known-empty wallet instead of a dash', () => {
    renderWithProviders(
      <PortfolioHeader
        wallet={makeWallet()}
        portfolio={makePortfolio({
          total_value_usd: '0',
          valuation_status: ValuationStatus.COMPLETE,
          data_quality: DataQuality.COMPLETE,
          change_24h_pct: '0',
          change_24h_usd: '0',
          positions: [],
        })}
        onRefresh={noop}
        isRefreshing={false}
      />,
    )

    expect(screen.getByText('$0.00')).toBeInTheDocument()
    expect(
      screen.getByText('This wallet has no holdings on the selected chain.'),
    ).toBeInTheDocument()
    expect(
      screen.queryByLabelText('Portfolio valuation is unavailable'),
    ).not.toBeInTheDocument()
  })

  it('labels stale data instead of hiding it', () => {
    renderWithProviders(
      <PortfolioHeader
        wallet={makeWallet()}
        portfolio={makePortfolio({ data_freshness: DataFreshness.STALE })}
        onRefresh={noop}
        isRefreshing={false}
      />,
    )

    expect(screen.getByText('Stale')).toBeInTheDocument()
  })
})
