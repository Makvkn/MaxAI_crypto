import { KnownChainId, type Chain, type ChainId } from '@/api/types'

/**
 * Chain presentation layer.
 *
 * Adding a chain is a matter of adding an entry here — no onboarding, wallet or
 * portfolio code changes. This file holds presentation metadata only:
 *
 *   - display name and native symbol
 *   - accent colour used for monograms and chips
 *   - an address placeholder and a *format hint* for inline UX validation
 *
 * The address pattern is a courtesy check so the user gets immediate feedback.
 * The backend validates authoritatively and may reject an address this pattern
 * accepts; `INVALID_WALLET_ADDRESS` is always handled as the real answer.
 *
 * When the backend exposes reference data for chains, this registry becomes a
 * presentation supplement to that response rather than the source of the list.
 */
export interface ChainPresentation {
  id: ChainId
  name: string
  nativeSymbol: string
  /** One line describing what the user is analysing. */
  summary: string
  accent: string
  addressPlaceholder: string
  addressPattern: RegExp | null
  addressHint: string
}

export const SUPPORTED_CHAINS: readonly ChainPresentation[] = [
  {
    id: KnownChainId.ETHEREUM,
    name: 'Ethereum',
    nativeSymbol: 'ETH',
    summary: 'ETH and ERC-20 balances',
    accent: '#7B8CFF',
    addressPlaceholder: '0x71C7656EC7ab88b098defB751B7401B5f6d8976F',
    addressPattern: /^0x[a-fA-F0-9]{40}$/,
    addressHint: 'An Ethereum address starts with 0x and has 42 characters.',
  },
  {
    id: KnownChainId.BITCOIN,
    name: 'Bitcoin',
    nativeSymbol: 'BTC',
    summary: 'BTC balances and transfers',
    accent: '#F7931A',
    addressPlaceholder: 'bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq',
    addressPattern:
      /^(bc1[a-zA-HJ-NP-Z0-9]{11,71}|[13][a-km-zA-HJ-NP-Z1-9]{25,34})$/,
    addressHint: 'Bitcoin addresses start with bc1, 1 or 3.',
  },
  {
    id: KnownChainId.BNB,
    name: 'BNB Chain',
    nativeSymbol: 'BNB',
    summary: 'BNB and BEP-20 balances',
    accent: '#F0B90B',
    addressPlaceholder: '0x8894E0a0c962CB723c1976a4421c95949bE2D4E3',
    addressPattern: /^0x[a-fA-F0-9]{40}$/,
    addressHint: 'A BNB Chain address starts with 0x and has 42 characters.',
  },
  {
    id: KnownChainId.SOLANA,
    name: 'Solana',
    nativeSymbol: 'SOL',
    summary: 'SOL and SPL token balances',
    accent: '#14F195',
    addressPlaceholder: '7xKXtg2CW87d9TXJd1nQ9AaLQ9Nf6WcYbeRbNZbDbBoE',
    addressPattern: /^[1-9A-HJ-NP-Za-km-z]{32,44}$/,
    addressHint: 'Solana addresses are 32–44 base58 characters.',
  },
  {
    id: KnownChainId.LITECOIN,
    name: 'Litecoin',
    nativeSymbol: 'LTC',
    summary: 'LTC balances and transfers',
    accent: '#A6A9AA',
    addressPlaceholder: 'ltc1qhz8x0m3sm9dr0m5rvxq4c0m6l0ndsz5f0kj2q4',
    addressPattern:
      /^(ltc1[a-zA-HJ-NP-Z0-9]{11,71}|[LM3][a-km-zA-HJ-NP-Z1-9]{26,34})$/,
    addressHint: 'Litecoin addresses start with ltc1, L or M.',
  },
  {
    id: KnownChainId.XRPL,
    name: 'XRP Ledger',
    nativeSymbol: 'XRP',
    summary: 'XRP balances and payments',
    accent: '#24C3D9',
    addressPlaceholder: 'rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn',
    addressPattern: /^r[1-9A-HJ-NP-Za-km-z]{24,34}$/,
    addressHint: 'XRP Ledger addresses start with r.',
  },
  {
    // TRON is the chain; TRX is its native token.
    id: KnownChainId.TRON,
    name: 'TRON',
    nativeSymbol: 'TRX',
    summary: 'TRX and TRC-20 balances',
    accent: '#FF3B3B',
    addressPlaceholder: 'TJRyWwFs9wTFGZg3JbrVriFbNfCug5tDeC',
    addressPattern: /^T[1-9A-HJ-NP-Za-km-z]{33}$/,
    addressHint: 'TRON addresses start with T and have 34 characters.',
  },
  {
    id: KnownChainId.DOGECOIN,
    name: 'Dogecoin',
    nativeSymbol: 'DOGE',
    summary: 'DOGE balances and transfers',
    accent: '#C3A634',
    addressPlaceholder: 'DH5yaieqoZN36fDVciNyRueRGvGLR3mr7L',
    addressPattern: /^[DA9][1-9A-HJ-NP-Za-km-z]{25,34}$/,
    addressHint: 'Dogecoin addresses start with D, A or 9.',
  },
] as const

const FALLBACK: ChainPresentation = {
  id: 'unknown',
  name: 'Unknown network',
  nativeSymbol: '—',
  summary: 'Network metadata is not available',
  accent: '#69727E',
  addressPlaceholder: '',
  addressPattern: null,
  addressHint: '',
}

const BY_ID = new Map(SUPPORTED_CHAINS.map((chain) => [chain.id, chain]))

/**
 * Presentation for a chain id. A chain the backend added but the frontend does
 * not recognise degrades to a neutral presentation instead of crashing.
 */
export function chainPresentation(chainId: ChainId): ChainPresentation {
  return BY_ID.get(chainId) ?? { ...FALLBACK, id: chainId }
}

export function isSupportedChain(chainId: ChainId): boolean {
  return BY_ID.has(chainId)
}

/** Domain projection, matching what a `GET /chains` endpoint would return. */
export function toChain(presentation: ChainPresentation): Chain {
  return {
    id: presentation.id,
    name: presentation.name,
    native_asset_symbol: presentation.nativeSymbol,
    is_supported: true,
  }
}
