import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import type { ChainId } from '@/api/types'
import { Container } from '@/components/layout/Container'
import { Logo } from '@/components/layout/Logo'
import { ChainSelectStep } from '@/features/onboarding/components/ChainSelectStep'
import { AddressStep } from '@/features/onboarding/components/AddressStep'
import { useCreateWallet } from '@/features/wallets/hooks/useWallets'
import { analytics } from '@/lib/analytics/analytics'
import { useOnboardingStore } from '@/stores/onboardingStore'
import { usePreferencesStore } from '@/stores/preferencesStore'

/**
 * Onboarding: network -> address -> analyse.
 *
 * Creating the wallet only enqueues the backend job, so this page hands off to
 * the wallet route, which renders the real synchronisation state.
 */
export function AnalyzePage() {
  const navigate = useNavigate()
  const { chainId, address, setChain, setAddress, reset } = useOnboardingStore()
  const selectWallet = usePreferencesStore((state) => state.selectWallet)
  const markOnboardingComplete = usePreferencesStore(
    (state) => state.markOnboardingComplete,
  )
  const createWallet = useCreateWallet()

  const [step, setStep] = useState<'chain' | 'address'>(
    chainId ? 'address' : 'chain',
  )

  const handleSelectChain = (nextChainId: ChainId) => {
    setChain(nextChainId)
    setStep('address')
  }

  const handleSubmit = async (submittedAddress: string) => {
    if (!chainId) return

    analytics.track('wallet_entered', { chain_id: chainId })

    try {
      const wallet = await createWallet.mutateAsync({
        chain_id: chainId,
        address: submittedAddress,
      })
      selectWallet(wallet.id)
      markOnboardingComplete()
      reset()
      navigate(`/wallets/${wallet.id}`, { replace: true })
    } catch {
      // Rendered from `createWallet.error`; domain copy is resolved there.
    }
  }

  return (
    <div className="min-h-dvh bg-base">
      <header className="border-b border-line/70">
        <Container className="flex h-16 items-center">
          <Logo />
        </Container>
      </header>

      <main>
        <Container size="narrow" className="py-14 lg:py-20">
          <p className="mb-8 text-[11px] font-medium tracking-[0.12em] text-fg-subtle uppercase">
            Step {step === 'chain' ? 1 : 2} of 2
          </p>

          {step === 'chain' ? (
            <ChainSelectStep selected={chainId} onSelect={handleSelectChain} />
          ) : chainId ? (
            <AddressStep
              chainId={chainId}
              address={address}
              onAddressChange={setAddress}
              onBack={() => setStep('chain')}
              onSubmit={handleSubmit}
              submitting={createWallet.isPending}
              error={createWallet.error}
            />
          ) : null}
        </Container>
      </main>
    </div>
  )
}
