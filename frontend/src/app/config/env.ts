/**
 * Validated environment configuration.
 *
 * The frontend holds no secrets: no OpenAI, Zerion, Tatum or CoinGecko keys
 * exist here. It only needs to know where the MaxAI backend lives and whether
 * to run against the mock adapter.
 *
 * Validation is deliberately hand-written: this runs on every page load before
 * anything else, and a misconfigured client must fail loudly at boot rather
 * than silently talk to the wrong backend.
 */

const problems: string[] = []

function readString(key: string, fallback: string): string {
  const value = import.meta.env[key]
  if (value === undefined || value === '') return fallback
  if (typeof value !== 'string') {
    problems.push(`${key} must be a string`)
    return fallback
  }
  return value
}

function readEnum<T extends string>(
  key: string,
  allowed: readonly T[],
  fallback: T,
): T {
  const value = readString(key, fallback)
  if (!allowed.includes(value as T)) {
    problems.push(`${key} must be one of: ${allowed.join(', ')}`)
    return fallback
  }
  return value as T
}

function readPositiveInt(key: string, fallback: number): number {
  const raw = readString(key, String(fallback))
  const value = Number(raw)
  if (!Number.isSafeInteger(value) || value <= 0) {
    problems.push(`${key} must be a positive integer`)
    return fallback
  }
  return value
}

function readBoolean(key: string, fallback: boolean): boolean {
  const raw = readString(key, String(fallback))
  if (raw === 'true' || raw === '1') return true
  if (raw === 'false' || raw === '0') return false
  problems.push(`${key} must be true or false`)
  return fallback
}

const apiMode =
  import.meta.env.MODE === 'test'
    ? 'mock'
    : readEnum('VITE_API_MODE', ['mock', 'real'] as const, 'mock')
/** Empty means same-origin, which the dev proxy relies on. */
const apiBaseUrl = readString('VITE_API_BASE_URL', '')
const apiTimeoutMs = readPositiveInt('VITE_API_TIMEOUT_MS', 15_000)
const analyticsEnabled = readBoolean('VITE_ANALYTICS_ENABLED', false)
const analyticsDebug = readBoolean('VITE_ANALYTICS_DEBUG', false)

if (problems.length > 0) {
  throw new Error(
    `Invalid frontend environment configuration:\n- ${problems.join('\n- ')}`,
  )
}

/** API version prefix. Every request goes through `/api/v1`. */
export const API_VERSION_PATH = '/api/v1'

export const env = {
  apiMode,
  apiBaseUrl: apiBaseUrl.replace(/\/$/, ''),
  apiTimeoutMs,
  analyticsEnabled,
  analyticsDebug,
  isDev: import.meta.env.DEV,
} as const

export type Env = typeof env

/** Absolute base for all REST calls, e.g. `https://api.maxai.crypto/api/v1`. */
export const apiRoot = `${env.apiBaseUrl}${API_VERSION_PATH}`
