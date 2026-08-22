import type { HttpClient } from '../client'
import type { PortfolioApi } from '../contract'
import type { Portfolio } from '../types'

/** `GET /api/v1/wallets/:id/portfolio` — backend-computed valuation. */
export function createPortfolioApi(http: HttpClient): PortfolioApi {
  return {
    get: (walletId, options) =>
      http.get<Portfolio>(
        `/wallets/${encodeURIComponent(walletId)}/portfolio`,
        undefined,
        options,
      ),
  }
}
