import type { HttpClient } from '../client'
import type { AiUsageApi } from '../contract'
import type { AIUsage } from '../types'

/**
 * `GET /api/v1/ai/usage`
 *
 * Display only. The daily limit is enforced by the backend; a client-side
 * counter is never treated as protection.
 */
export function createAiUsageApi(http: HttpClient): AiUsageApi {
  return {
    getUsage: (options) => http.get<AIUsage>('/ai/usage', undefined, options),
  }
}
