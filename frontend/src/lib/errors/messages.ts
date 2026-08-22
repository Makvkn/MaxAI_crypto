import { ApiError } from '@/api/errors'
import { ApiErrorCategory, ApiErrorCode, type ApiErrorBody } from '@/api/types'

/**
 * User-facing error copy, resolved from domain error codes.
 *
 * The backend `message` field is never rendered: it is developer-facing and may
 * describe upstream provider behaviour. The user sees product language only —
 * "Portfolio data is temporarily unavailable", not "Tatum 429".
 */
export interface ErrorCopy {
  title: string
  description: string
  /** Whether offering a retry makes sense for this class of failure. */
  retryable: boolean
}

const BY_CODE: Partial<Record<string, ErrorCopy>> = {
  [ApiErrorCode.INVALID_WALLET_ADDRESS]: {
    title: 'Enter a valid wallet address',
    description:
      'That address does not look right for the selected network. Check it and try again.',
    retryable: false,
  },
  [ApiErrorCode.UNSUPPORTED_CHAIN]: {
    title: 'Network not supported yet',
    description: 'This network is not available for analysis at the moment.',
    retryable: false,
  },
  [ApiErrorCode.VALIDATION_ERROR]: {
    title: 'Check the details you entered',
    description: 'Something in the request was not accepted. Review and retry.',
    retryable: false,
  },

  [ApiErrorCode.AUTHENTICATION_ERROR]: {
    title: 'Sign in to continue',
    description: 'Your session is no longer valid.',
    retryable: false,
  },
  [ApiErrorCode.SESSION_EXPIRED]: {
    title: 'Session expired',
    description: 'Sign in again to pick up where you left off.',
    retryable: false,
  },
  [ApiErrorCode.INVALID_CREDENTIALS]: {
    title: 'Those credentials did not work',
    description: 'Check your email and password and try again.',
    retryable: false,
  },
  [ApiErrorCode.EMAIL_ALREADY_REGISTERED]: {
    title: 'This email already has an account',
    description: 'Sign in instead, and your existing data will be there.',
    retryable: false,
  },

  [ApiErrorCode.WALLET_NOT_FOUND]: {
    title: 'Wallet not found',
    description: 'This wallet is no longer available on your account.',
    retryable: false,
  },
  [ApiErrorCode.TRANSACTION_NOT_FOUND]: {
    title: 'Transaction not found',
    description: 'This transaction is not part of the analysed history.',
    retryable: false,
  },
  [ApiErrorCode.CONVERSATION_NOT_FOUND]: {
    title: 'Conversation not found',
    description: 'This conversation is no longer available.',
    retryable: false,
  },

  [ApiErrorCode.WALLET_NOT_READY]: {
    title: 'Still analysing this wallet',
    description:
      'The first synchronisation has not finished, so there is nothing to value yet.',
    retryable: true,
  },
  [ApiErrorCode.WALLET_SYNC_FAILED]: {
    title: 'Unable to analyse this wallet',
    description:
      'We could not read this wallet from its network. You can retry the analysis.',
    retryable: true,
  },
  [ApiErrorCode.PORTFOLIO_DATA_UNAVAILABLE]: {
    title: 'Portfolio data is unavailable',
    description:
      'We will not show a value we cannot verify. Your balances are unaffected.',
    retryable: true,
  },
  [ApiErrorCode.PORTFOLIO_DATA_TEMPORARILY_UNAVAILABLE]: {
    title: 'Portfolio data is temporarily unavailable',
    description:
      'Valuation data could not be retrieved just now. Nothing has changed in your wallet.',
    retryable: true,
  },
  [ApiErrorCode.PERFORMANCE_DATA_UNAVAILABLE]: {
    title: 'Performance is not available yet',
    description:
      'Performance is measured from stored portfolio snapshots, and there is not enough history for this period.',
    retryable: false,
  },
  [ApiErrorCode.PRICE_DATA_UNAVAILABLE]: {
    title: 'Price unavailable',
    description:
      'There is no reliable market price for this asset, so it cannot be valued.',
    retryable: true,
  },

  [ApiErrorCode.AI_DAILY_LIMIT_REACHED]: {
    title: "You've reached your daily AI limit",
    description:
      'Your free plan includes 10 AI operations per day. The dashboard stays fully available.',
    retryable: false,
  },
  [ApiErrorCode.WALLET_LIMIT_REACHED]: {
    title: 'Wallet limit reached',
    description: 'Your current plan analyses one wallet at a time.',
    retryable: false,
  },
  [ApiErrorCode.RATE_LIMIT]: {
    title: 'Too many requests',
    description: 'Give it a moment and try again.',
    retryable: true,
  },

  [ApiErrorCode.AI_UNAVAILABLE]: {
    title: 'AI is unavailable right now',
    description: 'The analysis service could not be reached. Try again shortly.',
    retryable: true,
  },
  [ApiErrorCode.AI_STREAM_FAILED]: {
    title: 'The answer was interrupted',
    description: 'The response stopped before it finished. You can ask again.',
    retryable: true,
  },

  [ApiErrorCode.NETWORK_ERROR]: {
    title: 'No connection',
    description: 'Check your internet connection and try again.',
    retryable: true,
  },
  [ApiErrorCode.TIMEOUT]: {
    title: 'That took too long',
    description: 'The request timed out before it completed.',
    retryable: true,
  },
}

const BY_CATEGORY: Record<ApiErrorCategory, ErrorCopy> = {
  [ApiErrorCategory.VALIDATION_ERROR]: {
    title: 'Check the details you entered',
    description: 'Something in the request was not accepted.',
    retryable: false,
  },
  [ApiErrorCategory.AUTHENTICATION_ERROR]: {
    title: 'Sign in to continue',
    description: 'Your session is no longer valid.',
    retryable: false,
  },
  [ApiErrorCategory.NOT_FOUND]: {
    title: 'Not found',
    description: 'This item is no longer available.',
    retryable: false,
  },
  [ApiErrorCategory.PROVIDER_ERROR]: {
    title: 'Data source is unavailable',
    description: 'We could not retrieve this data. Please try again shortly.',
    retryable: true,
  },
  [ApiErrorCategory.DATA_UNAVAILABLE]: {
    title: 'Data is temporarily unavailable',
    description: 'We will not display values we cannot verify.',
    retryable: true,
  },
  [ApiErrorCategory.RATE_LIMIT]: {
    title: 'Limit reached',
    description: 'Give it a moment and try again.',
    retryable: true,
  },
  [ApiErrorCategory.INTERNAL_ERROR]: {
    title: 'Something went wrong',
    description: 'Please try again. If it keeps happening, come back later.',
    retryable: true,
  },
  [ApiErrorCategory.TRANSPORT_ERROR]: {
    title: 'Connection problem',
    description: 'The request did not reach us. Check your connection.',
    retryable: true,
  },
}

const FALLBACK: ErrorCopy = {
  title: 'Something went wrong',
  description: 'Please try again.',
  retryable: true,
}

/** Resolves copy for anything thrown by the API layer. */
export function errorCopy(error: unknown): ErrorCopy {
  if (error instanceof ApiError) {
    return BY_CODE[error.code] ?? BY_CATEGORY[error.category] ?? FALLBACK
  }
  return FALLBACK
}

/** Resolves copy for an error body carried inside a payload (AI messages). */
export function errorBodyCopy(body: ApiErrorBody | null): ErrorCopy {
  if (!body) return FALLBACK
  return (
    BY_CODE[body.code] ??
    BY_CATEGORY[new ApiError({ code: body.code, message: body.message }).category] ??
    FALLBACK
  )
}
