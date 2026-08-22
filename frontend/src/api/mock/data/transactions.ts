import {
  AssetVisibility,
  KnownChainId,
  TransactionStatus,
  TransactionType,
  type ChainId,
  type Transaction,
  type Wallet,
  type WalletPosition,
} from '../../types'
import * as d from '../support/decimal'
import { createRandom, hashString, pick } from '../support/random'

/**
 * Canonical transactions — MOCK BACKEND SIMULATION of
 * `TransactionNormalizer` + `TransactionClassifier`.
 *
 * Types are decided here, i.e. server-side. `UNKNOWN` is emitted on purpose so
 * the UI has to render it as unknown.
 */

const TRANSACTION_COUNT = 124

const TYPE_WEIGHTS: TransactionType[] = [
  ...Array<TransactionType>(9).fill(TransactionType.TRANSFER),
  ...Array<TransactionType>(7).fill(TransactionType.SWAP),
  ...Array<TransactionType>(2).fill(TransactionType.APPROVE),
  TransactionType.STAKE,
  TransactionType.UNSTAKE,
  TransactionType.CLAIM,
  ...Array<TransactionType>(2).fill(TransactionType.CONTRACT_INTERACTION),
  ...Array<TransactionType>(2).fill(TransactionType.UNKNOWN),
]

const PROTOCOLS: Partial<Record<KnownChainId, string[]>> = {
  [KnownChainId.ETHEREUM]: ['Uniswap', 'Aave', 'Lido', 'Curve', '1inch'],
  [KnownChainId.BNB]: ['PancakeSwap', 'Venus'],
  [KnownChainId.SOLANA]: ['Jupiter', 'Orca', 'Marinade'],
  [KnownChainId.TRON]: ['SunSwap', 'JustLend'],
}

const EXPLORERS: Record<KnownChainId, string> = {
  [KnownChainId.ETHEREUM]: 'https://etherscan.io/tx/',
  [KnownChainId.BITCOIN]: 'https://mempool.space/tx/',
  [KnownChainId.BNB]: 'https://bscscan.com/tx/',
  [KnownChainId.SOLANA]: 'https://solscan.io/tx/',
  [KnownChainId.LITECOIN]: 'https://blockchair.com/litecoin/transaction/',
  [KnownChainId.XRPL]: 'https://livenet.xrpl.org/transactions/',
  [KnownChainId.TRON]: 'https://tronscan.org/#/transaction/',
  [KnownChainId.DOGECOIN]: 'https://blockchair.com/dogecoin/transaction/',
}

function explorerUrl(chainId: ChainId, hash: string): string | null {
  const base = EXPLORERS[chainId as KnownChainId]
  return base ? `${base}${hash}` : null
}

function hexHash(random: () => number, length: number): string {
  const digits = '0123456789abcdef'
  let out = ''
  for (let index = 0; index < length; index += 1) {
    out += digits[Math.floor(random() * 16)]
  }
  return out
}

function address(random: () => number, chainId: ChainId): string {
  if (chainId === KnownChainId.ETHEREUM || chainId === KnownChainId.BNB) {
    return `0x${hexHash(random, 40)}`
  }
  if (chainId === KnownChainId.TRON) return `T${hexHash(random, 33)}`
  return hexHash(random, 34)
}

export function buildTransactions(
  wallet: Wallet,
  positions: WalletPosition[],
  now: Date,
): Transaction[] {
  const tradable = positions.filter(
    (position) =>
      position.visibility === AssetVisibility.VISIBLE &&
      position.price?.value_usd,
  )
  if (tradable.length === 0) return []

  const native = tradable[0] as WalletPosition
  const random = createRandom(hashString(`${wallet.id}:transactions`))
  const transactions: Transaction[] = []

  for (let index = 0; index < TRANSACTION_COUNT; index += 1) {
    const type = pick(random, TYPE_WEIGHTS)
    const hash =
      wallet.chain_id === KnownChainId.SOLANA
        ? hexHash(random, 64)
        : `0x${hexHash(random, 64)}`
    const timestamp = new Date(
      now.getTime() - index * (4.5 * 3_600_000) - Math.floor(random() * 7_200_000),
    )
    const failed = random() < 0.05
    const protocolPool = PROTOCOLS[wallet.chain_id as KnownChainId]
    const protocol =
      protocolPool && type !== TransactionType.TRANSFER
        ? pick(random, protocolPool)
        : null

    const feeAmount = (0.00012 + random() * 0.0021).toFixed(6)
    const feeValue = native.price?.value_usd
      ? d.multiply(feeAmount, native.price.value_usd, 2)
      : null

    const base: Transaction = {
      id: `tx_${wallet.id}_${index}`,
      wallet_id: wallet.id,
      chain_id: wallet.chain_id,
      tx_hash: hash,
      block_number: 21_400_000 - index * 317,
      timestamp: timestamp.toISOString(),
      status: failed ? TransactionStatus.FAILED : TransactionStatus.SUCCESS,
      type,
      from_address: wallet.address,
      to_address: address(random, wallet.chain_id),
      asset_in: null,
      amount_in: null,
      value_in_usd: null,
      asset_out: null,
      amount_out: null,
      value_out_usd: null,
      fee_asset: native.asset,
      fee_amount: feeAmount,
      fee_value_usd: feeValue,
      protocol,
      counterparty: protocol,
      explorer_url: explorerUrl(wallet.chain_id, hash),
      created_at: timestamp.toISOString(),
      updated_at: timestamp.toISOString(),
    }

    transactions.push(classify(base, type, tradable, random))
  }

  return transactions
}

/** Assigns the canonical type and the amounts that type implies. */
function classify(
  transaction: Transaction,
  requestedType: TransactionType,
  tradable: WalletPosition[],
  random: () => number,
): Transaction {
  const subject = pick(random, tradable)
  const counterAssets = tradable.filter(
    (position) => position.asset.id !== subject.asset.id,
  )

  // A single-asset chain cannot produce a swap: the classifier reports a
  // transfer rather than inventing a counter-asset.
  const type =
    requestedType === TransactionType.SWAP && counterAssets.length === 0
      ? TransactionType.TRANSFER
      : requestedType

  const amountOf = (position: WalletPosition, factor: number): string =>
    d.multiply(position.balance, factor.toFixed(6), 6)

  const valueOf = (position: WalletPosition, amount: string): string | null =>
    position.price?.value_usd
      ? d.multiply(amount, position.price.value_usd, 2)
      : null

  const fields = ((): Partial<Transaction> => {
    switch (type) {
      case TransactionType.SWAP: {
        const bought = pick(random, counterAssets)
        const amountOut = amountOf(subject, 0.04 + random() * 0.22)
        const valueOut = valueOf(subject, amountOut)
        const amountIn =
          valueOut !== null && bought.price?.value_usd
            ? (d.divide(valueOut, bought.price.value_usd, 6) ?? '0')
            : '0'

        return {
          asset_out: subject.asset,
          amount_out: amountOut,
          value_out_usd: valueOut,
          asset_in: bought.asset,
          amount_in: amountIn,
          value_in_usd: valueOut,
        }
      }

      case TransactionType.TRANSFER: {
        const amount = amountOf(subject, 0.02 + random() * 0.15)
        const value = valueOf(subject, amount)
        const incoming = random() < 0.45

        return incoming
          ? {
              asset_in: subject.asset,
              amount_in: amount,
              value_in_usd: value,
              from_address: transaction.to_address,
              to_address: transaction.from_address,
            }
          : {
              asset_out: subject.asset,
              amount_out: amount,
              value_out_usd: value,
            }
      }

      case TransactionType.STAKE: {
        const amount = amountOf(subject, 0.05 + random() * 0.1)
        return {
          asset_out: subject.asset,
          amount_out: amount,
          value_out_usd: valueOf(subject, amount),
        }
      }

      case TransactionType.UNSTAKE: {
        const amount = amountOf(subject, 0.05 + random() * 0.1)
        return {
          asset_in: subject.asset,
          amount_in: amount,
          value_in_usd: valueOf(subject, amount),
        }
      }

      case TransactionType.CLAIM: {
        const amount = amountOf(subject, 0.002 + random() * 0.01)
        return {
          asset_in: subject.asset,
          amount_in: amount,
          value_in_usd: valueOf(subject, amount),
        }
      }

      // APPROVE, CONTRACT_INTERACTION and UNKNOWN carry no canonical amounts.
      default:
        return {}
    }
  })()

  return { ...transaction, ...fields, type }
}
