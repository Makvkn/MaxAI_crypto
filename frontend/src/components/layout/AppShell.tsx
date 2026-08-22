import type { ReactNode } from 'react'
import { AccountMenu } from '@/features/auth/components/AccountMenu'
import { WalletSwitcher } from '@/features/wallets/components/WalletSwitcher'
import { Container } from './Container'
import { Logo } from './Logo'

/**
 * Dashboard chrome.
 *
 * Deliberately thin: the header carries identity, wallet selection and account
 * state, and everything else is the analysis itself.
 */
export function AppShell({
  walletId,
  children,
}: {
  walletId?: string
  children: ReactNode
}) {
  return (
    <div className="min-h-dvh bg-base">
      <header className="sticky top-0 z-40 border-b border-line/70 bg-base/85 backdrop-blur-xl">
        <Container size="wide" className="flex h-16 items-center gap-4">
          <Logo />

          <div className="ml-auto flex items-center gap-2.5">
            {walletId ? <WalletSwitcher activeWalletId={walletId} /> : null}
            <AccountMenu />
          </div>
        </Container>
      </header>

      <main>{children}</main>
    </div>
  )
}
