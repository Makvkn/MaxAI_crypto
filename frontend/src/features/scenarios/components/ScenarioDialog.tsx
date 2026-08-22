import { useState, type ReactNode } from 'react'
import {
  AssetVisibility,
  type Portfolio,
  type ScenarioResult,
  type WalletPosition,
} from '@/api/types'
import { Button } from '@/components/ui/Button'
import { Dialog } from '@/components/ui/Dialog'
import { Input } from '@/components/ui/Input'
import { Scenario, Sparkle } from '@/components/ui/Icon'
import {
  SegmentedControl,
  type SegmentOption,
} from '@/components/ui/SegmentedControl'
import { Spinner } from '@/components/ui/Spinner'
import { EmptyState, ErrorState } from '@/components/feedback/States'
import { Delta, Money } from '@/components/finance/Money'
import { AiAnswer } from '@/features/ai/components/AiAnswer'
import { useAiUsage } from '@/features/ai/hooks/useAiConversation'
import { useScenarioSimulation } from '../hooks/useScenarioSimulation'
import { formatPercentDelta } from '@/lib/formatting/percent'

const PRESETS: readonly SegmentOption<string>[] = [
  { value: '-20', label: '-20%' },
  { value: '-10', label: '-10%' },
  { value: '10', label: '+10%' },
  { value: '20', label: '+20%' },
]

/**
 * Scenario simulator.
 *
 * The user picks an asset and a percentage; the backend computes the projection
 * and the AI explains it. No impact is calculated in the browser — this dialog
 * only submits the request and renders the structured result.
 */
export function ScenarioDialog({
  walletId,
  portfolio,
  initialPosition,
  open,
  onClose,
}: {
  walletId: string
  portfolio: Portfolio | undefined
  initialPosition: WalletPosition | null
  open: boolean
  onClose: () => void
}) {
  const priced = (portfolio?.positions ?? []).filter(
    (position) =>
      position.visibility === AssetVisibility.VISIBLE &&
      position.value_usd !== null,
  )

  // Mounted only while open, so opening from an asset row starts on that asset.
  const [assetId, setAssetId] = useState<string | null>(
    initialPosition?.asset.id ?? priced[0]?.asset.id ?? null,
  )
  const [changePct, setChangePct] = useState('-20')
  const simulation = useScenarioSimulation(walletId)
  const usage = useAiUsage(open)

  const selected = priced.find((position) => position.asset.id === assetId)
  const limitReached = usage.data ? usage.data.remaining <= 0 : false
  const numericChange = Number(changePct)
  const validChange =
    changePct.trim() !== '' &&
    Number.isFinite(numericChange) &&
    numericChange !== 0 &&
    Math.abs(numericChange) <= 100

  const run = () => {
    if (!selected || !validChange) return
    simulation.mutate({
      assetId: selected.asset.id,
      assetSymbol: selected.asset.symbol,
      changePct: changePct.trim(),
    })
  }

  return (
    <Dialog
      open={open}
      onClose={onClose}
      title="Scenario simulation"
      description="Test a price move against your actual positions."
      size="lg"
      footer={
        <>
          <Button variant="quiet" onClick={onClose}>
            Close
          </Button>
          <Button
            onClick={run}
            loading={simulation.isPending}
            disabled={!selected || !validChange || limitReached}
            iconLeft={<Scenario className="size-3.5" />}
          >
            Simulate
          </Button>
        </>
      }
    >
      {priced.length === 0 ? (
        <EmptyState
          title="No priced assets to simulate"
          description="A scenario needs at least one held asset with a reliable market price."
        />
      ) : (
        <div className="space-y-6">
          <div className="grid gap-4 sm:grid-cols-2">
            <div>
              <label
                htmlFor="scenario-asset"
                className="mb-2 block text-[13px] font-medium text-fg-muted"
              >
                Asset
              </label>
              <select
                id="scenario-asset"
                value={assetId ?? ''}
                onChange={(event) => setAssetId(event.target.value)}
                className="w-full rounded-lg border border-line-strong bg-base-elevated px-3.5 py-3 text-sm text-fg focus:outline-none focus-visible:border-accent focus-visible:ring-2 focus-visible:ring-accent/25"
              >
                {priced.map((position) => (
                  <option key={position.asset.id} value={position.asset.id}>
                    {position.asset.symbol} — {position.asset.name}
                  </option>
                ))}
              </select>
            </div>

            <div>
              <span className="mb-2 block text-[13px] font-medium text-fg-muted">
                Price change
              </span>
              <SegmentedControl
                label="Price change preset"
                options={PRESETS}
                value={changePct}
                onChange={setChangePct}
              />
              <Input
                label="Custom change"
                hideLabel
                className="mt-2"
                value={changePct}
                onChange={(event) => setChangePct(event.target.value)}
                inputMode="decimal"
                adornment="%"
                error={
                  changePct.trim() !== '' && !validChange
                    ? 'Enter a percentage between -100 and 100.'
                    : null
                }
              />
            </div>
          </div>

          {selected ? (
            <p className="text-[13px] text-fg-muted">
              What if{' '}
              <span className="text-fg">{selected.asset.symbol}</span>{' '}
              {numericChange < 0 ? 'falls' : 'rises'}{' '}
              {Math.abs(numericChange) || 0}%?
            </p>
          ) : null}

          {limitReached ? (
            <p className="text-[12px] text-caution">
              You've reached your daily AI limit, so simulations are paused
              until it resets.
            </p>
          ) : null}

          {simulation.isPending ? (
            <p className="flex items-center gap-2 text-[13px] text-fg-subtle">
              <Spinner className="size-3.5 text-accent" />
              Running the calculation and preparing the explanation
            </p>
          ) : simulation.isError ? (
            <ErrorState error={simulation.error} onRetry={run} compact />
          ) : simulation.data ? (
            <ScenarioResultView result={simulation.data} />
          ) : null}
        </div>
      )}
    </Dialog>
  )
}

function ScenarioResultView({ result }: { result: ScenarioResult }) {
  return (
    <div className="space-y-5 rounded-card border border-line bg-base-elevated px-4 py-4">
      <div className="grid gap-4 sm:grid-cols-3">
        <Figure label="Portfolio now">
          <Money value={result.baseline.portfolio_value_usd} />
        </Figure>
        <Figure
          label={`If ${result.asset.symbol} moves ${formatPercentDelta(result.change_pct)}`}
        >
          <Money value={result.projection.portfolio_value_usd} />
        </Figure>
        <Figure label="Difference">
          <Delta
            percent={result.projection.portfolio_change_pct}
            amount={result.projection.portfolio_change_usd}
            size="sm"
          />
        </Figure>
      </div>

      <div className="grid gap-4 border-t border-line pt-4 sm:grid-cols-3">
        <Figure label={`${result.asset.symbol} now`}>
          <Money value={result.baseline.asset_value_usd} />
        </Figure>
        <Figure label={`${result.asset.symbol} after`}>
          <Money value={result.projection.asset_value_usd} />
        </Figure>
        <Figure label="Impact on this asset">
          <Money value={result.projection.asset_impact_usd} />
        </Figure>
      </div>

      {result.explanation ? (
        <div className="border-t border-line pt-4">
          <p className="mb-3 flex items-center gap-2 text-[11px] tracking-[0.08em] text-fg-subtle uppercase">
            <Sparkle className="size-3.5 text-accent" />
            What this means
          </p>
          <AiAnswer response={result.explanation} />
        </div>
      ) : null}
    </div>
  )
}

function Figure({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div>
      <p className="text-[11px] tracking-[0.06em] text-fg-subtle uppercase">
        {label}
      </p>
      <p className="mt-1 text-[15px] font-medium text-fg">{children}</p>
    </div>
  )
}
