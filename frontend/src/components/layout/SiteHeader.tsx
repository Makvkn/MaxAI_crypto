import { Link } from 'react-router-dom'
import { ButtonLink } from '@/components/ui/Button'
import { ArrowRight } from '@/components/ui/Icon'
import { useSession } from '@/features/auth/sessionContext'
import { usePreferencesStore } from '@/stores/preferencesStore'
import { Container } from './Container'
import { Logo } from './Logo'

/** Marketing header used by the landing page. */
export function SiteHeader() {
  const { isAuthenticated } = useSession()
  const selectedWalletId = usePreferencesStore((state) => state.selectedWalletId)

  return (
    <header className="sticky top-0 z-40 border-b border-line/70 bg-base/85 backdrop-blur-xl">
      <Container className="flex h-16 items-center justify-between">
        <Logo />

        <nav className="flex items-center gap-2" aria-label="Main">
          {isAuthenticated && selectedWalletId ? (
            <ButtonLink
              to={`/wallets/${selectedWalletId}`}
              variant="secondary"
              size="sm"
            >
              Open dashboard
            </ButtonLink>
          ) : (
            <Link
              to="/sign-in"
              className="rounded-md px-3 py-2 text-sm text-fg-muted transition-colors hover:text-fg"
            >
              Sign in
            </Link>
          )}

          <ButtonLink
            to="/analyze"
            size="sm"
            iconRight={<ArrowRight className="size-3.5" />}
          >
            Analyze your wallet
          </ButtonLink>
        </nav>
      </Container>
    </header>
  )
}

/** Compact footer for public pages. */
export function SiteFooter() {
  return (
    <footer className="border-t border-line/70 py-10">
      <Container className="flex flex-col gap-4 text-[13px] text-fg-subtle sm:flex-row sm:items-center sm:justify-between">
        <p>
          MaxAI Crypto — AI financial intelligence for crypto portfolios.
        </p>
        <p className="max-w-md">
          Read-only analysis of public blockchain addresses. Not financial
          advice, and not a trading service.
        </p>
      </Container>
    </footer>
  )
}
