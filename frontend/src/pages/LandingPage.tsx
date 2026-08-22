import { useEffect } from 'react'
import { ButtonLink } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import {
  ArrowRight,
  Check,
  Layers,
  Question,
  Scenario,
  Shield,
  Sparkle,
} from '@/components/ui/Icon'
import { Container } from '@/components/layout/Container'
import { SiteFooter, SiteHeader } from '@/components/layout/SiteHeader'
import { ChainMonogram } from '@/features/wallets/components/ChainMonogram'
import { SUPPORTED_CHAINS } from '@/app/config/chains'
import { analytics } from '@/lib/analytics/analytics'

/**
 * Landing page.
 *
 * The job of this page is to make one distinction obvious: a tracker shows what
 * you own, MaxAI Crypto explains what it means. Everything else supports the
 * single call to action — analyse a wallet.
 */
export function LandingPage() {
  useEffect(() => {
    analytics.track('landing_visit', { referrer: document.referrer || undefined })
  }, [])

  return (
    <div className="min-h-dvh bg-base">
      <SiteHeader />
      <main>
        <Hero />
        <ThreeQuestions />
        <Difference />
        <HowItWorks />
        <Networks />
        <Boundaries />
        <FinalCta />
      </main>
      <SiteFooter />
    </div>
  )
}

function Hero() {
  return (
    <section className="relative overflow-hidden border-b border-line/70">
      <div
        className="pointer-events-none absolute inset-0 grid-noise opacity-[0.35]"
        aria-hidden="true"
      />
      <div
        className="pointer-events-none absolute -top-40 left-1/2 h-80 w-[52rem] -translate-x-1/2 rounded-full bg-accent/8 blur-3xl"
        aria-hidden="true"
      />

      <Container className="relative py-20 lg:py-28">
        <div className="grid gap-14 lg:grid-cols-[1.05fr_0.95fr] lg:items-center">
          <div>
            <p className="text-[13px] font-medium tracking-[0.14em] text-accent uppercase">
              AI Financial Intelligence for Crypto
            </p>

            <h1 className="mt-5 text-balance text-4xl leading-[1.05] font-medium tracking-[-0.03em] text-fg sm:text-5xl lg:text-display">
              Understand your crypto,
              <br />
              <span className="text-fg-muted">not just track it.</span>
            </h1>

            <p className="mt-6 max-w-xl text-[17px] leading-relaxed text-fg-muted">
              Enter a public wallet address and MaxAI Crypto turns blockchain,
              market and portfolio history into plain answers: what your
              portfolio is doing, why it moved, and what a price change would
              mean for it.
            </p>

            <div className="mt-9 flex flex-wrap items-center gap-3">
              <ButtonLink
                to="/analyze"
                size="lg"
                iconRight={<ArrowRight className="size-4" />}
              >
                Analyze your wallet
              </ButtonLink>
              <ButtonLink to="/sign-in" size="lg" variant="quiet">
                Sign in
              </ButtonLink>
            </div>

            <p className="mt-6 flex items-center gap-2 text-[13px] text-fg-subtle">
              <Shield className="size-4 shrink-0 text-positive" />
              Read-only. Public addresses only — never private keys, seed
              phrases or signatures.
            </p>
          </div>

          <HeroPreview />
        </div>
      </Container>
    </section>
  )
}

/**
 * Illustrative preview.
 *
 * Explicitly labelled as an example so no visitor mistakes these numbers for
 * live data.
 */
function HeroPreview() {
  return (
    <Card className="animate-rise overflow-hidden bg-surface/80 backdrop-blur">
      <div className="flex items-center justify-between border-b border-line px-5 py-3">
        <span className="text-[11px] font-medium tracking-[0.08em] text-fg-subtle uppercase">
          Example analysis
        </span>
        <Badge tone="accent" icon={<Sparkle className="size-3" />}>
          AI insight
        </Badge>
      </div>

      <div className="grid gap-0 sm:grid-cols-2">
        <div className="border-line px-5 py-5 sm:border-r">
          <p className="text-[11px] tracking-[0.08em] text-fg-subtle uppercase">
            Portfolio
          </p>
          <p className="mt-2 text-3xl font-medium tracking-tight text-fg tabular">
            $24,850
          </p>
          <p className="mt-1 text-sm text-negative tabular">-4.21% · 24h</p>

          <ul className="mt-5 space-y-2.5 text-sm">
            {[
              ['ETH', '52%', '$12,926'],
              ['WBTC', '21%', '$5,216'],
              ['LINK', '9%', '$2,236'],
            ].map(([symbol, share, value]) => (
              <li key={symbol} className="flex items-center justify-between gap-3">
                <span className="text-fg-muted">{symbol}</span>
                <span className="text-fg-subtle tabular">{share}</span>
                <span className="text-fg tabular">{value}</span>
              </li>
            ))}
          </ul>
        </div>

        <div className="bg-base-elevated/60 px-5 py-5">
          <p className="text-[11px] tracking-[0.08em] text-fg-subtle uppercase">
            Why is my portfolio down?
          </p>
          <p className="mt-3 text-[13px] leading-relaxed text-fg-muted">
            Your portfolio is down 4.21% over the last 24 hours, a change of
            −$1,092. ETH accounts for the largest share of that decline at
            −$516, which follows from it being 52% of the total rather than an
            unusual event.
          </p>
          <div className="mt-4 flex flex-wrap gap-1.5">
            {['ETH', 'LINK', 'performance · 24h'].map((chip) => (
              <span
                key={chip}
                className="rounded-md border border-line-strong bg-surface-raised px-2 py-1 text-[11px] text-fg-subtle"
              >
                {chip}
              </span>
            ))}
          </div>
        </div>
      </div>
    </Card>
  )
}

function ThreeQuestions() {
  const items = [
    {
      icon: <Layers className="size-4" />,
      title: 'What is happening?',
      body: 'Total value, allocation, per-asset balances and 24h movement — every figure calculated server-side from your actual positions.',
    },
    {
      icon: <Question className="size-4" />,
      title: 'Why is it happening?',
      body: 'Which assets drove the change and by how much, based on stored portfolio snapshots rather than guesswork.',
    },
    {
      icon: <Scenario className="size-4" />,
      title: 'What happens if…?',
      body: 'Move one asset by a percentage and see the deterministic effect on your portfolio, explained in words.',
    },
  ]

  return (
    <section className="border-b border-line/70 py-16 lg:py-20">
      <Container>
        <div className="grid gap-8 sm:grid-cols-3">
          {items.map((item) => (
            <div key={item.title}>
              <span className="grid size-9 place-items-center rounded-lg border border-line-strong bg-surface-raised text-accent">
                {item.icon}
              </span>
              <h2 className="mt-4 text-[15px] font-medium text-fg">
                {item.title}
              </h2>
              <p className="mt-2 text-sm leading-relaxed text-fg-muted">
                {item.body}
              </p>
            </div>
          ))}
        </div>
      </Container>
    </section>
  )
}

function Difference() {
  return (
    <section className="border-b border-line/70 py-16 lg:py-24">
      <Container>
        <h2 className="max-w-2xl text-balance text-2xl leading-snug font-medium tracking-tight text-fg sm:text-3xl">
          A portfolio tracker answers one question. This answers the ones that
          follow.
        </h2>

        <div className="mt-10 grid gap-4 lg:grid-cols-2">
          <Card className="p-6">
            <p className="text-[11px] tracking-[0.08em] text-fg-subtle uppercase">
              A normal portfolio tracker
            </p>
            <p className="mt-3 text-lg text-fg-muted">What do I own?</p>
            <ul className="mt-5 space-y-2.5 text-sm text-fg-subtle">
              <li>A list of balances and prices</li>
              <li>A chart you have to interpret yourself</li>
              <li>No explanation of what changed, or why</li>
              <li>No way to test a "what if"</li>
            </ul>
          </Card>

          <Card className="border-accent/25 bg-accent-quiet/25 p-6">
            <p className="text-[11px] tracking-[0.08em] text-accent uppercase">
              MaxAI Crypto
            </p>
            <p className="mt-3 text-lg text-fg">
              What do I own — and what does it mean?
            </p>
            <ul className="mt-5 space-y-2.5 text-sm text-fg-muted">
              {[
                'The facts, calculated by the backend and clearly labelled',
                'The reason a period moved, attributed per asset',
                'Plain-language explanations of individual transactions',
                'Deterministic scenarios: “what if ETH falls 20%?”',
              ].map((line) => (
                <li key={line} className="flex items-start gap-2.5">
                  <Check className="mt-0.5 size-4 shrink-0 text-positive" />
                  {line}
                </li>
              ))}
            </ul>
          </Card>
        </div>

        <p className="mt-8 max-w-2xl text-sm leading-relaxed text-fg-subtle">
          The split is deliberate: the dashboard is facts, the AI is
          intelligence. Numbers come from deterministic calculations, and the
          model explains them — it never invents them.
        </p>
      </Container>
    </section>
  )
}

function HowItWorks() {
  const steps = [
    {
      title: 'Select a network',
      body: 'Choose the blockchain your wallet is on. You stay in control of what gets analysed.',
    },
    {
      title: 'Enter a public address',
      body: 'No connection, no signature, no keys. A public address is all that is needed to read.',
    },
    {
      title: 'Understand your portfolio',
      body: 'Balances, valuation, history and AI answers — as soon as the first synchronisation finishes.',
    },
  ]

  return (
    <section className="border-b border-line/70 py-16 lg:py-20">
      <Container>
        <div className="grid gap-10 lg:grid-cols-[0.8fr_1.2fr] lg:items-start">
          <div>
            <h2 className="text-2xl font-medium tracking-tight text-fg">
              Three steps to an answer
            </h2>
            <p className="mt-3 max-w-md text-sm leading-relaxed text-fg-muted">
              The first analysis runs as a background job, so heavy blockchain
              work never blocks the screen. You see the real stage it has
              reached — never a fake progress bar.
            </p>
          </div>

          <ol className="space-y-3">
            {steps.map((step, index) => (
              <li key={step.title}>
                <Card className="flex items-start gap-4 p-5">
                  <span className="grid size-7 shrink-0 place-items-center rounded-full border border-line-strong bg-base-elevated text-[12px] text-fg-muted tabular">
                    {index + 1}
                  </span>
                  <div>
                    <p className="text-[15px] font-medium text-fg">
                      {step.title}
                    </p>
                    <p className="mt-1.5 text-sm leading-relaxed text-fg-muted">
                      {step.body}
                    </p>
                  </div>
                </Card>
              </li>
            ))}
          </ol>
        </div>
      </Container>
    </section>
  )
}

function Networks() {
  return (
    <section className="border-b border-line/70 py-16">
      <Container>
        <h2 className="text-[13px] font-medium tracking-[0.1em] text-fg-subtle uppercase">
          Supported networks
        </h2>
        <ul className="mt-6 grid grid-cols-2 gap-3 sm:grid-cols-4">
          {SUPPORTED_CHAINS.map((chain) => (
            <li
              key={chain.id}
              className="flex items-center gap-3 rounded-card border border-line bg-surface px-4 py-3"
            >
              <ChainMonogram chainId={chain.id} size="sm" />
              <span className="min-w-0">
                <span className="block truncate text-sm text-fg">
                  {chain.name}
                </span>
                <span className="block text-[12px] text-fg-subtle">
                  {chain.nativeSymbol}
                </span>
              </span>
            </li>
          ))}
        </ul>
      </Container>
    </section>
  )
}

function Boundaries() {
  const items = [
    {
      title: 'Read-only by design',
      body: 'There is no send, swap, bridge or trade anywhere in this product. It cannot move your funds because it never has the ability to.',
    },
    {
      title: 'Analysis, not advice',
      body: 'The AI analyses, explains, compares and simulates. It will not tell you to buy or sell, and it will say so if you ask.',
    },
    {
      title: 'Honest about data',
      body: 'If a price is unknown, you see the balance and no value. If data is stale, it says how old it is. Nothing is quietly turned into zero.',
    },
  ]

  return (
    <section className="border-b border-line/70 py-16 lg:py-20">
      <Container>
        <div className="grid gap-8 sm:grid-cols-3">
          {items.map((item) => (
            <div key={item.title}>
              <h3 className="text-[15px] font-medium text-fg">{item.title}</h3>
              <p className="mt-2 text-sm leading-relaxed text-fg-muted">
                {item.body}
              </p>
            </div>
          ))}
        </div>
      </Container>
    </section>
  )
}

function FinalCta() {
  return (
    <section className="py-20">
      <Container className="text-center">
        <h2 className="text-balance text-3xl font-medium tracking-tight text-fg sm:text-4xl">
          Start with one wallet.
        </h2>
        <p className="mx-auto mt-4 max-w-lg text-[15px] leading-relaxed text-fg-muted">
          Select a network, paste a public address, and see what your portfolio
          is actually doing.
        </p>
        <div className="mt-8 flex justify-center">
          <ButtonLink
            to="/analyze"
            size="lg"
            iconRight={<ArrowRight className="size-4" />}
          >
            Analyze your wallet
          </ButtonLink>
        </div>
      </Container>
    </section>
  )
}
