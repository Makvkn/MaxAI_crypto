import type { RequestOptions } from '../client'
import { http } from '../http'
import type { AuthSession, EmailCredentials, GoogleAuthRequest, UpgradeAccountRequest, User } from '../types'

/**
 * `POST /api/v1/auth/...`
 *
 * All authorisation decisions stay on the backend; this module only manages
 * the session lifecycle.
 */

export const apiCreateGuestSession = (
  options?: RequestOptions,
): Promise<AuthSession> => http.postUnauthenticated<AuthSession>('/auth/guest', {}, options)

export const apiRegisterWithEmail = (
  credentials: EmailCredentials,
  options?: RequestOptions,
): Promise<AuthSession> =>
  http.postUnauthenticated<AuthSession>('/auth/email/register', credentials, options)

export const apiLoginWithEmail = (
  credentials: EmailCredentials,
  options?: RequestOptions,
): Promise<AuthSession> =>
  http.postUnauthenticated<AuthSession>('/auth/email/login', credentials, options)

export const apiLoginWithGoogle = (
  request: GoogleAuthRequest,
  options?: RequestOptions,
): Promise<AuthSession> =>
  http.postUnauthenticated<AuthSession>('/auth/google', request, options)

export const apiUpgradeAccount = (
  request: UpgradeAccountRequest,
  options?: RequestOptions,
): Promise<AuthSession> => http.post<AuthSession>('/auth/upgrade', request, options)

export const apiGetCurrentUser = (options?: RequestOptions): Promise<User> =>
  http.get<User>('/auth/session', undefined, options)

export const apiInitializeSession = async (options?: RequestOptions): Promise<User> => {
  await http.ensureValidAccessToken()
  return http.get<User>('/auth/session', undefined, options)
}

export const apiLogout = (options?: RequestOptions): Promise<void> =>
  http.post<void>('/auth/logout', {}, options)

/**
 * Refresh is wired directly into the client so that token rotation cannot be
 * triggered as a normal feature-level call.
 */
export function installRefreshHandler(): void {
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
