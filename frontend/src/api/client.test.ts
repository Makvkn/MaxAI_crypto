import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { HttpClient, isAccessTokenExpired } from './client'
import { resetAuthGate, syncAuthGate } from './authGate'
import { createTokenStore, type StoredTokens } from './tokenStore'

function futureExpiry(): string {
  return new Date(Date.now() + 60 * 60_000).toISOString()
}

function pastExpiry(): string {
  return new Date(Date.now() - 60_000).toISOString()
}

function tokens(overrides?: Partial<StoredTokens>): StoredTokens {
  return {
    access_token: 'access-token',
    refresh_token: 'refresh-token',
    expires_at: futureExpiry(),
    ...overrides,
  }
}

describe('isAccessTokenExpired', () => {
  it('returns false for a token that expires in the future', () => {
    expect(isAccessTokenExpired(futureExpiry())).toBe(false)
  })

  it('returns true for a token that already expired', () => {
    expect(isAccessTokenExpired(pastExpiry())).toBe(true)
  })
})

describe('HttpClient.ensureValidAccessToken', () => {
  beforeEach(() => {
    resetAuthGate()
    syncAuthGate('authenticated')
  })

  afterEach(() => {
    resetAuthGate()
    syncAuthGate('authenticated')
    vi.restoreAllMocks()
  })

  it('skips refresh when the access token is still valid', async () => {
    const store = createTokenStore('test.tokens')
    store.set(tokens())

    const refresh = vi.fn()
    const client = new HttpClient({
      baseUrl: 'https://api.test',
      timeoutMs: 5_000,
      tokens: store,
    })
    client.setRefreshHandler(refresh)

    await expect(client.ensureValidAccessToken()).resolves.toBe(true)
    expect(refresh).not.toHaveBeenCalled()
  })

  it('refreshes proactively when the access token is expired', async () => {
    const store = createTokenStore('test.tokens.expired')
    store.set(tokens({ expires_at: pastExpiry() }))

    const fetchMock = vi.fn(async (input: RequestInfo) => {
      const url = String(input)
      if (url.endsWith('/auth/refresh')) {
        return new Response(
          JSON.stringify({
            access_token: 'new-access',
            refresh_token: 'new-refresh',
            expires_at: futureExpiry(),
            user: { id: 'user-1' },
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }
      return new Response('unexpected', { status: 500 })
    })
    vi.stubGlobal('fetch', fetchMock)

    const client = new HttpClient({
      baseUrl: 'https://api.test/api/v1',
      timeoutMs: 5_000,
      tokens: store,
    })
    client.setRefreshHandler(async (refreshToken) => {
      const response = await fetch(`https://api.test/api/v1/auth/refresh`, {
        method: 'POST',
        body: JSON.stringify({ refresh_token: refreshToken }),
      })
      const session = (await response.json()) as StoredTokens
      return session
    })

    await expect(client.ensureValidAccessToken()).resolves.toBe(true)
    expect(fetchMock).toHaveBeenCalledTimes(1)
    expect(store.get()?.access_token).toBe('new-access')
  })

  it('deduplicates concurrent proactive refreshes', async () => {
    const store = createTokenStore('test.tokens.concurrent')
    store.set(tokens({ expires_at: pastExpiry() }))

    let refreshCalls = 0
    const fetchMock = vi.fn(async () => {
      refreshCalls += 1
      await new Promise((resolve) => setTimeout(resolve, 20))
      return new Response(
        JSON.stringify({
          access_token: 'new-access',
          refresh_token: 'new-refresh',
          expires_at: futureExpiry(),
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      )
    })
    vi.stubGlobal('fetch', fetchMock)

    const client = new HttpClient({
      baseUrl: 'https://api.test/api/v1',
      timeoutMs: 5_000,
      tokens: store,
    })
    client.setRefreshHandler(async (refreshToken) => {
      const response = await fetch(`https://api.test/api/v1/auth/refresh`, {
        method: 'POST',
        body: JSON.stringify({ refresh_token: refreshToken }),
      })
      return (await response.json()) as StoredTokens
    })

    const [first, second] = await Promise.all([
      client.ensureValidAccessToken(),
      client.ensureValidAccessToken(),
    ])

    expect(first).toBe(true)
    expect(second).toBe(true)
    expect(refreshCalls).toBe(1)
  })

  it('retries an authenticated request after a runtime 401', async () => {
    const store = createTokenStore('test.tokens.runtime401')
    store.set(tokens())

    let sessionCalls = 0
    const fetchMock = vi.fn(async (input: RequestInfo) => {
      const url = String(input)
      if (url.endsWith('/auth/refresh')) {
        return new Response(
          JSON.stringify({
            access_token: 'rotated-access',
            refresh_token: 'rotated-refresh',
            expires_at: futureExpiry(),
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }
      if (url.endsWith('/auth/session')) {
        sessionCalls += 1
        if (sessionCalls === 1) {
          return new Response(
            JSON.stringify({
              code: 'AUTHENTICATION_ERROR',
              message: 'Authentication is required.',
            }),
            { status: 401, headers: { 'Content-Type': 'application/json' } },
          )
        }
        return new Response(JSON.stringify({ id: 'user-1' }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        })
      }
      return new Response('unexpected', { status: 500 })
    })
    vi.stubGlobal('fetch', fetchMock)

    const client = new HttpClient({
      baseUrl: 'https://api.test/api/v1',
      timeoutMs: 5_000,
      tokens: store,
    })
    client.setRefreshHandler(async (refreshToken) => {
      const response = await fetch(`https://api.test/api/v1/auth/refresh`, {
        method: 'POST',
        body: JSON.stringify({ refresh_token: refreshToken }),
      })
      return (await response.json()) as StoredTokens
    })

    await expect(client.get('/auth/session')).resolves.toEqual({ id: 'user-1' })
    expect(sessionCalls).toBe(2)
    expect(
      fetchMock.mock.calls.filter(([url]) =>
        String(url).endsWith('/auth/refresh'),
      ),
    ).toHaveLength(1)
  })
})
