import {
  ApiError,
  apiErrorFromResponse,
  normalizeUnknownError,
} from './errors'
import type { StoredTokens, TokenStore } from './tokenStore'
import { ApiErrorCode } from './types'
import {
  isAuthBootstrapping,
  waitForAuthReady,
} from './authGate'

/**
 * The single HTTP boundary of the application.
 *
 * Responsibilities: base URL and `/api/v1` prefixing, bearer auth, single
 * flight refresh-token rotation, JSON encoding, timeouts, cancellation and
 * error normalisation. Nothing above this layer touches `fetch`.
 */

export type QueryValue = string | number | boolean | null | undefined

export interface RequestOptions {
  signal?: AbortSignal
  /** Overrides the client default. `0` disables the timeout. */
  timeoutMs?: number
}

interface HttpRequest extends RequestOptions {
  method: 'GET' | 'POST' | 'PATCH' | 'PUT' | 'DELETE'
  path: string
  query?: Record<string, QueryValue>
  body?: unknown
  /** Set to false for endpoints that must not carry a bearer token. */
  auth?: boolean
}

/** Exchanges a refresh token for a new pair. Injected to avoid a cycle. */
export type RefreshHandler = (refreshToken: string) => Promise<StoredTokens>

export interface HttpClientConfig {
  baseUrl: string
  timeoutMs: number
  tokens: TokenStore
  /** Called once the session is unrecoverable, so the app can reset state. */
  onSessionExpired?: () => void
}

export class HttpClient {
  private readonly config: HttpClientConfig
  private refreshHandler: RefreshHandler | null = null
  private refreshInFlight: Promise<StoredTokens | null> | null = null

  constructor(config: HttpClientConfig) {
    this.config = config
  }

  setRefreshHandler(handler: RefreshHandler | null): void {
    this.refreshHandler = handler
  }

  /**
   * Proactively rotates an expired access token before the first protected call.
   * Uses the same single-flight refresh as reactive 401 handling.
   */
  async ensureValidAccessToken(): Promise<boolean> {
    const tokens = this.config.tokens.get()
    if (!tokens) return false
    if (!isAccessTokenExpired(tokens.expires_at)) return true
    return this.tryRefresh()
  }

  get<T>(
    path: string,
    query?: Record<string, QueryValue>,
    options?: RequestOptions,
  ): Promise<T> {
    return this.request<T>({ method: 'GET', path, query, ...options })
  }

  post<T>(path: string, body?: unknown, options?: RequestOptions): Promise<T> {
    return this.request<T>({ method: 'POST', path, body, ...options })
  }

  patch<T>(path: string, body?: unknown, options?: RequestOptions): Promise<T> {
    return this.request<T>({ method: 'PATCH', path, body, ...options })
  }

  delete<T>(path: string, options?: RequestOptions): Promise<T> {
    return this.request<T>({ method: 'DELETE', path, ...options })
  }

  /** Same as `post`, but without a bearer token (auth bootstrap endpoints). */
  postUnauthenticated<T>(
    path: string,
    body?: unknown,
    options?: RequestOptions,
  ): Promise<T> {
    return this.request<T>({
      method: 'POST',
      path,
      body,
      auth: false,
      ...options,
    })
  }

  async request<T>(request: HttpRequest): Promise<T> {
    const response = await this.send(request, true)
    return this.readJson<T>(response)
  }

  /**
   * Opens a streaming response (SSE over POST). Timeouts do not apply: an AI
   * stream is long-lived and is ended by the caller's `signal`.
   */
  async openStream(
    path: string,
    body: unknown,
    options?: { signal?: AbortSignal },
  ): Promise<Response> {
    return this.send(
      {
        method: 'POST',
        path,
        body,
        timeoutMs: 0,
        signal: options?.signal,
      },
      true,
      { Accept: 'text/event-stream' },
    )
  }

  /* ---------------------------------------------------------------------- */

  private async send(
    request: HttpRequest,
    allowRefresh: boolean,
    extraHeaders?: Record<string, string>,
  ): Promise<Response> {
    const url = this.buildUrl(request.path, request.query)
    const withAuth = request.auth !== false

    if (withAuth && this.config.tokens.get() && !isAuthBootstrapping()) {
      await waitForAuthReady()
    }

    const headers = new Headers({ Accept: 'application/json', ...extraHeaders })

    if (request.body !== undefined) {
      headers.set('Content-Type', 'application/json')
    }
    if (withAuth) {
      const token = this.config.tokens.get()?.access_token
      if (token) headers.set('Authorization', `Bearer ${token}`)
    }

    const timeoutMs = request.timeoutMs ?? this.config.timeoutMs
    const { signal, dispose } = combineSignals(request.signal, timeoutMs)

    let response: Response
    try {
      response = await fetch(url, {
        method: request.method,
        headers,
        body: request.body === undefined ? null : JSON.stringify(request.body),
        signal,
        credentials: 'omit',
      })
    } catch (error) {
      throw normalizeUnknownError(
        request.signal?.aborted ? error : mapTimeout(error, timeoutMs),
      )
    } finally {
      dispose()
    }

    if (response.ok) return response

    if (response.status === 401 && withAuth && allowRefresh) {
      const refreshed = await this.tryRefresh()
      if (refreshed) return this.send(request, false, extraHeaders)
    }

    throw apiErrorFromResponse(response.status, await safeJson(response))
  }

  /** Single-flight refresh: concurrent 401s wait on one rotation. */
  private async tryRefresh(): Promise<boolean> {
    const refreshToken = this.config.tokens.get()?.refresh_token
    if (!refreshToken || !this.refreshHandler) return false

    this.refreshInFlight ??= this.refreshHandler(refreshToken)
      .then((tokens) => {
        this.config.tokens.set(tokens)
        return tokens
      })
      .catch(() => {
        this.config.tokens.clear()
        this.config.onSessionExpired?.()
        return null
      })
      .finally(() => {
        this.refreshInFlight = null
      })

    return (await this.refreshInFlight) !== null
  }

  private buildUrl(path: string, query?: Record<string, QueryValue>): string {
    const url = `${this.config.baseUrl}${path}`
    if (!query) return url

    const search = new URLSearchParams()
    for (const [key, value] of Object.entries(query)) {
      if (value === null || value === undefined || value === '') continue
      search.set(key, String(value))
    }
    const qs = search.toString()
    return qs ? `${url}?${qs}` : url
  }

  private async readJson<T>(response: Response): Promise<T> {
    if (response.status === 204) return undefined as T
    const text = await response.text()
    if (!text) return undefined as T
    try {
      return JSON.parse(text) as T
    } catch (error) {
      throw new ApiError({
        code: ApiErrorCode.MALFORMED_RESPONSE,
        message: 'The API returned a response that could not be parsed.',
        status: response.status,
        cause: error,
      })
    }
  }
}

/** Small skew so refresh happens slightly before the server rejects the token. */
export function isAccessTokenExpired(
  expiresAt: string,
  skewMs = 30_000,
): boolean {
  const expires = Date.parse(expiresAt)
  if (Number.isNaN(expires)) return false
  return Date.now() >= expires - skewMs
}

function mapTimeout(error: unknown, timeoutMs: number): unknown {
  const aborted = error instanceof DOMException && error.name === 'AbortError'
  if (aborted && timeoutMs > 0) {
    return new DOMException('Request timed out', 'TimeoutError')
  }
  return error
}

async function safeJson(response: Response): Promise<unknown> {
  try {
    const text = await response.text()
    return text ? JSON.parse(text) : null
  } catch {
    return null
  }
}

/**
 * Merges the caller's signal with a timeout signal. `AbortSignal.any` is used
 * when available, with a manual controller as a fallback.
 */
function combineSignals(
  signal: AbortSignal | undefined,
  timeoutMs: number,
): { signal: AbortSignal | undefined; dispose: () => void } {
  if (timeoutMs <= 0) return { signal, dispose: () => {} }

  const timeoutSignal = AbortSignal.timeout(timeoutMs)
  if (!signal) return { signal: timeoutSignal, dispose: () => {} }

  if (typeof AbortSignal.any === 'function') {
    return { signal: AbortSignal.any([signal, timeoutSignal]), dispose: () => {} }
  }

  const controller = new AbortController()
  const abort = () => controller.abort()
  signal.addEventListener('abort', abort, { once: true })
  timeoutSignal.addEventListener('abort', abort, { once: true })
  return {
    signal: controller.signal,
    dispose: () => signal.removeEventListener('abort', abort),
  }
}
