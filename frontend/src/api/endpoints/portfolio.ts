import type { RequestOptions } from '../client'
import { http } from '../http'
import type { Portfolio, WalletIdPath } from '../types'

/** `GET /api/v1/wallets/:id/portfolio` — backend-computed valuation. */
export const apiGetPortfolio = (
  { walletId }: WalletIdPath,
  options?: RequestOptions,
): Promise<Portfolio> =>
  http.get<Portfolio>(
    `/wallets/${encodeURIComponent(walletId)}/portfolio`,
    undefined,
    options,
  )
