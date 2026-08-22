/**
 * Authentication and entitlements.
 *
 * A guest is a real anonymous account, so upgrading to Google/Email keeps the
 * same `user.id` and all previously collected data.
 *
 * PROVISIONAL CONTRACT LAYER — see `src/api/types/index.ts`.
 */
import type {
  AuthMethod,
  SubscriptionPlan,
  SubscriptionStatus,
  UserKind,
} from './enums'
import type { Timestamp } from './primitives'

/**
 * Plan capabilities. Feature access is read from here rather than scattered
 * `isPro` checks — and the backend remains authoritative regardless.
 */
export interface Entitlements {
  max_wallets: number
  ai_operations_per_day: number
  features: string[]
}

export interface Subscription {
  plan: SubscriptionPlan
  status: SubscriptionStatus
  current_period_end: Timestamp | null
  entitlements: Entitlements
}

export interface User {
  id: string
  kind: UserKind
  email: string | null
  display_name: string | null
  auth_methods: AuthMethod[]
  subscription: Subscription
  created_at: Timestamp
}

/** Token pair plus the resolved user. */
export interface AuthSession {
  access_token: string
  refresh_token: string
  /** Absolute expiry of the access token. */
  expires_at: Timestamp
  user: User
}

export interface EmailCredentials {
  email: string
  password: string
}

export interface GoogleAuthRequest {
  /** Google ID token obtained by the frontend OAuth flow. */
  id_token: string
}

export interface RefreshRequest {
  refresh_token: string
}

/**
 * Upgrading the current anonymous account. The backend keeps `user.id`, so no
 * data migration happens client-side.
 */
export type UpgradeAccountRequest =
  | ({ method: 'EMAIL' } & EmailCredentials)
  | ({ method: 'GOOGLE' } & GoogleAuthRequest)
