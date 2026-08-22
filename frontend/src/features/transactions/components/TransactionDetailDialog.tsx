import { useEffect, type ReactNode } from 'react'
import { TransactionStatus, type Transaction } from '@/api/types'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Dialog } from '@/components/ui/Dialog'
import { ExternalLink, Sparkle } from '@/components/ui/Icon'
import { Skeleton } from '@/components/ui/Skeleton'
import { ErrorState } from '@/components/feedback/States'
import { Money, UnknownValue } from '@/components/finance/Money'
import { AiStreamView } from '@/features/ai/components/AiStreamView'
import { useAiStream, useAiUsage } from '@/features/ai/hooks/useAiConversation'
import { useTransaction } from '../hooks/useTransactions'
import { formatBalance } from '@/lib/formatting/money'
import { formatDateTime } from '@/lib/dates/format'
import { truncateMiddle } from '@/lib/formatting/address'
import { transactionStatusLabels, transactionTypeLabels } from '@/lib/copy/labels'
import { analytics } from '@/lib/analytics/analytics'

/**
 * Transaction detail and explanation.
 *
 * The facts are the backend's canonical transaction record — amounts and fees
 * are never derived from raw chain data here. "Explain" asks the AI about that
 * record by id, so the explanation is anchored to the same facts on screen.
 */
export function TransactionDetailDialog({
  walletId,
  transactionId,
  onClose,
}: {
  walletId: string
  transactionId: string | null
  onClose: () => void
}) {
  const transactionQuery = useTransaction(walletId, transactionId)
  const usage = useAiUsage(Boolean(transactionId))
  const { state, send, cancel, reset, isSending } = useAiStream({
    walletId,
    conversationId: null,
  })

  useEffect(() => {
    if (!transactionId) reset()
  }, [reset, transactionId])

  const transaction = transactionQuery.data
  const limitReached = usage.data ? usage.data.remaining <= 0 : false

  const explain = () => {
    if (!transaction) return
    analytics.track('transaction_explained', {
      wallet_id: walletId,
      transaction_type: transaction.type,
    })
    void send({
      content: 'Explain this transaction.',
      context: { transaction_id: transaction.id },
    })
  }

  return (
    <Dialog
      open={Boolean(transactionId)}
      onClose={() => {
        cancel()
        onClose()
      }}
      title={
        transaction
          ? transactionTypeLabels[transaction.type]
          : 'Transaction details'
      }
      description={
        transaction ? formatDateTime(transaction.timestamp) : undefined
      }
      size="lg"
      footer={
        transaction ? (
          <>
            {transaction.explorer_url ? (
              <Button
                variant="quiet"
                onClick={() =>
                  window.open(
                    transaction.explorer_url as string,
                    '_blank',
                    'noopener,noreferrer',
                  )
                }
                iconRight={<ExternalLink className="size-3.5" />}
              >
                View on explorer
              </Button>
            ) : null}
            <Button
              onClick={explain}
              loading={isSending}
              disabled={limitReached}
              iconLeft={<Sparkle className="size-3.5" />}
            >
              Explain with AI
            </Button>
          </>
        ) : null
      }
    >
      {transactionQuery.isPending ? (
        <div className="space-y-3">
          <Skeleton className="h-4 w-40" />
          <Skeleton className="h-4 w-64" />
          <Skeleton className="h-4 w-52" />
        </div>
      ) : transactionQuery.isError ? (
        <ErrorState
          error={transactionQuery.error}
          onRetry={() => void transactionQuery.refetch()}
          compact
        />
      ) : transaction ? (
        <div className="space-y-6">
          <TransactionFacts transaction={transaction} />

          {state.status !== 'idle' ? (
            <div className="rounded-card border border-line bg-base-elevated px-4 py-4">
              <p className="mb-3 flex items-center gap-2 text-[11px] tracking-[0.08em] text-fg-subtle uppercase">
                <Sparkle className="size-3.5 text-accent" />
                Explanation
              </p>
              <AiStreamView state={state} />
            </div>
          ) : limitReached ? (
            <p className="text-[12px] text-caution">
              You've reached your daily AI limit, so explanations are paused
              until it resets.
            </p>
          ) : null}
        </div>
      ) : null}
    </Dialog>
  )
}

function TransactionFacts({ transaction }: { transaction: Transaction }) {
  return (
    <dl className="grid gap-x-8 gap-y-4 sm:grid-cols-2">
      <Fact label="Status">
        <Badge
          tone={
            transaction.status === TransactionStatus.SUCCESS
              ? 'positive'
              : transaction.status === TransactionStatus.FAILED
                ? 'negative'
                : 'caution'
          }
        >
          {transactionStatusLabels[transaction.status]}
        </Badge>
      </Fact>

      <Fact label="Type">{transactionTypeLabels[transaction.type]}</Fact>

      {transaction.asset_out ? (
        <Fact label="Sent">
          {formatBalance(transaction.amount_out, {
            symbol: transaction.asset_out.symbol,
          })}
          <span className="ml-2 text-fg-subtle">
            <Money
              value={transaction.value_out_usd}
              unknownReason="Value at the time is unknown"
            />
          </span>
        </Fact>
      ) : null}

      {transaction.asset_in ? (
        <Fact label="Received">
          {formatBalance(transaction.amount_in, {
            symbol: transaction.asset_in.symbol,
          })}
          <span className="ml-2 text-fg-subtle">
            <Money
              value={transaction.value_in_usd}
              unknownReason="Value at the time is unknown"
            />
          </span>
        </Fact>
      ) : null}

      <Fact label="Network fee">
        {transaction.fee_amount === null ? (
          <UnknownValue reason="Fee is not available" />
        ) : (
          <>
            {formatBalance(transaction.fee_amount, {
              symbol: transaction.fee_asset?.symbol,
            })}
            <span className="ml-2 text-fg-subtle">
              <Money value={transaction.fee_value_usd} />
            </span>
          </>
        )}
      </Fact>

      {transaction.protocol ? (
        <Fact label="Protocol">{transaction.protocol}</Fact>
      ) : null}

      <Fact label="From">
        {transaction.from_address ? (
          <span className="font-mono text-[12px]">
            {truncateMiddle(transaction.from_address, { lead: 10, tail: 8 })}
          </span>
        ) : (
          <UnknownValue reason="Sender is not available" />
        )}
      </Fact>

      <Fact label="To">
        {transaction.to_address ? (
          <span className="font-mono text-[12px]">
            {truncateMiddle(transaction.to_address, { lead: 10, tail: 8 })}
          </span>
        ) : (
          <UnknownValue reason="Recipient is not available" />
        )}
      </Fact>

      <Fact label="Transaction hash" wide>
        <span className="font-mono text-[12px] break-all">
          {transaction.tx_hash}
        </span>
      </Fact>
    </dl>
  )
}

function Fact({
  label,
  children,
  wide,
}: {
  label: string
  children: ReactNode
  wide?: boolean
}) {
  return (
    <div className={wide ? 'sm:col-span-2' : undefined}>
      <dt className="text-[11px] tracking-[0.06em] text-fg-subtle uppercase">
        {label}
      </dt>
      <dd className="mt-1 text-[13px] text-fg">{children}</dd>
    </div>
  )
}
