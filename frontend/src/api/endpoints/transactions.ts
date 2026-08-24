import type { RequestOptions } from '../client'
import { http } from '../http'
import type {
  CursorPage,
  Transaction,
  TransactionIdPath,
  TransactionListParams,
  WalletIdPath,
} from '../types'

/**
 * `GET /api/v1/wallets/:id/transactions`
 *
 * Cursor pagination only — there is no page/offset parameter.
 */

export const apiGetTransactions = (
  { walletId }: WalletIdPath,
  params?: TransactionListParams,
  options?: RequestOptions,
): Promise<CursorPage<Transaction>> =>
  http.get<CursorPage<Transaction>>(
    `/wallets/${encodeURIComponent(walletId)}/transactions`,
    {
      limit: params?.limit,
      cursor: params?.cursor,
      type: params?.type,
    },
    options,
  )

export const apiGetTransaction = (
  { walletId, transactionId }: WalletIdPath & TransactionIdPath,
  options?: RequestOptions,
): Promise<Transaction> =>
  http.get<Transaction>(
    `/wallets/${encodeURIComponent(walletId)}/transactions/${encodeURIComponent(transactionId)}`,
    undefined,
    options,
  )
