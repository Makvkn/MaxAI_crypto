import { useEffect, useState } from 'react'
import { Navigate, useParams } from 'react-router-dom'
import { SyncStatus, type WalletPosition } from '@/api/types'
import { Button, ButtonLink } from '@/components/ui/Button'
import { Card, CardBody } from '@/components/ui/Card'
import { Skeleton, SkeletonRows } from '@/components/ui/Skeleton'
import { Scenario, Sparkle } from '@/components/ui/Icon'
import { ErrorState } from '@/components/feedback/States'
import { AppShell } from '@/components/layout/AppShell'
import { Container } from '@/components/layout/Container'
import {
  DataQualityBanner,
  DataQualityFootnotes,
} from '@/components/data-quality/DataQualityBanner'
import { StaleDataWarning } from '@/components/data-quality/StaleDataWarning'
import { WalletProvider } from '@/features/wallets/WalletProvider'
import { useActiveWallet } from '@/features/wallets/walletContext'
import { InitialSyncScreen } from '@/features/wallets/components/InitialSyncScreen'
import {
  PortfolioHeader,
  PortfolioHeaderSkeleton,
} from '@/features/portfolio/components/PortfolioHeader'
import { AllocationCard } from '@/features/portfolio/components/AllocationCard'
import {
  usePortfolio,
  useRefreshWalletData,
} from '@/features/portfolio/hooks/usePortfolio'
import { PerformanceCard } from '@/features/performance/components/PerformanceCard'
import { AssetsTable } from '@/features/assets/components/AssetsTable'
import { TransactionsCard } from '@/features/transactions/components/TransactionsCard'
import { TransactionDetailDialog } from '@/features/transactions/components/TransactionDetailDialog'
import { AiPanel } from '@/features/ai/components/AiPanel'
import { suggestedQuestions } from '@/features/ai/suggestions'
import { ScenarioDialog } from '@/features/scenarios/components/ScenarioDialog'
import { analytics } from '@/lib/analytics/analytics'
import { usePreferencesStore } from '@/stores/preferencesStore'
import { useUiStore } from '@/stores/uiStore'
import { cn } from '@/lib/utils/cn'

/**
 * Wallet dashboard.
 *
 * Facts on the left, intelligence on the right. Every figure is rendered from a
 * backend response; the page composes features and owns no financial logic.
 */
export function WalletDashboardPage() {
  const { walletId } = useParams<{ walletId: string }>()
  if (!walletId) return <Navigate to="/dashboard" replace />

  return (
    <WalletProvider walletId={walletId}>
      <AppShell walletId={walletId}>
        <DashboardRouter />
      </AppShell>
    </WalletProvider>
  )
}

/** Chooses between the synchronisation screen and the analysis itself. */
function DashboardRouter() {
  const { wallet, walletId, isLoading, error, isAnalysed, refetch } =
    useActiveWallet()

  useEffect(() => {
    if (!wallet) return
    if (wallet.sync.status === SyncStatus.READY || wallet.sync.status === SyncStatus.PARTIAL) {
      analytics.trackOnce(`sync_completed:${walletId}`, 'sync_completed', {
        wallet_id: walletId,
        sync_status: wallet.sync.status,
      })
    }
  }, [wallet, walletId])

  if (isLoading) {
    return (
      <Container size="wide" className="py-8">
        <PortfolioHeaderSkeleton />
      </Container>
    )
  }

  if (error || !wallet) {
    return (
      <Container size="wide" className="py-16">
        <Card>
          <ErrorState error={error} onRetry={refetch} />
          <CardBody className="border-t border-line text-center">
            <ButtonLink to="/analyze" variant="secondary" size="sm">
              Analyze a wallet
            </ButtonLink>
          </CardBody>
        </Card>
      </Container>
    )
  }

  if (!isAnalysed) {
    return (
      <Container size="wide">
        <InitialSyncScreen wallet={wallet} onCheckAgain={refetch} />
      </Container>
    )
  }

  return <Dashboard />
}

function Dashboard() {
  const { walletId, wallet } = useActiveWallet()
  const period = usePreferencesStore((state) => state.performancePeriod)
  const setPeriod = usePreferencesStore((state) => state.setPerformancePeriod)
  const selectedTransactionId = useUiStore(
    (state) => state.selectedTransactionId,
  )
  const selectTransaction = useUiStore((state) => state.selectTransaction)
  const aiPanelOpen = useUiStore((state) => state.aiPanelOpen)
  const toggleAiPanel = useUiStore((state) => state.toggleAiPanel)

  const portfolioQuery = usePortfolio(walletId)
  const refresh = useRefreshWalletData(walletId)
  const [isRefreshing, setIsRefreshing] = useState(false)
  const [scenarioPosition, setScenarioPosition] =
    useState<WalletPosition | null>(null)
  const [scenarioOpen, setScenarioOpen] = useState(false)

  const portfolio = portfolioQuery.data

  useEffect(() => {
    if (!portfolio) return
    analytics.trackOnce(`portfolio_loaded:${walletId}`, 'portfolio_loaded', {
      wallet_id: walletId,
      data_quality: portfolio.data_quality,
      valuation_status: portfolio.valuation_status,
    })
  }, [portfolio, walletId])

  const onRefresh = () => {
    setIsRefreshing(true)
    void refresh().finally(() => setIsRefreshing(false))
  }

  const openScenario = (position: WalletPosition | null) => {
    setScenarioPosition(position)
    setScenarioOpen(true)
  }

  return (
    <Container size="wide" className="py-7 lg:py-9">
      {portfolioQuery.isPending ? (
        <PortfolioHeaderSkeleton />
      ) : portfolioQuery.isError ? (
        <Card>
          <ErrorState
            error={portfolioQuery.error}
            onRetry={() => void portfolioQuery.refetch()}
            retrying={portfolioQuery.isFetching}
          />
        </Card>
      ) : portfolio && wallet ? (
        <PortfolioHeader
          wallet={wallet}
          portfolio={portfolio}
          onRefresh={onRefresh}
          isRefreshing={isRefreshing}
        />
      ) : null}

      {portfolio ? (
        <div className="mt-6 space-y-3">
          <DataQualityBanner notices={portfolio.notices} />
          <StaleDataWarning
            freshness={portfolio.data_freshness}
            lastSyncedAt={portfolio.last_synced_at}
            onRefresh={onRefresh}
            refreshing={isRefreshing}
          />
        </div>
      ) : null}

      <div className="mt-6 flex items-center justify-between gap-3 lg:hidden">
        <Button
          variant="secondary"
          size="sm"
          onClick={toggleAiPanel}
          iconLeft={<Sparkle className="size-3.5" />}
        >
          {aiPanelOpen ? 'Hide AI' : 'Ask AI'}
        </Button>
        <Button
          variant="quiet"
          size="sm"
          onClick={() => openScenario(null)}
          iconLeft={<Scenario className="size-3.5" />}
        >
          Simulate
        </Button>
      </div>

      <div className="mt-6 grid items-start gap-5 lg:grid-cols-[minmax(0,1fr)_384px]">
        <div className="min-w-0 space-y-5">
          <PerformanceCard
            walletId={walletId}
            period={period}
            onPeriodChange={(next) => {
              setPeriod(next)
              analytics.track('performance_period_changed', {
                wallet_id: walletId,
                period: next,
              })
            }}
          />

          {portfolioQuery.isPending ? (
            <Card>
              <CardBody className="space-y-3 py-6">
                <Skeleton className="h-3 w-24" />
                <SkeletonRows rows={4} />
              </CardBody>
            </Card>
          ) : portfolio ? (
            <>
              <AllocationCard portfolio={portfolio} />
              <AssetsTable portfolio={portfolio} onSimulate={openScenario} />
              <DataQualityFootnotes notices={portfolio.notices} />
            </>
          ) : null}

          <TransactionsCard
            walletId={walletId}
            onSelect={(transactionId) => {
              selectTransaction(transactionId)
            }}
          />
        </div>

        <aside className={cn('lg:block', aiPanelOpen ? 'block' : 'hidden')}>
          <div className="lg:sticky lg:top-[88px]">
            <AiPanel
              walletId={walletId}
              suggestions={suggestedQuestions(portfolio, period)}
              className="h-[560px] lg:h-[calc(100dvh-120px)]"
            />
            <div className="mt-3 hidden lg:block">
              <Button
                variant="secondary"
                size="sm"
                className="w-full"
                onClick={() => openScenario(null)}
                iconLeft={<Scenario className="size-3.5" />}
              >
                Run a scenario
              </Button>
            </div>
          </div>
        </aside>
      </div>

      <TransactionDetailDialog
        walletId={walletId}
        transactionId={selectedTransactionId}
        onClose={() => selectTransaction(null)}
      />

      {scenarioOpen ? (
        <ScenarioDialog
          walletId={walletId}
          portfolio={portfolio}
          initialPosition={scenarioPosition}
          open={scenarioOpen}
          onClose={() => setScenarioOpen(false)}
        />
      ) : null}
    </Container>
  )
}
