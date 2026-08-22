import { useMemo } from 'react'
import {
  Area,
  AreaChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip as ChartTooltip,
  XAxis,
  YAxis,
} from 'recharts'
import {
  PerformancePeriod,
  ValuationStatus,
  type PortfolioSnapshotPoint,
} from '@/api/types'
import { formatUsd } from '@/lib/formatting/money'
import { formatDate, formatDateTime, formatTime } from '@/lib/dates/format'

/**
 * Historical portfolio value.
 *
 * Every point is a stored backend snapshot. Snapshots whose valuation was
 * unavailable are rendered as gaps — the line is never bridged across missing
 * data, because that would invent history.
 */
export function PortfolioChart({
  series,
  period,
  positive,
}: {
  series: PortfolioSnapshotPoint[]
  period: PerformancePeriod
  positive: boolean
}) {
  const data = useMemo(
    () =>
      series.map((point) => ({
        t: new Date(point.captured_at).getTime(),
        // Recharts needs a number; the underlying decimal is never mutated.
        value:
          point.total_value_usd === null ? null : Number(point.total_value_usd),
        raw: point.total_value_usd,
        partial: point.status === ValuationStatus.PARTIAL,
      })),
    [series],
  )

  const stroke = positive ? 'var(--color-positive)' : 'var(--color-negative)'
  const gradientId = positive ? 'perf-up' : 'perf-down'

  return (
    <div className="h-56 w-full sm:h-64">
      <ResponsiveContainer width="100%" height="100%">
        <AreaChart data={data} margin={{ top: 8, right: 8, bottom: 0, left: 0 }}>
          <defs>
            <linearGradient id={gradientId} x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor={stroke} stopOpacity={0.22} />
              <stop offset="100%" stopColor={stroke} stopOpacity={0} />
            </linearGradient>
          </defs>

          <CartesianGrid
            stroke="var(--color-line)"
            strokeDasharray="2 6"
            vertical={false}
          />

          <XAxis
            dataKey="t"
            type="number"
            scale="time"
            domain={['dataMin', 'dataMax']}
            tickFormatter={(value: number) => axisTick(value, period)}
            tick={{ fill: 'var(--color-fg-subtle)', fontSize: 11 }}
            axisLine={false}
            tickLine={false}
            minTickGap={40}
          />

          <YAxis
            dataKey="value"
            domain={['auto', 'auto']}
            tickFormatter={(value: number) =>
              formatUsd(String(value), { compact: true })
            }
            tick={{ fill: 'var(--color-fg-subtle)', fontSize: 11 }}
            axisLine={false}
            tickLine={false}
            width={58}
          />

          <ChartTooltip
            content={<SnapshotTooltip />}
            cursor={{ stroke: 'var(--color-line-strong)' }}
          />

          <Area
            type="monotone"
            dataKey="value"
            stroke={stroke}
            strokeWidth={1.75}
            fill={`url(#${gradientId})`}
            // Gaps stay gaps: missing snapshots are not interpolated.
            connectNulls={false}
            dot={false}
            activeDot={{ r: 3, fill: stroke, stroke: 'var(--color-base)' }}
            isAnimationActive={false}
          />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  )
}

interface TooltipPayloadItem {
  payload?: { t: number; raw: string | null; partial: boolean }
}

function SnapshotTooltip({
  active,
  payload,
}: {
  active?: boolean
  payload?: TooltipPayloadItem[]
}) {
  const point = active ? payload?.[0]?.payload : undefined
  if (!point) return null

  return (
    <div className="rounded-lg border border-line-strong bg-surface-raised px-3 py-2 shadow-lg">
      <p className="text-[11px] text-fg-subtle">
        {formatDateTime(new Date(point.t).toISOString())}
      </p>
      <p className="mt-0.5 text-sm font-medium text-fg tabular">
        {formatUsd(point.raw)}
      </p>
      {point.partial ? (
        <p className="mt-1 text-[11px] text-caution">Partially valued</p>
      ) : null}
    </div>
  )
}

function axisTick(value: number, period: PerformancePeriod): string {
  const iso = new Date(value).toISOString()
  return period === PerformancePeriod.H24 ? formatTime(iso) : formatDate(iso)
}
