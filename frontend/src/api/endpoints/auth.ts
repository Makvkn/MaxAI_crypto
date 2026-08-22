import type { AuthApi } from '../contract'
import type { HttpClient } from '../client'
import type { AuthSession, User } from '../types'

/**
 * `POST /api/v1/auth/...`
 *
 * All authorisation decisions stay on the backend; this module only manages
 * the session lifecycle.
 */
export function createAuthApi(http: HttpClient): AuthApi {
  return {
    createGuestSession: (options) =>
      http.postUnauthenticated<AuthSession>('/auth/guest', {}, options),

    registerWithEmail: (credentials, options) =>
      http.postUnauthenticated<AuthSession>(
        '/auth/email/register',
        credentials,
        options,
      ),

    loginWithEmail: (credentials, options) =>
      http.postUnauthenticated<AuthSession>(
        '/auth/email/login',
        credentials,
        options,
      ),

    loginWithGoogle: (request, options) =>
      http.postUnauthenticated<AuthSession>('/auth/google', request, options),

    // Authenticated: the backend upgrades the account behind the current
    // access token, preserving `user.id` and all existing data.
    upgradeAccount: (request, options) =>
      http.post<AuthSession>('/auth/upgrade', request, options),

    getCurrentUser: (options) =>
      http.get<User>('/auth/session', undefined, options),

    logout: (options) => http.post<void>('/auth/logout', {}, options),
  }
}

/**
 * Refresh is wired directly into the client (not exposed on `AuthApi`) so that
 * token rotation cannot be triggered as a normal feature-level call.
 */
export function installRefreshHandler(http: HttpClient): void {
  http.setRefreshHandler(async (refreshToken) => {
    const session = await http.postUnauthenticated<AuthSession>(
      '/auth/refresh',
      { refresh_token: refreshToken },
      { timeoutMs: 10_000 },
    )
    return {
      access_token: session.access_token,
      refresh_token: session.refresh_token,
      expires_at: session.expires_at,
    }
  })
}
