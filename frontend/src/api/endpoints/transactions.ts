import type { HttpClient } from '../client'
import type { TransactionsApi } from '../contract'
import type { CursorPage, Transaction } from '../types'

/**
 * `GET /api/v1/wallets/:id/transactions`
 *
 * Cursor pagination only — there is no page/offset parameter.
 */
export function createTransactionsApi(http: HttpClient): TransactionsApi {
  return {
    list: (walletId, params, options) =>
      http.get<CursorPage<Transaction>>(
        `/wallets/${encodeURIComponent(walletId)}/transactions`,
        {
          limit: params?.limit,
          cursor: params?.cursor,
          type: params?.type,
        },
        options,
      ),

    get: (walletId, transactionId, options) =>
      http.get<Transaction>(
        `/wallets/${encodeURIComponent(walletId)}/transactions/${encodeURIComponent(transactionId)}`,
        undefined,
        options,
      ),
  }
}
