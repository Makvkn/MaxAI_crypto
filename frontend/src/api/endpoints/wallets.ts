import type { HttpClient } from '../client'
import type { WalletsApi } from '../contract'
import type { CursorPage, Wallet } from '../types'

/**
 * `/api/v1/wallets`
 *
 * `create` returns as soon as the wallet row exists and the initial sync job
 * is enqueued. Callers must observe `wallet.sync` rather than assuming a
 * portfolio is available.
 */
export function createWalletsApi(http: HttpClient): WalletsApi {
  return {
    list: (params, options) =>
      http.get<CursorPage<Wallet>>(
        '/wallets',
        { limit: params?.limit, cursor: params?.cursor },
        options,
      ),

    get: (walletId, options) =>
      http.get<Wallet>(`/wallets/${encodeURIComponent(walletId)}`, undefined, options),

    create: (request, options) => http.post<Wallet>('/wallets', request, options),
  }
}
