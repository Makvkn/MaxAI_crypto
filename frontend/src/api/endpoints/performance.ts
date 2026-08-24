import type { RequestOptions } from '../client'
import { http } from '../http'
import type { PerformancePeriod, PortfolioPerformance, WalletIdPath } from '../types'

/**
 * `GET /api/v1/wallets/:id/performance?period=...`
 *
 * Returns both the period result and the snapshot series used to draw the
 * historical chart, so the frontend never reconstructs history.
 */
export const apiGetPerformance = (
  { walletId }: WalletIdPath,
  period: PerformancePeriod,
  options?: RequestOptions,
): Promise<PortfolioPerformance> =>
  http.get<PortfolioPerformance>(
    `/wallets/${encodeURIComponent(walletId)}/performance`,
    { period },
    options,
  )
