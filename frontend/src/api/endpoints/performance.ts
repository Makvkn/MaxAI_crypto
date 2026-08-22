import type { HttpClient } from '../client'
import type { PerformanceApi } from '../contract'
import type { PortfolioPerformance } from '../types'

/**
 * `GET /api/v1/wallets/:id/performance?period=...`
 *
 * Returns both the period result and the snapshot series used to draw the
 * historical chart, so the frontend never reconstructs history.
 */
export function createPerformanceApi(http: HttpClient): PerformanceApi {
  return {
    get: (walletId, period, options) =>
      http.get<PortfolioPerformance>(
        `/wallets/${encodeURIComponent(walletId)}/performance`,
        { period },
        options,
      ),
  }
}
