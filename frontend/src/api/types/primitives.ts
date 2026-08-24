/**
 * Shared primitive aliases for the MaxAI API contract.
 *
 * PROVISIONAL CONTRACT LAYER — see `src/api/types/index.ts`.
 */

/**
 * A fixed-point decimal serialised as a string (PostgreSQL `NUMERIC`, Go
 * `decimal.Decimal`).
 *
 * The backend owns every financial calculation. The frontend must never do
 * arithmetic on these values; it may only parse them for presentation
 * formatting or chart plotting.
 */
export type Decimal = string

/** RFC 3339 / ISO 8601 timestamp, always UTC. */
export type Timestamp = string

/** Calendar date, `YYYY-MM-DD`, UTC. */
export type DateOnly = string

/** ISO 4217 currency code. The MVP values everything in USD. */
export type CurrencyCode = 'USD'

/** Opaque cursor. Never parsed or constructed by the frontend. */
export type Cursor = string

/** Request parameters for every cursor-paginated collection. */
export interface CursorParams {
  limit?: number
  cursor?: Cursor | null
}

/** Path parameter objects for resource URLs. */
export interface WalletIdPath {
  walletId: string
}

export interface ConversationIdPath {
  conversationId: string
}

export interface TransactionIdPath {
  transactionId: string
}

/**
 * The single pagination envelope used by the API. The backend uses
 * cursor-based pagination exclusively — there is no page/offset concept.
 */
export interface CursorPage<T> {
  items: T[]
  next_cursor: Cursor | null
  has_more: boolean
}
