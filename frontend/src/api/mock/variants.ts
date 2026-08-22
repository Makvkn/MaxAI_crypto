/**
 * Deterministic mock scenarios.
 *
 * The address a developer types selects which backend condition to simulate,
 * so every data-quality and failure path is reachable without code changes:
 *
 *   ...partial       -> a visible asset has no price; valuation is PARTIAL
 *   ...stale         -> data is old; freshness is STALE
 *   ...verystale     -> freshness is VERY_STALE
 *   ...fail          -> initial sync ends in FAILED
 *   ...syncpartial   -> initial sync ends in PARTIAL
 *   ...unavailable   -> portfolio valuation is UNAVAILABLE
 *   ...empty         -> wallet holds nothing
 *   ...nohistory     -> no snapshots yet, so performance is UNAVAILABLE
 *   ...slow          -> long-running initial sync
 *
 * Anything else behaves as a healthy wallet.
 */
export interface MockVariant {
  partialValuation: boolean
  stale: boolean
  veryStale: boolean
  syncFails: boolean
  syncPartial: boolean
  portfolioUnavailable: boolean
  empty: boolean
  noHistory: boolean
  slowSync: boolean
}

const NONE: MockVariant = {
  partialValuation: false,
  stale: false,
  veryStale: false,
  syncFails: false,
  syncPartial: false,
  portfolioUnavailable: false,
  empty: false,
  noHistory: false,
  slowSync: false,
}

export function resolveVariant(address: string): MockVariant {
  const key = address.toLowerCase()
  const has = (token: string) => key.includes(token)

  return {
    ...NONE,
    partialValuation: has('partial') && !has('syncpartial'),
    stale: has('stale') && !has('verystale'),
    veryStale: has('verystale'),
    syncFails: has('fail'),
    syncPartial: has('syncpartial'),
    portfolioUnavailable: has('unavailable'),
    empty: has('empty'),
    noHistory: has('nohistory'),
    slowSync: has('slow'),
  }
}
