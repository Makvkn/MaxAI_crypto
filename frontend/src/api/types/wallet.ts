/**
 * Wallets. The domain is `1 user -> N wallets`; the MVP UI analyses one wallet
 * at a time but the contract is already plural.
 *
 * PROVISIONAL CONTRACT LAYER — see `src/api/types/index.ts`.
 */
import type { ChainId } from './chain'
import type { ApiErrorBody } from './errors'
import type { DataFreshness, SyncStage, SyncStatus, WalletStatus } from './enums'
import type { Timestamp } from './primitives'

/**
 * Backend-reported synchronisation state for a wallet.
 *
 * `stages_completed` lets the UI show real progress: only stages the backend
 * has actually finished are marked done. The frontend never advances stages
 * on its own.
 */
export interface WalletSyncState {
  status: SyncStatus
  /** Stage currently executing, or `null` when not syncing. */
  stage: SyncStage | null
  stages_completed: SyncStage[]
  started_at: Timestamp | null
  completed_at: Timestamp | null
  last_synced_at: Timestamp | null
  data_freshness: DataFreshness | null
  /** Domain-level failure reason for `FAILED` / `PARTIAL`. */
  error: ApiErrorBody | null
}

export interface Wallet {
  id: string
  chain_id: ChainId
  address: string
  label: string | null
  status: WalletStatus
  sync: WalletSyncState
  created_at: Timestamp
  updated_at: Timestamp
}

/**
 * `POST /wallets` only enqueues `InitialWalletSyncJob`, so the wallet it
 * returns has `sync.status = PENDING` and no portfolio yet.
 */
export interface CreateWalletRequest {
  chain_id: ChainId
  address: string
  label?: string | null
}
