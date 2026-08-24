import type { RequestOptions } from '../client'
import { http } from '../http'
import type {
  CreateWalletRequest,
  CursorPage,
  CursorParams,
  Wallet,
  WalletIdPath,
} from '../types'

/**
 * `/api/v1/wallets`
 *
 * `apiCreateWallet` returns as soon as the wallet row exists and the initial
 * sync job is enqueued. Callers must observe `wallet.sync` rather than assuming
 * a portfolio is available.
 */

export const apiGetWallets = (
  params?: CursorParams,
  options?: RequestOptions,
): Promise<CursorPage<Wallet>> =>
  http.get<CursorPage<Wallet>>(
    '/wallets',
    { limit: params?.limit, cursor: params?.cursor },
    options,
  )

export const apiGetWallet = (
  { walletId }: WalletIdPath,
  options?: RequestOptions,
): Promise<Wallet> =>
  http.get<Wallet>(`/wallets/${encodeURIComponent(walletId)}`, undefined, options)

export const apiCreateWallet = (
  request: CreateWalletRequest,
  options?: RequestOptions,
): Promise<Wallet> => http.post<Wallet>('/wallets', request, options)
