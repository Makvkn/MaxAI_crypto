/**
 * MaxAI Crypto — API contract types.
 *
 * ---------------------------------------------------------------------------
 * PROVISIONAL CONTRACT LAYER
 * ---------------------------------------------------------------------------
 * OpenAPI is the source of truth for the MaxAI REST API. That specification
 * does not exist yet, so this directory hand-maintains the same contract in
 * one isolated place.
 *
 * When the OpenAPI document lands, the intended migration is:
 *
 *   openapi.yaml -> generated types -> re-export from this directory
 *
 * Everything else in the app imports from `@/api/types` only. No DTO shape is
 * ever declared inside a feature or a component, so replacing this layer with
 * generated code is a change to these files alone.
 *
 * Rules enforced by these types:
 *   - Financial values are `Decimal` (strings). The frontend formats them; it
 *     never computes with them.
 *   - `null` means "unknown", never "zero".
 *   - Provider identities (Zerion / Tatum / CoinGecko / OpenAI) do not appear.
 *   - Every state machine is an explicit union, never a loose string.
 */

export * from './ai'
export * from './asset'
export * from './auth'
export * from './chain'
export * from './enums'
export * from './errors'
export * from './performance'
export * from './portfolio'
export * from './primitives'
export * from './scenario'
export * from './transaction'
export * from './wallet'
