/**
 * Chains are domain entities, not hardcoded branches. Adding a chain is a
 * matter of metadata; it must never require touching onboarding logic.
 *
 * PROVISIONAL CONTRACT LAYER — see `src/api/types/index.ts`.
 */

/** Chains supported by the MVP. TRON is the chain; TRX is its native token. */
export const KnownChainId = {
  ETHEREUM: 'ethereum',
  BITCOIN: 'bitcoin',
  BNB: 'bnb',
  SOLANA: 'solana',
  LITECOIN: 'litecoin',
  XRPL: 'xrpl',
  TRON: 'tron',
  DOGECOIN: 'dogecoin',
} as const
export type KnownChainId = (typeof KnownChainId)[keyof typeof KnownChainId]

/** Open union so a chain added by the backend does not break the client. */
export type ChainId = KnownChainId | (string & {})

export interface Chain {
  id: ChainId
  /** Human-readable chain name, e.g. "XRP Ledger". */
  name: string
  /** Native token symbol, e.g. `TRX` for TRON. */
  native_asset_symbol: string
  is_supported: boolean
}
