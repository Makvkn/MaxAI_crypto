import type { HttpClient } from '../client'
import type { ScenariosApi } from '../contract'
import type { ScenarioResult } from '../types'

/**
 * `POST /api/v1/ai/scenarios`
 *
 * The backend runs the deterministic calculation and attaches the AI
 * explanation. No part of the scenario is computed in the browser.
 */
export function createScenariosApi(http: HttpClient): ScenariosApi {
  return {
    simulate: (request, options) =>
      http.post<ScenarioResult>('/ai/scenarios', request, options),
  }
}
