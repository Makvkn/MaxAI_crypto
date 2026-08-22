import { useState, type FormEvent } from 'react'
import type { ChainId } from '@/api/types'
import { Button } from '@/components/ui/Button'
import { ArrowLeft, ArrowRight, Shield } from '@/components/ui/Icon'
import { Input } from '@/components/ui/Input'
import { ErrorState } from '@/components/feedback/States'
import { chainPresentation } from '@/app/config/chains'
import { ChainMonogram } from '@/features/wallets/components/ChainMonogram'
import {
  looksLikeSecret,
  validateWalletAddress,
} from '@/lib/validation/walletAddress'

/**
 * Address entry.
 *
 * Only a public address is ever requested. Anything resembling a secret is
 * refused locally and never sent anywhere.
 */
export function AddressStep({
  chainId,
  address,
  onAddressChange,
  onBack,
  onSubmit,
  submitting,
  error,
}: {
  chainId: ChainId
  address: string
  onAddressChange: (value: string) => void
  onBack: () => void
  onSubmit: (address: string) => void
  submitting: boolean
  error: unknown
}) {
  const chain = chainPresentation(chainId)
  const [touched, setTouched] = useState(false)

  const secret = looksLikeSecret(address)
  const validation = validateWalletAddress(chainId, address)
  const localError = secret
    ? 'That looks like a private key or seed phrase. Enter a public wallet address only — never a secret.'
    : touched
      ? validation.message
      : null

  const handleSubmit = (event: FormEvent) => {
    event.preventDefault()
    setTouched(true)
    if (secret || !validation.valid) return
    onSubmit(address.trim())
  }

  return (
    <div>
      <button
        type="button"
        onClick={onBack}
        className="mb-6 inline-flex items-center gap-1.5 rounded-md text-[13px] text-fg-subtle transition-colors hover:text-fg"
      >
        <ArrowLeft className="size-3.5" />
        Change network
      </button>

      <div className="flex items-center gap-3">
        <ChainMonogram chainId={chainId} size="lg" />
        <div>
          <h1 className="text-2xl font-medium tracking-tight text-fg">
            {chain.name} wallet
          </h1>
          <p className="text-sm text-fg-muted">{chain.summary}</p>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="mt-8" noValidate>
        <Input
          label="Wallet address"
          value={address}
          onChange={(event) => onAddressChange(event.target.value)}
          onBlur={() => setTouched(true)}
          placeholder={chain.addressPlaceholder}
          monospace
          autoComplete="off"
          autoCorrect="off"
          spellCheck={false}
          autoFocus
          error={localError}
          hint={chain.addressHint}
        />

        <div className="mt-5 flex items-start gap-2.5 rounded-lg border border-line bg-surface px-4 py-3">
          <Shield className="mt-0.5 size-4 shrink-0 text-positive" />
          <p className="text-[13px] leading-relaxed text-fg-subtle">
            Read-only analysis. MaxAI Crypto never asks for a private key, seed
            phrase or signature, and cannot move funds.
          </p>
        </div>

        {error ? (
          <ErrorState
            error={error}
            compact
            className="mt-5 rounded-lg border border-line bg-surface"
          />
        ) : null}

        <Button
          type="submit"
          size="lg"
          className="mt-6 w-full"
          loading={submitting}
          iconRight={<ArrowRight className="size-4" />}
        >
          Analyze wallet
        </Button>
      </form>
    </div>
  )
}
