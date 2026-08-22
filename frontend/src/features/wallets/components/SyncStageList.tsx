import { SyncStage, type WalletSyncState } from '@/api/types'
import { Check } from '@/components/ui/Icon'
import { Spinner } from '@/components/ui/Spinner'
import { syncStageLabels } from '@/lib/copy/labels'
import { cn } from '@/lib/utils/cn'

/** Canonical pipeline order, used only to lay out the stages the backend reports. */
const STAGE_ORDER: readonly SyncStage[] = [
  SyncStage.FETCHING_BALANCES,
  SyncStage.FETCHING_TRANSACTIONS,
  SyncStage.NORMALIZING_ASSETS,
  SyncStage.FETCHING_PRICES,
  SyncStage.CALCULATING_PORTFOLIO,
  SyncStage.PREPARING_ANALYSIS,
]

/**
 * Synchronisation stages.
 *
 * A stage is only "done" when it appears in `stages_completed`, and only
 * "running" when it equals `sync.stage`. There is no timer here: if the backend
 * stops reporting, the list stops moving, which is the honest behaviour.
 */
export function SyncStageList({ sync }: { sync: WalletSyncState }) {
  const completed = new Set(sync.stages_completed)

  return (
    <ol className="space-y-1" aria-live="polite">
      {STAGE_ORDER.map((stage) => {
        const isComplete = completed.has(stage)
        const isRunning = sync.stage === stage && !isComplete

        return (
          <li
            key={stage}
            className={cn(
              'flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm transition-colors',
              isRunning && 'bg-surface-raised',
            )}
          >
            <span className="grid size-5 shrink-0 place-items-center">
              {isComplete ? (
                <Check className="size-4 text-positive" />
              ) : isRunning ? (
                <Spinner className="size-3.5 text-accent" />
              ) : (
                <span className="size-1.5 rounded-full bg-line-strong" />
              )}
            </span>

            <span
              className={cn(
                isComplete
                  ? 'text-fg-muted'
                  : isRunning
                    ? 'text-fg'
                    : 'text-fg-subtle',
              )}
            >
              {syncStageLabels[stage]}
            </span>

            {isRunning ? (
              <span className="sr-only">in progress</span>
            ) : isComplete ? (
              <span className="sr-only">completed</span>
            ) : null}
          </li>
        )
      })}
    </ol>
  )
}
