import {
  PerformancePeriod,
  PerformanceStatus,
  type PortfolioPerformance,
} from '@/api/types'
import { Card, CardBody, CardHeader } from '@/components/ui/Card'
import { Info } from '@/components/ui/Icon'
import { SegmentedControl } from '@/components/ui/SegmentedControl'
import { Skeleton } from '@/components/ui/Skeleton'
import { EmptyState, ErrorState } from '@/components/feedback/States'
import { Delta, Money } from '@/components/finance/Money'
import { PortfolioChart } from './PortfolioChart'
import { PerformanceDrivers } from './PerformanceDrivers'
import { usePerformance } from '../hooks/usePerformance'
import { periodDescriptions, periodLabels } from '@/lib/copy/labels'
import { errorBodyCopy } from '@/lib/errors/messages'
import { deltaDirection } from '@/lib/formatting/percent'
import { formatDateTime } from '@/lib/dates/format'

const PERIOD_OPTIONS = [
  { value: PerformancePeriod.H24, label: periodLabels[PerformancePeriod.H24] },
  { value: PerformancePeriod.D7, label: periodLabels[PerformancePeriod.D7] },
  { value: PerformancePeriod.D30, label: periodLabels[PerformancePeriod.D30] },
  { value: PerformancePeriod.ALL, label: periodLabels[PerformancePeriod.ALL] },
] as const

/**
 * Portfolio performance for the selected period.
 *
 * Deliberately called performance, not PnL: the MVP has no cost basis. Values,
 * the snapshot series and the per-asset drivers all come from the backend.
 */
export function PerformanceCard({
  walletId,
  period,
  onPeriodChange,
  enabled = true,
}: {
  walletId: string
  period: PerformancePeriod
  onPeriodChange: (period: PerformancePeriod) => void
  enabled?: boolean
}) {
  const performanceQuery = usePerformance(walletId, period, { enabled })

  return (
    <Card>
      <CardHeader
        title="Portfolio performance"
        action={
          <SegmentedControl
            label="Performance period"
            options={PERIOD_OPTIONS}
            value={period}
            onChange={onPeriodChange}
            size="sm"
          />
        }
      />

      {performanceQuery.isPending ? (
        <CardBody className="space-y-4 py-6">
          <Skeleton className="h-8 w-48" />
          <Skeleton className="h-56 w-full" />
        </CardBody>
      ) : performanceQuery.isError ? (
        <ErrorState
          error={performanceQuery.error}
          onRetry={() => void performanceQuery.refetch()}
          retrying={performanceQuery.isFetching}
        />
      ) : performanceQuery.data ? (
        <PerformanceBody performance={performanceQuery.data} period={period} />
      ) : null}
    </Card>
  )
}

function PerformanceBody({
  performance,
  period,
}: {
  performance: PortfolioPerformance
  period: PerformancePeriod
}) {
  if (performance.status === PerformanceStatus.UNAVAILABLE) {
    return (
      <EmptyState
        title="Performance is not available for this period"
        description={
          performance.unavailable_reason
            ? errorBodyCopy({
                code: performance.unavailable_reason,
                message: '',
              }).description
            : `Performance is measured from stored portfolio snapshots, and there is not enough history for ${periodDescriptions[period]}.`
        }
      />
    )
  }

  const direction = deltaDirection(performance.change_pct)

  return (
    <>
      <CardBody className="pb-2">
        <div className="flex flex-wrap items-end justify-between gap-4">
          <div>
            <Delta
              percent={performance.change_pct}
              amount={performance.change_usd}
              size="lg"
            />
            <p className="mt-1.5 text-[12px] text-fg-subtle">
              over {periodDescriptions[period]}
            </p>
          </div>

          {performance.opening && performance.closing ? (
            <dl className="flex gap-6 text-[12px]">
              <div>
                <dt className="text-fg-subtle">Opening</dt>
                <dd className="mt-0.5 text-fg-muted">
                  <Money value={performance.opening.value_usd} />
                </dd>
              </div>
              <div>
                <dt className="text-fg-subtle">Closing</dt>
                <dd className="mt-0.5 text-fg-muted">
                  <Money value={performance.closing.value_usd} />
                </dd>
              </div>
            </dl>
          ) : null}
        </div>
      </CardBody>

      <div className="px-2 pt-2">
        {performance.series.length === 0 ? (
          <EmptyState
            title="No history yet"
            description="Portfolio snapshots are captured over time. The chart appears once there is more than one."
          />
        ) : (
          <PortfolioChart
            series={performance.series}
            period={period}
            positive={direction !== 'down'}
          />
        )}
      </div>

      <div className="space-y-3 px-5 pt-2 pb-4">
        {performance.closing ? (
          <p className="text-[12px] text-fg-subtle">
            Last snapshot {formatDateTime(performance.closing.captured_at)}.
          </p>
        ) : null}
        {performance.status === PerformanceStatus.PARTIAL ? (
          <p className="flex items-center gap-1.5 text-[12px] text-caution">
            <Info className="size-3.5 shrink-0" />
            Some snapshots in this period are incomplete, so the result covers
            only the history that exists.
          </p>
        ) : null}
      </div>

      {performance.drivers.length > 0 ? (
        <PerformanceDrivers drivers={performance.drivers} period={period} />
      ) : null}
    </>
  )
}
