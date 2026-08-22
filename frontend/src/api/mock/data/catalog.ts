import {
  AssetType,
  AssetVisibility,
  KnownChainId,
  type ChainId,
} from '../../types'

/**
 * Holdings the mock backend reports per chain.
 *
 * `price_usd: null` models an asset with no reliable market-data mapping. Such
 * an asset still has a balance, and must never be valued at zero.
 */
export interface MockAssetSpec {
  symbol: string
  name: string
  decimals: number
  contract_address: string | null
  asset_type: AssetType
  price_usd: string | null
  change_24h_pct: string | null
  balance: string
  visibility: AssetVisibility
}

const native = (
  spec: Omit<MockAssetSpec, 'asset_type' | 'contract_address' | 'visibility'>,
): MockAssetSpec => ({
  ...spec,
  contract_address: null,
  asset_type: AssetType.NATIVE,
  visibility: AssetVisibility.VISIBLE,
})

const token = (spec: Omit<MockAssetSpec, 'asset_type'>): MockAssetSpec => ({
  ...spec,
  asset_type: AssetType.TOKEN,
})

/** Spam and dust the backend classifier keeps out of the main asset list. */
const NOISE: MockAssetSpec[] = [
  token({
    symbol: 'USDT-CLAIM',
    name: 'Claim 5,000 USDT at usdt-reward.xyz',
    decimals: 18,
    contract_address: '0x9f1a4c0b7e2d5a8f3b6c1d4e7a0b3c6d9e2f5a81',
    price_usd: null,
    change_24h_pct: null,
    balance: '5000',
    visibility: AssetVisibility.HIDDEN_SPAM,
  }),
  token({
    symbol: 'AIRDROP',
    name: 'Airdrop Voucher',
    decimals: 18,
    contract_address: '0x3c7d9e1f4a6b8c0d2e5f7a9b1c3d5e7f9a1b3c5d',
    price_usd: null,
    change_24h_pct: null,
    balance: '1200000',
    visibility: AssetVisibility.HIDDEN_SPAM,
  }),
  token({
    symbol: 'SHIB',
    name: 'Shiba Inu',
    decimals: 18,
    contract_address: '0x95ad61b0a150d79219dcf64e1e6cc01f0b64c4ce',
    price_usd: '0.0000241',
    change_24h_pct: '-3.10',
    balance: '41230',
    visibility: AssetVisibility.HIDDEN_DUST,
  }),
  token({
    symbol: 'CRV',
    name: 'Curve DAO Token',
    decimals: 18,
    contract_address: '0xd533a949740bb3306d119cc777fa900ba034cd52',
    price_usd: '0.6210',
    change_24h_pct: '-2.44',
    balance: '0.8312',
    visibility: AssetVisibility.HIDDEN_DUST,
  }),
  token({
    symbol: 'XYZ',
    name: 'Unknown Token',
    decimals: 18,
    contract_address: '0x7b2e4f6a8c0d1e3f5a7b9c1d3e5f7a9b1c3d5e7f',
    price_usd: null,
    change_24h_pct: null,
    balance: '500000',
    visibility: AssetVisibility.UNKNOWN,
  }),
]

const CHAIN_HOLDINGS: Record<KnownChainId, MockAssetSpec[]> = {
  [KnownChainId.ETHEREUM]: [
    native({
      symbol: 'ETH',
      name: 'Ethereum',
      decimals: 18,
      price_usd: '4210.52',
      change_24h_pct: '-3.84',
      balance: '3.0700',
    }),
    token({
      symbol: 'WBTC',
      name: 'Wrapped Bitcoin',
      decimals: 8,
      contract_address: '0x2260fac5e5542a773aa44fbcfedf7c193bc2c599',
      price_usd: '96420.10',
      change_24h_pct: '-1.92',
      balance: '0.05410',
      visibility: AssetVisibility.VISIBLE,
    }),
    token({
      symbol: 'LINK',
      name: 'Chainlink',
      decimals: 18,
      contract_address: '0x514910771af9ca656af840dff83e8264ecf986ca',
      price_usd: '21.84',
      change_24h_pct: '-5.62',
      balance: '102.4000',
      visibility: AssetVisibility.VISIBLE,
    }),
    token({
      symbol: 'AAVE',
      name: 'Aave',
      decimals: 18,
      contract_address: '0x7fc66500c84a76ad7e9c93437bfc5ac33e2ddae9',
      price_usd: '268.40',
      change_24h_pct: '-2.11',
      balance: '6.2000',
      visibility: AssetVisibility.VISIBLE,
    }),
    token({
      symbol: 'UNI',
      name: 'Uniswap',
      decimals: 18,
      contract_address: '0x1f9840a85d5af5bf1d1762f925bdaddc4201f984',
      price_usd: '12.06',
      change_24h_pct: '-4.35',
      balance: '118.0000',
      visibility: AssetVisibility.VISIBLE,
    }),
    token({
      symbol: 'USDC',
      name: 'USD Coin',
      decimals: 6,
      contract_address: '0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48',
      price_usd: '1.0001',
      change_24h_pct: '0.01',
      balance: '1204.5100',
      visibility: AssetVisibility.VISIBLE,
    }),
    token({
      symbol: 'ENS',
      name: 'Ethereum Name Service',
      decimals: 18,
      contract_address: '0xc18360217d8f7ab5e7c516566761ea12ce7f9d72',
      price_usd: '21.50',
      change_24h_pct: '-1.20',
      balance: '8.3400',
      visibility: AssetVisibility.VISIBLE,
    }),
    ...NOISE,
  ],

  [KnownChainId.BITCOIN]: [
    native({
      symbol: 'BTC',
      name: 'Bitcoin',
      decimals: 8,
      price_usd: '96420.10',
      change_24h_pct: '-1.92',
      balance: '0.41820000',
    }),
  ],

  [KnownChainId.BNB]: [
    native({
      symbol: 'BNB',
      name: 'BNB',
      decimals: 18,
      price_usd: '712.40',
      change_24h_pct: '-2.40',
      balance: '12.4000',
    }),
    token({
      symbol: 'CAKE',
      name: 'PancakeSwap',
      decimals: 18,
      contract_address: '0x0e09fabb73bd3ade0a17ecc321fd13a19e81ce82',
      price_usd: '2.61',
      change_24h_pct: '-3.80',
      balance: '420.0000',
      visibility: AssetVisibility.VISIBLE,
    }),
    token({
      symbol: 'USDT',
      name: 'Tether USD',
      decimals: 18,
      contract_address: '0x55d398326f99059ff775485246999027b3197955',
      price_usd: '1.0000',
      change_24h_pct: '0.00',
      balance: '803.2200',
      visibility: AssetVisibility.VISIBLE,
    }),
  ],

  [KnownChainId.SOLANA]: [
    native({
      symbol: 'SOL',
      name: 'Solana',
      decimals: 9,
      price_usd: '214.86',
      change_24h_pct: '-6.10',
      balance: '84.2000',
    }),
    token({
      symbol: 'USDC',
      name: 'USD Coin',
      decimals: 6,
      contract_address: 'EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v',
      price_usd: '1.0001',
      change_24h_pct: '0.01',
      balance: '512.4000',
      visibility: AssetVisibility.VISIBLE,
    }),
    token({
      symbol: 'JUP',
      name: 'Jupiter',
      decimals: 6,
      contract_address: 'JUPyiwrYJFskUPiHa7hkeR8VUtAeFoSYbKedZNsDvCN',
      price_usd: '0.9400',
      change_24h_pct: '-7.20',
      balance: '1240.0000',
      visibility: AssetVisibility.VISIBLE,
    }),
    token({
      symbol: 'FREEPUMP',
      name: 'Free Pump Reward',
      decimals: 6,
      contract_address: '8xPuMp1NkR2vQz7YbWq3sTdF6mJhLcXeA9gN4rZ5uK1t',
      price_usd: null,
      change_24h_pct: null,
      balance: '9000000',
      visibility: AssetVisibility.HIDDEN_SPAM,
    }),
  ],

  [KnownChainId.LITECOIN]: [
    native({
      symbol: 'LTC',
      name: 'Litecoin',
      decimals: 8,
      price_usd: '118.42',
      change_24h_pct: '-1.10',
      balance: '32.50000000',
    }),
  ],

  [KnownChainId.XRPL]: [
    native({
      symbol: 'XRP',
      name: 'XRP',
      decimals: 6,
      price_usd: '2.4400',
      change_24h_pct: '3.20',
      balance: '5420.000000',
    }),
  ],

  [KnownChainId.TRON]: [
    native({
      symbol: 'TRX',
      name: 'TRON',
      decimals: 6,
      price_usd: '0.2612',
      change_24h_pct: '1.40',
      balance: '48200.000000',
    }),
    token({
      symbol: 'USDT',
      name: 'Tether USD',
      decimals: 6,
      contract_address: 'TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t',
      price_usd: '1.0000',
      change_24h_pct: '0.00',
      balance: '2400.000000',
      visibility: AssetVisibility.VISIBLE,
    }),
  ],

  [KnownChainId.DOGECOIN]: [
    native({
      symbol: 'DOGE',
      name: 'Dogecoin',
      decimals: 8,
      price_usd: '0.3184',
      change_24h_pct: '-4.20',
      balance: '62400.00000000',
    }),
  ],
}

/** A visible position with no market price, used by the `partial` variant. */
export const UNPRICED_VISIBLE_ASSET: MockAssetSpec = {
  symbol: 'MTKN',
  name: 'Mystery Token',
  decimals: 18,
  contract_address: '0x5f8a2b4c6d8e0f2a4b6c8d0e2f4a6b8c0d2e4f6a',
  asset_type: AssetType.TOKEN,
  price_usd: null,
  change_24h_pct: null,
  balance: '500000',
  visibility: AssetVisibility.VISIBLE,
}

export function holdingsForChain(chainId: ChainId): MockAssetSpec[] {
  return CHAIN_HOLDINGS[chainId as KnownChainId] ?? []
}
