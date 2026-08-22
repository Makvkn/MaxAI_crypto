import { SyncStatus, type Wallet } from '@/api/types'
import { Button, ButtonLink } from '@/components/ui/Button'
import { Card, CardBody } from '@/components/ui/Card'
import { Refresh, Warning } from '@/components/ui/Icon'
import { Spinner } from '@/components/ui/Spinner'
import { chainPresentation } from '@/app/config/chains'
import { ChainMonogram } from './ChainMonogram'
import { SyncStageList } from './SyncStageList'
import { errorBodyCopy } from '@/lib/errors/messages'
import { truncateMiddle } from '@/lib/formatting/address'
import { syncStatusLabels } from '@/lib/copy/labels'

/**
 * Initial synchronisation.
 *
 * `POST /wallets` only enqueues a job, so this screen exists for the window
 * between "wallet created" and "portfolio available". It reflects backend state
 * exactly: queued, the stage actually running, or a failure.
 */
export function InitialSyncScreen({
  wallet,
  onCheckAgain,
  isChecking,
}: {
  wallet: Wallet
  onCheckAgain: () => void
  isChecking?: boolean
}) {
  const chain = chainPresentation(wallet.chain_id)
  const failed = wallet.sync.status === SyncStatus.FAILED

  return (
    <div className="mx-auto max-w-lg py-10 sm:py-16">
      <Card>
        <CardBody className="p-7">
          <div className="flex items-center gap-3.5">
            <ChainMonogram chainId={wallet.chain_id} size="lg" />
            <div className="min-w-0">
              <p className="text-sm font-medium text-fg">{chain.name}</p>
              <p className="truncate font-mono text-[12px] text-fg-subtle">
                {truncateMiddle(wallet.address, { lead: 10, tail: 8 })}
              </p>
            </div>
          </div>

          {failed ? (
            <SyncFailed
              wallet={wallet}
              onCheckAgain={onCheckAgain}
              isChecking={isChecking}
            />
          ) : (
            <SyncRunning wallet={wallet} />
          )}
        </CardBody>
      </Card>
    </div>
  )
}

function SyncRunning({ wallet }: { wallet: Wallet }) {
  const isQueued = wallet.sync.status === SyncStatus.PENDING

  return (
    <>
      <div className="mt-7 flex items-center gap-2.5">
        <Spinner className="size-4 text-accent" />
        <h1 className="text-lg font-medium tracking-tight text-fg">
          {isQueued ? 'Analysis queued' : 'Analysing wallet'}
        </h1>
      </div>

      <p className="mt-2 text-[13px] leading-relaxed text-fg-muted">
        {isQueued
          ? 'Your wallet is in the queue. The first analysis starts automatically.'
          : 'Reading the chain, normalising assets and valuing your positions. You can leave this page — the work continues in the background.'}
      </p>

      <div className="mt-6 border-t border-line pt-4">
        <SyncStageList sync={wallet.sync} />
      </div>

      <p className="mt-5 text-[12px] text-fg-subtle">
        Status: {syncStatusLabels[wallet.sync.status]}
      </p>
    </>
  )
}

function SyncFailed({
  wallet,
  onCheckAgain,
  isChecking,
}: {
  wallet: Wallet
  onCheckAgain: () => void
  isChecking?: boolean
}) {
  const copy = errorBodyCopy(wallet.sync.error)

  return (
    <div role="alert">
      <div className="mt-7 flex items-center gap-2.5">
        <Warning className="size-4 text-caution" />
        <h1 className="text-lg font-medium tracking-tight text-fg">
          {copy.title}
        </h1>
      </div>

      <p className="mt-2 text-[13px] leading-relaxed text-fg-muted">
        {copy.description}
      </p>

      <div className="mt-6 flex flex-wrap gap-2.5">
        <Button
          onClick={onCheckAgain}
          loading={isChecking}
          iconLeft={<Refresh className="size-3.5" />}
        >
          Check again
        </Button>
        <ButtonLink to="/analyze" variant="secondary">
          Try another wallet
        </ButtonLink>
      </div>

      <p className="mt-5 text-[12px] leading-relaxed text-fg-subtle">
        Nothing has changed in your wallet. This only affects our ability to
        read {chainPresentation(wallet.chain_id).name} data right now.
      </p>
    </div>
  )
}
