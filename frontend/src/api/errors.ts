import {
  ApiErrorCategory,
  ApiErrorCode,
  type ApiErrorBody,
  type ApiErrorCodeValue,
  type RateLimitDetails,
  type ValidationErrorDetails,
} from './types'

/**
 * Normalised API error.
 *
 * Everything the app catches is an `ApiError` with a domain-level `code`.
 * Transport failures (offline, timeout, abort, malformed body) are normalised
 * into the same shape so callers never branch on `instanceof TypeError`.
 *
 * `message` is developer-facing. User-facing copy is resolved from `code` by
 * `@/lib/errors/messages` — raw backend text is never rendered, because it may
 * carry upstream provider details.
 */
export class ApiError extends Error {
  readonly code: ApiErrorCodeValue
  readonly category: ApiErrorCategory
  readonly status: number | null
  readonly details: Record<string, unknown>

  constructor(params: {
    code: ApiErrorCodeValue
    message: string
    status?: number | null
    details?: Record<string, unknown>
    cause?: unknown
  }) {
    super(params.message, { cause: params.cause })
    this.name = 'ApiError'
    this.code = params.code
    this.category = categoryForCode(params.code, params.status ?? null)
    this.status = params.status ?? null
    this.details = params.details ?? {}
  }

  get rateLimit(): RateLimitDetails | null {
    return this.category === ApiErrorCategory.RATE_LIMIT
      ? (this.details as RateLimitDetails)
      : null
  }

  get fieldErrors(): Record<string, string> {
    const details = this.details as ValidationErrorDetails
    return details.fields ?? {}
  }
}

/** Maps a domain code (or HTTP status as fallback) onto a handling category. */
export function categoryForCode(
  code: ApiErrorCodeValue,
  status: number | null,
): ApiErrorCategory {
  switch (code) {
    case ApiErrorCode.VALIDATION_ERROR:
    case ApiErrorCode.INVALID_WALLET_ADDRESS:
    case ApiErrorCode.UNSUPPORTED_CHAIN:
      return ApiErrorCategory.VALIDATION_ERROR

    case ApiErrorCode.AUTHENTICATION_ERROR:
    case ApiErrorCode.SESSION_EXPIRED:
    case ApiErrorCode.INVALID_CREDENTIALS:
    case ApiErrorCode.EMAIL_ALREADY_REGISTERED:
      return ApiErrorCategory.AUTHENTICATION_ERROR

    case ApiErrorCode.NOT_FOUND:
    case ApiErrorCode.WALLET_NOT_FOUND:
    case ApiErrorCode.TRANSACTION_NOT_FOUND:
    case ApiErrorCode.CONVERSATION_NOT_FOUND:
      return ApiErrorCategory.NOT_FOUND

    case ApiErrorCode.PROVIDER_ERROR:
      return ApiErrorCategory.PROVIDER_ERROR

    case ApiErrorCode.DATA_UNAVAILABLE:
    case ApiErrorCode.PORTFOLIO_DATA_UNAVAILABLE:
    case ApiErrorCode.PORTFOLIO_DATA_TEMPORARILY_UNAVAILABLE:
    case ApiErrorCode.PERFORMANCE_DATA_UNAVAILABLE:
    case ApiErrorCode.PRICE_DATA_UNAVAILABLE:
    case ApiErrorCode.WALLET_SYNC_FAILED:
    case ApiErrorCode.WALLET_NOT_READY:
    case ApiErrorCode.AI_UNAVAILABLE:
      return ApiErrorCategory.DATA_UNAVAILABLE

    case ApiErrorCode.RATE_LIMIT:
    case ApiErrorCode.AI_DAILY_LIMIT_REACHED:
    case ApiErrorCode.WALLET_LIMIT_REACHED:
      return ApiErrorCategory.RATE_LIMIT

    case ApiErrorCode.NETWORK_ERROR:
    case ApiErrorCode.TIMEOUT:
    case ApiErrorCode.CANCELLED:
    case ApiErrorCode.MALFORMED_RESPONSE:
      return ApiErrorCategory.TRANSPORT_ERROR

    case ApiErrorCode.INTERNAL_ERROR:
      return ApiErrorCategory.INTERNAL_ERROR

    default:
      return categoryForStatus(status)
  }
}

function categoryForStatus(status: number | null): ApiErrorCategory {
  if (status === null) return ApiErrorCategory.INTERNAL_ERROR
  if (status === 401 || status === 403)
    return ApiErrorCategory.AUTHENTICATION_ERROR
  if (status === 404) return ApiErrorCategory.NOT_FOUND
  if (status === 409 || status === 422)
    return ApiErrorCategory.VALIDATION_ERROR
  if (status === 429) return ApiErrorCategory.RATE_LIMIT
  if (status === 503) return ApiErrorCategory.DATA_UNAVAILABLE
  if (status >= 500) return ApiErrorCategory.INTERNAL_ERROR
  if (status >= 400) return ApiErrorCategory.VALIDATION_ERROR
  return ApiErrorCategory.INTERNAL_ERROR
}

/** Reads the `{ error: { code, message, details } }` envelope defensively. */
export function parseErrorEnvelope(payload: unknown): ApiErrorBody | null {
  if (typeof payload !== 'object' || payload === null) return null
  const envelope = payload as { error?: unknown }
  if (typeof envelope.error !== 'object' || envelope.error === null) return null

  const body = envelope.error as Partial<ApiErrorBody>
  if (typeof body.code !== 'string') return null

  return {
    code: body.code,
    message: typeof body.message === 'string' ? body.message : body.code,
    details:
      typeof body.details === 'object' && body.details !== null
        ? (body.details as Record<string, unknown>)
        : {},
  }
}

/** Builds an `ApiError` from an HTTP response body and status. */
export function apiErrorFromResponse(
  status: number,
  payload: unknown,
): ApiError {
  const body = parseErrorEnvelope(payload)
  if (body) {
    return new ApiError({
      code: body.code,
      message: body.message,
      status,
      details: body.details,
    })
  }
  return new ApiError({
    code: fallbackCodeForStatus(status),
    message: `Unexpected API response (HTTP ${status}).`,
    status,
  })
}

function fallbackCodeForStatus(status: number): ApiErrorCodeValue {
  if (status === 401) return ApiErrorCode.AUTHENTICATION_ERROR
  if (status === 404) return ApiErrorCode.NOT_FOUND
  if (status === 429) return ApiErrorCode.RATE_LIMIT
  if (status === 503) return ApiErrorCode.DATA_UNAVAILABLE
  if (status >= 500) return ApiErrorCode.INTERNAL_ERROR
  return ApiErrorCode.VALIDATION_ERROR
}

/** Converts anything thrown by the transport layer into an `ApiError`. */
export function normalizeUnknownError(error: unknown): ApiError {
  if (error instanceof ApiError) return error

  if (error instanceof DOMException && error.name === 'AbortError') {
    return new ApiError({
      code: ApiErrorCode.CANCELLED,
      message: 'Request cancelled.',
      cause: error,
    })
  }

  if (error instanceof DOMException && error.name === 'TimeoutError') {
    return new ApiError({
      code: ApiErrorCode.TIMEOUT,
      message: 'Request timed out.',
      cause: error,
    })
  }

  if (error instanceof SyntaxError) {
    return new ApiError({
      code: ApiErrorCode.MALFORMED_RESPONSE,
      message: 'Could not read the API response.',
      cause: error,
    })
  }

  return new ApiError({
    code: ApiErrorCode.NETWORK_ERROR,
    message: 'Could not reach the MaxAI API.',
    cause: error,
  })
}

/** Turns an `ApiError` back into the wire shape (used by the AI stream). */
export function toErrorBody(error: ApiError): ApiErrorBody {
  return { code: error.code, message: error.message, details: error.details }
}

export function isApiError(error: unknown): error is ApiError {
  return error instanceof ApiError
}

export function isCancelledError(error: unknown): boolean {
  return isApiError(error) && error.code === ApiErrorCode.CANCELLED
}

export function isAuthError(error: unknown): boolean {
  return (
    isApiError(error) &&
    error.category === ApiErrorCategory.AUTHENTICATION_ERROR
  )
}

export function isRateLimitError(error: unknown): boolean {
  return isApiError(error) && error.category === ApiErrorCategory.RATE_LIMIT
}

export function isAiLimitError(error: unknown): boolean {
  return isApiError(error) && error.code === ApiErrorCode.AI_DAILY_LIMIT_REACHED
}

export function isDataUnavailableError(error: unknown): boolean {
  return (
    isApiError(error) && error.category === ApiErrorCategory.DATA_UNAVAILABLE
  )
}

export function isNotFoundError(error: unknown): boolean {
  return isApiError(error) && error.category === ApiErrorCategory.NOT_FOUND
}

/** Retry policy shared by TanStack Query and the API client. */
export function isRetryableError(error: unknown): boolean {
  if (!isApiError(error)) return false
  switch (error.category) {
    case ApiErrorCategory.TRANSPORT_ERROR:
      return error.code !== ApiErrorCode.CANCELLED
    case ApiErrorCategory.INTERNAL_ERROR:
    case ApiErrorCategory.PROVIDER_ERROR:
      return true
    default:
      return false
  }
}
