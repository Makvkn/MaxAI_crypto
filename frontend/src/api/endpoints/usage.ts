import type { RequestOptions } from '../client'
import { http } from '../http'
import type { AIUsage } from '../types'

/**
 * `GET /api/v1/ai/usage`
 *
 * Display only. The daily limit is enforced by the backend; a client-side
 * counter is never treated as protection.
 */
export const apiGetAiUsage = (options?: RequestOptions): Promise<AIUsage> =>
  http.get<AIUsage>('/ai/usage', undefined, options)
