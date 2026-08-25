import { useState } from 'react'
import {
  TransactionStatus,
  TransactionType,
  type Transaction,
} from '@/api/types'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Card, CardHeader } from '@/components/ui/Card'
import { ChevronRight } from '@/components/ui/Icon'
import { SegmentedControl } from '@/components/ui/SegmentedControl'
import { SkeletonRows } from '@/components/ui/Skeleton'
import { Table, TableScroll, Td, Th, Tr } from '@/components/ui/Table'
import { EmptyState, ErrorState } from '@/components/feedback/States'
import { Money } from '@/components/finance/Money'
import { AssetMonogram } from '@/features/assets/components/AssetMonogram'
import { useTransactions } from '../hooks/useTransactions'
import { formatBalance } from '@/lib/formatting/money'
import { formatDateTime } from '@/lib/dates/format'
import { transactionStatusLabels, transactionTypeLabels } from '@/lib/copy/labels'

/** Filters map straight onto the backend `type` parameter. */
const FILTERS = [
  { value: 'ALL', label: 'All' },
  { value: TransactionType.TRANSFER, label: 'Transfers' },
  { value: TransactionType.SWAP, label: 'Swaps' },
  { value: TransactionType.STAKE, label: 'Staking' },
] as const

type FilterValue = (typeof FILTERS)[number]['value']

/**
 * Transaction history.
 *
 * Cursor-paginated: "Load more" follows the backend cursor, and no page numbers
 * exist anywhere. Types are shown exactly as classified — `UNKNOWN` is never
 * dressed up as something more specific.
 */
export function TransactionsCard({
  walletId,
  onSelect,
  enabled = true,
}: {
  walletId: string
  onSelect: (transactionId: string) => void
  enabled?: boolean
}) {
  const [filter, setFilter] = useState<FilterValue>('ALL')

  const transactions = useTransactions(walletId, {
    type: filter === 'ALL' ? undefined : filter,
    enabled,
  })

  return (
    <Card>
      <CardHeader
        title="Transactions"
        action={
          <SegmentedControl
            label="Transaction type"
            options={FILTERS}
            value={filter}
            onChange={setFilter}
            size="sm"
          />
        }
      />

      {transactions.isLoading ? (
        <div className="px-5 py-4">
          <SkeletonRows rows={5} />
        </div>
      ) : transactions.error ? (
        <ErrorState
          error={transactions.error}
          onRetry={transactions.refetch}
          retrying={transactions.isFetching}
        />
      ) : transactions.items.length === 0 ? (
        <EmptyState
          title="No transactions"
          description={
            filter === 'ALL'
              ? 'This wallet has no transactions in the analysed history.'
              : 'No transactions of this type were found. Try another filter.'
          }
        />
      ) : (
        <>
          <TableScroll aria-label="Transactions">
            <Table>
              <thead>
                <tr>
                  <Th>Type</Th>
                  <Th>Details</Th>
                  <Th align="right">Amount</Th>
                  <Th align="right">Value</Th>
                  <Th align="right">When</Th>
                  <Th align="right">
                    <span className="sr-only">Open</span>
                  </Th>
                </tr>
              </thead>
              <tbody>
                {transactions.items.map((transaction) => (
                  <TransactionRow
                    key={transaction.id}
                    transaction={transaction}
                    onSelect={onSelect}
                  />
                ))}
              </tbody>
            </Table>
          </TableScroll>

          {transactions.hasNextPage ? (
            <div className="border-t border-line px-5 py-3">
              <Button
                variant="secondary"
                size="sm"
                onClick={transactions.fetchNextPage}
                loading={transactions.isFetchingNextPage}
              >
                Load more
              </Button>
            </div>
          ) : null}
        </>
      )}
    </Card>
  )
}

function TransactionRow({
  transaction,
  onSelect,
}: {
  transaction: Transaction
  onSelect: (transactionId: string) => void
}) {
  const asset = transaction.asset_out ?? transaction.asset_in
  const amount = transaction.amount_out ?? transaction.amount_in
  const value = transaction.value_out_usd ?? transaction.value_in_usd

  return (
    <Tr interactive onClick={() => onSelect(transaction.id)}>
      <Td>
        <div className="flex items-center gap-2">
          <span className="text-[13px] text-fg">
            {transactionTypeLabels[transaction.type]}
          </span>
          {transaction.status !== TransactionStatus.SUCCESS ? (
            <Badge
              tone={
                transaction.status === TransactionStatus.FAILED
                  ? 'negative'
                  : 'caution'
              }
            >
              {transactionStatusLabels[transaction.status]}
            </Badge>
          ) : null}
        </div>
      </Td>

      <Td>
        <div className="flex items-center gap-2.5">
          {asset ? <AssetMonogram asset={asset} size="sm" /> : null}
          <span className="text-[13px] text-fg-muted">
            {transaction.protocol ??
              transaction.counterparty ??
              asset?.name ??
              '—'}
          </span>
        </div>
      </Td>

      <Td align="right" className="text-fg-muted">
        {amount === null
          ? '—'
          : formatBalance(amount, { symbol: asset?.symbol })}
      </Td>

      <Td align="right">
        <Money value={value} unknownReason="Value at the time is unknown" />
      </Td>

      <Td align="right" className="text-[12px] text-fg-subtle">
        {formatDateTime(transaction.timestamp)}
      </Td>

      <Td align="right">
        <button
          type="button"
          onClick={(event) => {
            event.stopPropagation()
            onSelect(transaction.id)
          }}
          className="rounded-md p-1 text-fg-subtle transition-colors hover:text-fg"
          aria-label={`Open ${transactionTypeLabels[transaction.type]} details`}
        >
          <ChevronRight className="size-4" />
        </button>
      </Td>
    </Tr>
  )
}
