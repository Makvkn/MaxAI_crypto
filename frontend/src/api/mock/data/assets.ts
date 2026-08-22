import type { Asset, ChainId } from '../../types'
import type { MockAssetSpec } from './catalog'

/** Assets are identified by chain + contract, never by symbol alone. */
export function assetId(chainId: ChainId, spec: MockAssetSpec): string {
  return `${chainId}:${spec.contract_address ?? 'native'}`
}

export function toAsset(chainId: ChainId, spec: MockAssetSpec): Asset {
  return {
    id: assetId(chainId, spec),
    chain_id: chainId,
    contract_address: spec.contract_address,
    symbol: spec.symbol,
    name: spec.name,
    decimals: spec.decimals,
    asset_type: spec.asset_type,
    icon_url: null,
    has_market_data: spec.price_usd !== null,
  }
}
