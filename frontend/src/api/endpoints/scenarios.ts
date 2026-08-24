import type { RequestOptions } from '../client'
import { http } from '../http'
import type { ScenarioRequest, ScenarioResult } from '../types'

/**
 * `POST /api/v1/ai/scenarios`
 *
 * The backend runs the deterministic calculation and attaches the AI
 * explanation. No part of the scenario is computed in the browser.
 */
export const apiSimulateScenario = (
  request: ScenarioRequest,
  options?: RequestOptions,
): Promise<ScenarioResult> =>
  http.post<ScenarioResult>('/ai/scenarios', request, options)
