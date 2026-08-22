import type { PerformanceDriver, PerformancePeriod } from '@/api/types'
import { Delta, Money } from '@/components/finance/Money'
import { AssetMonogram } from '@/features/assets/components/AssetMonogram'
import { periodDescriptions } from '@/lib/copy/labels'

/**
 * What moved the portfolio.
 *
 * `contribution_usd` is the backend's attribution of the period result to each
 * asset. This list orders and formats it; it never recomputes a contribution.
 */
export function PerformanceDrivers({
  drivers,
  period,
}: {
  drivers: PerformanceDriver[]
  period: PerformancePeriod
}) {
  return (
    <div className="border-t border-line px-5 py-4">
      <h3 className="text-[11px] font-medium tracking-[0.08em] text-fg-subtle uppercase">
        What moved it
      </h3>
      <p className="mt-1 text-[12px] text-fg-subtle">
        Contribution to the change over {periodDescriptions[period]}.
      </p>

      <ul className="mt-4 space-y-3">
        {drivers.map((driver) => (
          <li
            key={driver.asset.id}
            className="flex items-center gap-3 text-[13px]"
          >
            <AssetMonogram asset={driver.asset} size="sm" />
            <span className="min-w-0 flex-1 truncate text-fg-muted">
              {driver.asset.symbol}
            </span>
            <span className="w-24 text-right">
              <Money
                value={driver.contribution_usd}
                unknownReason="Contribution is not available"
                className="text-fg"
              />
            </span>
            <span className="w-24 text-right">
              <Delta percent={driver.change_pct} size="sm" showIcon={false} />
            </span>
          </li>
        ))}
      </ul>
    </div>
  )
}
