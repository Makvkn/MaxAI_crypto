import { ButtonLink } from '@/components/ui/Button'
import { Container } from '@/components/layout/Container'
import { Logo } from '@/components/layout/Logo'

export function NotFoundPage() {
  return (
    <div className="min-h-dvh bg-base">
      <header className="border-b border-line/70">
        <Container className="flex h-16 items-center">
          <Logo />
        </Container>
      </header>

      <Container size="narrow" className="py-24 text-center">
        <h1 className="text-2xl font-medium tracking-tight text-fg">
          This page does not exist
        </h1>
        <p className="mt-3 text-sm text-fg-muted">
          The link may be out of date, or the wallet may no longer be on your
          account.
        </p>
        <div className="mt-8 flex justify-center gap-3">
          <ButtonLink to="/">Back to home</ButtonLink>
          <ButtonLink to="/analyze" variant="secondary">
            Analyze a wallet
          </ButtonLink>
        </div>
      </Container>
    </div>
  )
}
