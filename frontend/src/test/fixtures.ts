import {
  AssetType,
  AssetVisibility,
  DataFreshness,
  DataNoticeCode,
  DataQuality,
  NoticeSeverity,
  PriceStatus,
  SyncStage,
  SyncStatus,
  ValuationStatus,
  WalletStatus,
  type Asset,
  type Portfolio,
  type Wallet,
  type WalletPosition,
} from '@/api/types'

/**
 * Fixtures built from the API contract types.
 *
 * Tests use the same DTOs as production code, so a contract change breaks the
 * fixtures rather than silently diverging from them.
 */

const NOW = '2026-08-21T12:00:00.000Z'

export function makeAsset(overrides: Partial<Asset> = {}): Asset {
  return {
    id: 'ethereum:native',
    chain_id: 'ethereum',
    contract_address: null,
    symbol: 'ETH',
    name: 'Ethereum',
    decimals: 18,
    asset_type: AssetType.NATIVE,
    icon_url: null,
    has_market_data: true,
    ...overrides,
  }
}

export function makePosition(
  overrides: Partial<WalletPosition> = {},
): WalletPosition {
  const asset = overrides.asset ?? makeAsset()

  return {
    asset,
    balance: '3.5',
    balance_raw: '3500000000000000000',
    price: {
      asset_id: asset.id,
      value_usd: '2400.00',
      currency: 'USD',
      status: PriceStatus.AVAILABLE,
      freshness: DataFreshness.FRESH,
      as_of: NOW,
      change_24h_pct: '-2.10',
    },
    value_usd: '8400.00',
    allocation_pct: '52.00',
    change_24h_pct: '-2.10',
    change_24h_usd: '-180.20',
    visibility: AssetVisibility.VISIBLE,
    valuation_status: ValuationStatus.COMPLETE,
    updated_at: NOW,
    ...overrides,
  }
}

/** A held asset with no reliable market price: balance known, value unknown. */
export function makeUnpricedPosition(): WalletPosition {
  const asset = makeAsset({
    id: 'ethereum:0xspam',
    contract_address: '0xspam',
    symbol: 'UNKNOWN',
    name: 'Unknown Token',
    asset_type: AssetType.TOKEN,
    has_market_data: false,
  })

  return makePosition({
    asset,
    balance: '500000',
    balance_raw: '500000000000000000000000',
    price: {
      asset_id: asset.id,
      value_usd: null,
      currency: 'USD',
      status: PriceStatus.UNAVAILABLE,
      freshness: DataFreshness.FRESH,
      as_of: null,
      change_24h_pct: null,
    },
    value_usd: null,
    allocation_pct: null,
    change_24h_pct: null,
    change_24h_usd: null,
    valuation_status: ValuationStatus.UNAVAILABLE,
  })
}

export function makePortfolio(overrides: Partial<Portfolio> = {}): Portfolio {
  const positions = overrides.positions ?? [makePosition()]

  return {
    wallet_id: 'wlt_1',
    currency: 'USD',
    total_value_usd: '24850.12',
    valuation_status: ValuationStatus.COMPLETE,
    data_quality: DataQuality.COMPLETE,
    data_freshness: DataFreshness.FRESH,
    change_24h_usd: '-1092.44',
    change_24h_pct: '-4.21',
    as_of: NOW,
    last_synced_at: NOW,
    calculation_version: 1,
    positions,
    visible_positions_count: positions.filter(
      (position) => position.visibility === AssetVisibility.VISIBLE,
    ).length,
    hidden_positions_count: positions.filter(
      (position) => position.visibility !== AssetVisibility.VISIBLE,
    ).length,
    exclusions: {
      unpriced_positions: 0,
      nfts_excluded: true,
      defi_positions_excluded: true,
    },
    notices: [
      {
        code: DataNoticeCode.NFTS_EXCLUDED_FROM_VALUATION,
        severity: NoticeSeverity.INFO,
      },
    ],
    ...overrides,
  }
}

export function makeWallet(overrides: Partial<Wallet> = {}): Wallet {
  return {
    id: 'wlt_1',
    chain_id: 'ethereum',
    address: '0x71C7656EC7ab88b098defB751B7401B5f6d8976F',
    label: null,
    status: WalletStatus.ACTIVE,
    sync: {
      status: SyncStatus.READY,
      stage: null,
      stages_completed: [
        SyncStage.FETCHING_BALANCES,
        SyncStage.FETCHING_TRANSACTIONS,
        SyncStage.NORMALIZING_ASSETS,
        SyncStage.FETCHING_PRICES,
        SyncStage.CALCULATING_PORTFOLIO,
        SyncStage.PREPARING_ANALYSIS,
      ],
      started_at: NOW,
      completed_at: NOW,
      last_synced_at: NOW,
      data_freshness: DataFreshness.FRESH,
      error: null,
    },
    created_at: NOW,
    updated_at: NOW,
    ...overrides,
  }
}
