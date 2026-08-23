import { Suspense, lazy, useEffect, type ReactNode } from 'react'
import {
  BrowserRouter,
  Navigate,
  Route,
  Routes,
  useLocation,
} from 'react-router-dom'
import { Spinner } from '@/components/ui/Spinner'
import { useSession } from '@/features/auth/sessionContext'
import { LandingPage } from '@/pages/LandingPage'
import { AnalyzePage } from '@/pages/AnalyzePage'
import { SignInPage } from '@/pages/SignInPage'
import { NotFoundPage } from '@/pages/NotFoundPage'
import { analytics } from '@/lib/analytics/analytics'
import { usePreferencesStore } from '@/stores/preferencesStore'

// The dashboard pulls in the charting library, so it is split out of the
// landing bundle.
const WalletDashboardPage = lazy(async () => ({
  default: (await import('@/pages/WalletDashboardPage')).WalletDashboardPage,
}))

export function AppRouter() {
  return (
    <BrowserRouter>
      <RouteTracker />
      <Suspense fallback={<RouteFallback />}>
        <Routes>
          <Route
            path="/"
            element={
              <GuestOnly>
                <LandingPage />
              </GuestOnly>
            }
          />
          <Route path="/analyze" element={<AnalyzePage />} />
          <Route path="/sign-in" element={<SignInPage />} />
          <Route path="/dashboard" element={<DashboardEntry />} />
          <Route
            path="/wallets/:walletId"
            element={
              <RequireSession>
                <WalletDashboardPage />
              </RequireSession>
            }
          />
          <Route path="*" element={<NotFoundPage />} />
        </Routes>
      </Suspense>
    </BrowserRouter>
  )
}

/**
 * Session gate.
 *
 * UX only: it avoids rendering a dashboard that has no session to query with.
 * Authorisation itself stays with the backend on every request.
 */
function RequireSession({ children }: { children: ReactNode }) {
  const { isAuthenticated } = useSession()
  if (!isAuthenticated) return <Navigate to="/analyze" replace />
  return <>{children}</>
}

/** Landing is marketing for signed-out visitors only. */
function GuestOnly({ children }: { children: ReactNode }) {
  const { isAuthenticated } = useSession()
  if (isAuthenticated) return <Navigate to="/dashboard" replace />
  return <>{children}</>
}

/** `/dashboard` resolves to whichever wallet is currently selected. */
function DashboardEntry() {
  const selectedWalletId = usePreferencesStore((state) => state.selectedWalletId)
  return (
    <Navigate
      to={selectedWalletId ? `/wallets/${selectedWalletId}` : '/analyze'}
      replace
    />
  )
}

function RouteTracker() {
  const location = useLocation()
  const hasCompletedOnboarding = usePreferencesStore(
    (state) => state.hasCompletedOnboarding,
  )
  const selectedWalletId = usePreferencesStore((state) => state.selectedWalletId)

  useEffect(() => {
    analytics.page(location.pathname)
  }, [location.pathname])

  useEffect(() => {
    if (!hasCompletedOnboarding) return
    analytics.trackOnce('return_visit', 'return_visit', {
      wallet_id: selectedWalletId ?? undefined,
    })
  }, [hasCompletedOnboarding, selectedWalletId])

  return null
}

function RouteFallback() {
  return (
    <div className="grid min-h-dvh place-items-center bg-base">
      <Spinner className="size-5 text-accent" />
      <span className="sr-only">Loading</span>
    </div>
  )
}
