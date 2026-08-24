import { useMutation, useQueryClient } from '@tanstack/react-query'
import { apiSimulateScenario } from '@/api'
import { ScenarioType, type ScenarioResult } from '@/api/types'
import { analytics } from '@/lib/analytics/analytics'
import { queryKeys } from '@/lib/query/queryKeys'

/**
 * Scenario simulation.
 *
 * The request carries an asset and a signed percentage; the backend runs the
 * deterministic calculation and returns the structured result together with the
 * AI explanation. Nothing about the impact is computed here.
 */
export function useScenarioSimulation(walletId: string) {
  const queryClient = useQueryClient()

  return useMutation<
    ScenarioResult,
    unknown,
    { assetId: string; assetSymbol: string; changePct: string }
  >({
    mutationFn: ({ assetId, changePct }) =>
      apiSimulateScenario({
        wallet_id: walletId,
        type: ScenarioType.ASSET_PRICE_CHANGE,
        asset_id: assetId,
        change_pct: changePct,
      }),
    onSuccess: (_result, variables) => {
      analytics.track('scenario_used', {
        wallet_id: walletId,
        asset_symbol: variables.assetSymbol,
        change_pct: variables.changePct,
      })
      // A scenario consumes one AI operation.
      void queryClient.invalidateQueries({ queryKey: queryKeys.aiUsage() })
    },
    onSettled: () => {
      void queryClient.invalidateQueries({ queryKey: queryKeys.aiUsage() })
    },
  })
}
