import { ValuationStatus, type Portfolio, type Wallet } from '@/api/types'
import { Button } from '@/components/ui/Button'
import { Refresh } from '@/components/ui/Icon'
import { Skeleton } from '@/components/ui/Skeleton'
import { Delta, Money, UnknownValue } from '@/components/finance/Money'
import { DataFreshnessBadge } from '@/components/data-quality/DataFreshnessBadge'
import { ValuationStatusNote } from '@/components/data-quality/ValuationStatusNote'
import { chainPresentation } from '@/app/config/chains'
import { truncateMiddle } from '@/lib/formatting/address'

/**
 * Portfolio headline.
 *
 * The total, its 24h change and every qualifier around it come from the
 * backend. When the valuation is unavailable the figure is absent rather than
 * zero, and the reason sits next to it.
 */
export function PortfolioHeader({
  wallet,
  portfolio,
  onRefresh,
  isRefreshing,
}: {
  wallet: Wallet
  portfolio: Portfolio
  onRefresh: () => void
  isRefreshing: boolean
}) {
  const chain = chainPresentation(wallet.chain_id)
  const unavailable =
    portfolio.valuation_status === ValuationStatus.UNAVAILABLE ||
    portfolio.total_value_usd === null

  return (
    <div className="flex flex-wrap items-start justify-between gap-6">
      <div>
        <div className="flex flex-wrap items-center gap-2.5">
          <h1 className="text-[11px] font-medium tracking-[0.1em] text-fg-subtle uppercase">
            Portfolio
          </h1>
          <span className="text-[12px] text-fg-subtle">
            {chain.name} · {truncateMiddle(wallet.address, { tail: 6 })}
          </span>
        </div>

        <div className="mt-3 flex flex-wrap items-baseline gap-x-4 gap-y-2">
          {unavailable ? (
            <UnknownValue
              reason="Portfolio valuation is unavailable"
              className="text-4xl font-medium tracking-tight sm:text-5xl"
            />
          ) : (
            <Money
              value={portfolio.total_value_usd}
              className="text-4xl font-medium tracking-tight text-fg sm:text-5xl"
            />
          )}

          <ValuationStatusNote
            valuationStatus={portfolio.valuation_status}
            dataQuality={portfolio.data_quality}
            unpricedCount={portfolio.exclusions.unpriced_positions}
          />
        </div>

        <div className="mt-3 flex flex-wrap items-center gap-3">
          <Delta
            percent={portfolio.change_24h_pct}
            amount={portfolio.change_24h_usd}
            size="lg"
          />
          <span className="text-[13px] text-fg-subtle">past 24h</span>
        </div>
      </div>

      <div className="flex items-center gap-2.5">
        <DataFreshnessBadge
          freshness={portfolio.data_freshness}
          asOf={portfolio.last_synced_at ?? portfolio.as_of}
        />
        <Button
          variant="secondary"
          size="sm"
          onClick={onRefresh}
          loading={isRefreshing}
          iconLeft={<Refresh className="size-3.5" />}
        >
          Refresh
        </Button>
      </div>
    </div>
  )
}

export function PortfolioHeaderSkeleton() {
  return (
    <div className="space-y-3">
      <Skeleton className="h-3 w-24" />
      <Skeleton className="h-12 w-64" />
      <Skeleton className="h-4 w-40" />
    </div>
  )
}
