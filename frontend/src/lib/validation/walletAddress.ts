import type { ChainId } from '@/api/types'
import { chainPresentation } from '@/app/config/chains'

/**
 * Inline address validation.
 *
 * This is a UX courtesy, not authority: it catches typos before a round trip.
 * The backend remains the validator of record, and its
 * `INVALID_WALLET_ADDRESS` response is always surfaced.
 */
export interface AddressValidationResult {
  valid: boolean
  /** Reason to show under the input, when invalid. */
  message: string | null
}

export function validateWalletAddress(
  chainId: ChainId | null,
  address: string,
): AddressValidationResult {
  const value = address.trim()

  if (value === '') {
    return { valid: false, message: 'Enter a wallet address.' }
  }
  if (!chainId) {
    return { valid: false, message: 'Select a network first.' }
  }

  const pattern = chainPresentation(chainId).addressPattern
  if (pattern && !pattern.test(value)) {
    return {
      valid: false,
      message: `Enter a valid wallet address. ${chainPresentation(chainId).addressHint}`,
    }
  }

  return { valid: true, message: null }
}

/** Never accept anything that looks like a secret. */
export function looksLikeSecret(value: string): boolean {
  const words = value.trim().split(/\s+/)
  if (words.length >= 12) return true
  return /^(0x)?[a-fA-F0-9]{64}$/.test(value.trim())
}
