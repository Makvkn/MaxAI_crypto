import { useState, type FormEvent } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { Button } from '@/components/ui/Button'
import { Card, CardBody } from '@/components/ui/Card'
import { Input } from '@/components/ui/Input'
import { ErrorState } from '@/components/feedback/States'
import { Container } from '@/components/layout/Container'
import { Logo } from '@/components/layout/Logo'
import { useSession } from '@/features/auth/sessionContext'
import { usePreferencesStore } from '@/stores/preferencesStore'

/**
 * Sign in, register, or attach credentials to a guest account.
 *
 * Upgrading keeps the same `user.id` server-side, so an analysis started as a
 * guest survives the transition. The backend owns every decision here.
 */
export function SignInPage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const {
    isGuest,
    signInWithEmail,
    registerWithEmail,
    signInWithGoogle,
    upgradeWithEmail,
    upgradeWithGoogle,
    isMutating,
  } = useSession()
  const selectedWalletId = usePreferencesStore((state) => state.selectedWalletId)

  const isUpgrade = searchParams.get('upgrade') === '1' && isGuest
  const [mode, setMode] = useState<'sign-in' | 'register'>(
    isUpgrade ? 'register' : 'sign-in',
  )
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<unknown>(null)

  const goBack = () => {
    navigate(selectedWalletId ? `/wallets/${selectedWalletId}` : '/', {
      replace: true,
    })
  }

  const run = async (action: () => Promise<unknown>) => {
    setError(null)
    try {
      await action()
      goBack()
    } catch (caught) {
      setError(caught)
    }
  }

  const onSubmit = (event: FormEvent) => {
    event.preventDefault()
    const credentials = { email: email.trim(), password }

    void run(() =>
      isUpgrade
        ? upgradeWithEmail(credentials)
        : mode === 'register'
          ? registerWithEmail(credentials)
          : signInWithEmail(credentials),
    )
  }

  return (
    <div className="min-h-dvh bg-base">
      <header className="border-b border-line/70">
        <Container className="flex h-16 items-center">
          <Logo />
        </Container>
      </header>

      <Container size="narrow" className="py-14 lg:py-20">
        <div className="mx-auto max-w-md">
          <h1 className="text-2xl font-medium tracking-tight text-fg">
            {isUpgrade
              ? 'Save your analysis'
              : mode === 'register'
                ? 'Create an account'
                : 'Sign in'}
          </h1>
          <p className="mt-2 text-sm leading-relaxed text-fg-muted">
            {isUpgrade
              ? 'Attach an email or Google account to keep this wallet analysis. Your existing data stays exactly as it is.'
              : 'Your wallets, portfolio history and conversations are tied to your account.'}
          </p>

          <Card className="mt-7">
            <CardBody className="p-6">
              <Button
                variant="secondary"
                size="lg"
                className="w-full"
                loading={isMutating}
                onClick={() =>
                  void run(() =>
                    isUpgrade ? upgradeWithGoogle() : signInWithGoogle(),
                  )
                }
              >
                Continue with Google
              </Button>

              <div className="my-5 flex items-center gap-3 text-[11px] tracking-[0.08em] text-fg-subtle uppercase">
                <span className="h-px flex-1 bg-line" />
                or
                <span className="h-px flex-1 bg-line" />
              </div>

              <form onSubmit={onSubmit} className="space-y-4" noValidate>
                <Input
                  label="Email"
                  type="email"
                  value={email}
                  onChange={(event) => setEmail(event.target.value)}
                  autoComplete="email"
                  required
                />
                <Input
                  label="Password"
                  type="password"
                  value={password}
                  onChange={(event) => setPassword(event.target.value)}
                  autoComplete={
                    mode === 'register' ? 'new-password' : 'current-password'
                  }
                  required
                />

                {error ? <ErrorState error={error} compact /> : null}

                <Button
                  type="submit"
                  size="lg"
                  className="w-full"
                  loading={isMutating}
                >
                  {isUpgrade
                    ? 'Save my analysis'
                    : mode === 'register'
                      ? 'Create account'
                      : 'Sign in'}
                </Button>
              </form>
            </CardBody>
          </Card>

          {!isUpgrade ? (
            <p className="mt-5 text-center text-[13px] text-fg-subtle">
              {mode === 'register'
                ? 'Already have an account?'
                : "Don't have an account?"}{' '}
              <button
                type="button"
                onClick={() =>
                  setMode(mode === 'register' ? 'sign-in' : 'register')
                }
                className="rounded-md text-accent underline decoration-accent/40 transition-colors hover:decoration-accent"
              >
                {mode === 'register' ? 'Sign in' : 'Create one'}
              </button>
            </p>
          ) : null}

          <p className="mt-8 text-center text-[13px] text-fg-subtle">
            You can also{' '}
            <button
              type="button"
              onClick={() => navigate('/analyze')}
              className="rounded-md text-fg-muted underline decoration-line-strong transition-colors hover:text-fg"
            >
              analyse a wallet without an account
            </button>
            .
          </p>
        </div>
      </Container>
    </div>
  )
}
