/**
 * The API error contract.
 *
 * Every backend error is a domain-level code. Provider names (Zerion, Tatum,
 * CoinGecko, OpenAI), HTTP details of upstream calls and stack traces stay
 * behind the API boundary and never reach this layer.
 *
 * PROVISIONAL CONTRACT LAYER — see `src/api/types/index.ts`.
 */

/** Broad category, used for handling strategy (retry, re-auth, ...). */
export const ApiErrorCategory = {
  VALIDATION_ERROR: 'VALIDATION_ERROR',
  AUTHENTICATION_ERROR: 'AUTHENTICATION_ERROR',
  NOT_FOUND: 'NOT_FOUND',
  PROVIDER_ERROR: 'PROVIDER_ERROR',
  DATA_UNAVAILABLE: 'DATA_UNAVAILABLE',
  RATE_LIMIT: 'RATE_LIMIT',
  INTERNAL_ERROR: 'INTERNAL_ERROR',
  /** Client-side only: the request never produced a backend response. */
  TRANSPORT_ERROR: 'TRANSPORT_ERROR',
} as const
export type ApiErrorCategory =
  (typeof ApiErrorCategory)[keyof typeof ApiErrorCategory]

/**
 * Known domain error codes. The union stays open (`(string & {})`) so an
 * unrecognised backend code degrades gracefully instead of breaking the build.
 */
export const ApiErrorCode = {
  VALIDATION_ERROR: 'VALIDATION_ERROR',
  INVALID_WALLET_ADDRESS: 'INVALID_WALLET_ADDRESS',
  UNSUPPORTED_CHAIN: 'UNSUPPORTED_CHAIN',

  AUTHENTICATION_ERROR: 'AUTHENTICATION_ERROR',
  SESSION_EXPIRED: 'SESSION_EXPIRED',
  EMAIL_ALREADY_REGISTERED: 'EMAIL_ALREADY_REGISTERED',
  INVALID_CREDENTIALS: 'INVALID_CREDENTIALS',

  NOT_FOUND: 'NOT_FOUND',
  WALLET_NOT_FOUND: 'WALLET_NOT_FOUND',
  TRANSACTION_NOT_FOUND: 'TRANSACTION_NOT_FOUND',
  CONVERSATION_NOT_FOUND: 'CONVERSATION_NOT_FOUND',

  PROVIDER_ERROR: 'PROVIDER_ERROR',

  DATA_UNAVAILABLE: 'DATA_UNAVAILABLE',
  PORTFOLIO_DATA_UNAVAILABLE: 'PORTFOLIO_DATA_UNAVAILABLE',
  PORTFOLIO_DATA_TEMPORARILY_UNAVAILABLE:
    'PORTFOLIO_DATA_TEMPORARILY_UNAVAILABLE',
  PERFORMANCE_DATA_UNAVAILABLE: 'PERFORMANCE_DATA_UNAVAILABLE',
  PRICE_DATA_UNAVAILABLE: 'PRICE_DATA_UNAVAILABLE',
  WALLET_SYNC_FAILED: 'WALLET_SYNC_FAILED',
  WALLET_NOT_READY: 'WALLET_NOT_READY',

  RATE_LIMIT: 'RATE_LIMIT',
  AI_DAILY_LIMIT_REACHED: 'AI_DAILY_LIMIT_REACHED',
  WALLET_LIMIT_REACHED: 'WALLET_LIMIT_REACHED',

  AI_UNAVAILABLE: 'AI_UNAVAILABLE',
  AI_STREAM_FAILED: 'AI_STREAM_FAILED',

  INTERNAL_ERROR: 'INTERNAL_ERROR',

  /** Client-side only. */
  NETWORK_ERROR: 'NETWORK_ERROR',
  TIMEOUT: 'TIMEOUT',
  CANCELLED: 'CANCELLED',
  MALFORMED_RESPONSE: 'MALFORMED_RESPONSE',
} as const
export type KnownApiErrorCode =
  (typeof ApiErrorCode)[keyof typeof ApiErrorCode]

/** Open union: unknown backend codes are carried through, not rejected. */
export type ApiErrorCodeValue = KnownApiErrorCode | (string & {})

/** The `error` object inside the API error envelope. */
export interface ApiErrorBody {
  code: ApiErrorCodeValue
  message: string
  details?: Record<string, unknown>
}

/** The only error shape the backend returns. */
export interface ApiErrorEnvelope {
  error: ApiErrorBody
}

/** Extra fields the backend may attach to a `RATE_LIMIT` error. */
export interface RateLimitDetails {
  limit?: number
  used?: number
  remaining?: number
  resets_at?: string
}

/** Extra fields the backend may attach to a `VALIDATION_ERROR`. */
export interface ValidationErrorDetails {
  fields?: Record<string, string>
}
