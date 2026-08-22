/**
 * Canonical transactions.
 *
 * Types are assigned by the backend `TransactionClassifier`. `UNKNOWN` stays
 * `UNKNOWN` — the frontend never re-infers a type, and neither does the LLM.
 *
 * PROVISIONAL CONTRACT LAYER — see `src/api/types/index.ts`.
 */
import type { Asset } from './asset'
import type { ChainId } from './chain'
import type { TransactionStatus, TransactionType } from './enums'
import type { Decimal, Timestamp } from './primitives'

export interface Transaction {
  id: string
  wallet_id: string
  chain_id: ChainId
  tx_hash: string
  block_number: number | null
  timestamp: Timestamp
  status: TransactionStatus
  type: TransactionType

  from_address: string | null
  to_address: string | null

  /** Asset that entered the wallet. */
  asset_in: Asset | null
  amount_in: Decimal | null
  value_in_usd: Decimal | null

  /** Asset that left the wallet. */
  asset_out: Asset | null
  amount_out: Decimal | null
  value_out_usd: Decimal | null

  fee_asset: Asset | null
  fee_amount: Decimal | null
  fee_value_usd: Decimal | null

  /** Normalised protocol label, e.g. "Uniswap". */
  protocol: string | null
  counterparty: string | null

  /**
   * Canonical explorer link built by the backend, so the frontend needs no
   * per-chain URL knowledge.
   */
  explorer_url: string | null

  created_at: Timestamp
  updated_at: Timestamp
}

export interface TransactionListParams {
  limit?: number
  cursor?: string | null
  /** Server-side filter; the backend decides what a type means. */
  type?: TransactionType
}
