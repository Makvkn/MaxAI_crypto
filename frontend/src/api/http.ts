import { apiRoot, env } from '@/app/config/env'
import { HttpClient } from './client'
import { tokenStore } from './tokenStore'

/**
 * Shared HTTP client for REST and SSE.
 *
 * Feature modules call named `api*` functions; those functions use this
 * singleton rather than receiving `HttpClient` through factories.
 */
export const http = new HttpClient({
  baseUrl: apiRoot,
  timeoutMs: env.apiTimeoutMs,
  tokens: tokenStore,
})
