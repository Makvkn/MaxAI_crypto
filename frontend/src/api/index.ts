import { apiRoot, env } from '@/app/config/env'
import { HttpClient } from './client'
import type { MaxAIApi } from './contract'
import { createAuthApi, installRefreshHandler } from './endpoints/auth'
import { createConversationsApi } from './endpoints/conversations'
import { createPerformanceApi } from './endpoints/performance'
import { createPortfolioApi } from './endpoints/portfolio'
import { createScenariosApi } from './endpoints/scenarios'
import { createTransactionsApi } from './endpoints/transactions'
import { createAiUsageApi } from './endpoints/usage'
import { createWalletsApi } from './endpoints/wallets'
import { createMockApi } from './mock'
import { createTokenStore, type TokenStore } from './tokenStore'

/**
 * API composition root.
 *
 * `VITE_API_MODE` selects the adapter. Everything above this module depends on
 * the `MaxAIApi` interface only, so the switch is invisible to features.
 */

export const tokenStore: TokenStore = createTokenStore()

export function createHttpApi(options?: {
  baseUrl?: string
  tokens?: TokenStore
}): MaxAIApi {
  const http = new HttpClient({
    baseUrl: options?.baseUrl ?? apiRoot,
    timeoutMs: env.apiTimeoutMs,
    tokens: options?.tokens ?? tokenStore,
  })

  installRefreshHandler(http)

  return {
    auth: createAuthApi(http),
    wallets: createWalletsApi(http),
    portfolio: createPortfolioApi(http),
    performance: createPerformanceApi(http),
    transactions: createTransactionsApi(http),
    conversations: createConversationsApi(http),
    ai: createAiUsageApi(http),
    scenarios: createScenariosApi(http),
  }
}

export function createApi(): MaxAIApi {
  return env.apiMode === 'mock' ? createMockApi() : createHttpApi()
}

/** The application-wide client. Feature hooks are its only consumers. */
export const api: MaxAIApi = createApi()

export type { MaxAIApi } from './contract'
export { ApiError } from './errors'
export * from './types'
