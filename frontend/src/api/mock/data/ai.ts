import {
  AIEvidenceType,
  AIIntent,
  AIReferenceType,
  AIToolName,
  AssetVisibility,
  DataFreshness,
  DataQuality,
  PerformanceStatus,
  TransactionStatus,
  TransactionType,
  type AIClaim,
  type AIReference,
  type AIResponse,
  type Portfolio,
  type PortfolioPerformance,
  type ScenarioResult,
  type Transaction,
  type WalletPosition,
} from '../../types'
import { syncAgeMinutes } from './portfolio'
import type { MockVariant } from '../variants'

/**
 * AI answers — MOCK BACKEND SIMULATION of the AI Orchestrator.
 *
 * Every number in an answer is taken from data the backend already computed;
 * nothing here derives new financial values. Intent routing, tool selection
 * and the answer text are all server-side concerns being stood in for.
 */

export interface AiContext {
  question: string
  portfolio: Portfolio
  performance: PortfolioPerformance
  transaction?: Transaction | null
  scenario?: ScenarioResult | null
  variant: MockVariant
}

const NEWS_PATTERNS = [
  'news',
  'ecosystem',
  'today in crypto',
  'twitter',
  'announcement',
  'regulation',
  'hack',
  'listing',
]
const ADVICE_PATTERNS = ['should i', 'buy', 'sell', 'short', 'long', 'leverage']

export function resolveIntent(
  question: string,
  context: { hasTransaction: boolean; hasScenario: boolean },
): AIIntent {
  if (context.hasScenario) return AIIntent.SCENARIO_SIMULATION
  if (context.hasTransaction) return AIIntent.TRANSACTION_EXPLANATION

  const text = question.toLowerCase()

  if (NEWS_PATTERNS.some((pattern) => text.includes(pattern))) {
    return AIIntent.UNSUPPORTED
  }
  if (text.includes('what if') || text.includes('drops') || text.includes('falls')) {
    return AIIntent.SCENARIO_SIMULATION
  }
  if (
    text.includes('allocation') ||
    text.includes('largest') ||
    text.includes('concentration') ||
    text.includes('diversif')
  ) {
    return AIIntent.PORTFOLIO_ALLOCATION
  }
  if (
    text.includes('week') ||
    text.includes('month') ||
    text.includes('performance') ||
    text.includes('down') ||
    text.includes('up ') ||
    text.includes('lose') ||
    text.includes('lost')
  ) {
    return AIIntent.PORTFOLIO_PERFORMANCE
  }
  if (text.includes('summary') || text.includes('overview') || text.includes('worth')) {
    return AIIntent.PORTFOLIO_SUMMARY
  }
  return AIIntent.GENERAL_PORTFOLIO_QUESTION
}

/** Tools the orchestrator would run for an intent, in execution order. */
export function toolsForIntent(intent: AIIntent): string[] {
  switch (intent) {
    case AIIntent.PORTFOLIO_SUMMARY:
      return [AIToolName.GET_PORTFOLIO, AIToolName.GET_PORTFOLIO_PERFORMANCE]
    case AIIntent.PORTFOLIO_PERFORMANCE:
      return [
        AIToolName.GET_PORTFOLIO,
        AIToolName.GET_PORTFOLIO_PERFORMANCE,
        AIToolName.GET_HISTORICAL_PORTFOLIO,
      ]
    case AIIntent.PORTFOLIO_ALLOCATION:
      return [AIToolName.GET_POSITIONS]
    case AIIntent.TRANSACTION_EXPLANATION:
      return [AIToolName.GET_TRANSACTION, AIToolName.GET_ASSET_PRICE]
    case AIIntent.SCENARIO_SIMULATION:
      return [AIToolName.GET_PORTFOLIO, AIToolName.SIMULATE_SCENARIO]
    case AIIntent.GENERAL_PORTFOLIO_QUESTION:
      return [AIToolName.GET_PORTFOLIO, AIToolName.GET_POSITIONS]
    case AIIntent.UNSUPPORTED:
      return []
    default:
      return [AIToolName.GET_PORTFOLIO]
  }
}

export function buildAiResponse(context: AiContext): AIResponse {
  const intent = resolveIntent(context.question, {
    hasTransaction: Boolean(context.transaction),
    hasScenario: Boolean(context.scenario),
  })

  if (intent === AIIntent.UNSUPPORTED) {
    return {
      intent,
      data_quality: context.portfolio.data_quality,
      answer:
        'I can only reason about your portfolio, its history and the market data needed to value it. Market news and ecosystem events are outside what I can verify right now, so I would be guessing — and I would rather not. Ask me why your portfolio moved, how it is allocated, or what a price change would do to it.',
      claims: [],
      references: [],
      unsupported_reason: 'NEWS_INTELLIGENCE_NOT_SUPPORTED',
    }
  }

  const body = (() => {
    switch (intent) {
      case AIIntent.TRANSACTION_EXPLANATION:
        return explainTransaction(context)
      case AIIntent.SCENARIO_SIMULATION:
        return explainScenario(context)
      case AIIntent.PORTFOLIO_ALLOCATION:
        return explainAllocation(context)
      case AIIntent.PORTFOLIO_PERFORMANCE:
        return explainPerformance(context)
      default:
        return explainSummary(context)
    }
  })()

  const qualifier = qualityQualifier(context)
  const advisory = ADVICE_PATTERNS.some((pattern) =>
    context.question.toLowerCase().includes(pattern),
  )
    ? ' I can\u2019t tell you whether to buy or sell — that is not what this product does. What I can do is show how this position affects your concentration and simulate price scenarios for it.'
    : ''

  return {
    intent,
    data_quality: context.portfolio.data_quality,
    answer: `${body.answer}${qualifier}${advisory}`,
    claims: body.claims,
    references: body.references,
    unsupported_reason: null,
  }
}

/* -------------------------------------------------------------------------- */

interface AnswerBody {
  answer: string
  claims: AIClaim[]
  references: AIReference[]
}

function explainSummary(context: AiContext): AnswerBody {
  const { portfolio } = context
  const top = visiblePositions(portfolio).slice(0, 3)

  if (portfolio.total_value_usd === null) {
    return {
      answer:
        'I can\u2019t value this portfolio at the moment: the pricing data behind it is unavailable, so any total I produced would be invented. The balances themselves are still accurate.',
      claims: [],
      references: [],
    }
  }

  const sentences = [
    `Your portfolio is worth ${usd(portfolio.total_value_usd)}.`,
    portfolio.change_24h_pct
      ? `It is ${direction(portfolio.change_24h_pct)} ${pct(portfolio.change_24h_pct)} over the last 24 hours, a change of ${usd(portfolio.change_24h_usd)}.`
      : 'A 24-hour comparison is not available yet.',
    top[0]
      ? `${top[0].asset.symbol} is your largest position at ${pct(top[0].allocation_pct)} of the total.`
      : '',
    describeDriver(context),
  ].filter(Boolean)

  return {
    answer: sentences.join(' '),
    claims: [
      claim(sentences[0] as string, [
        { type: AIEvidenceType.PORTFOLIO, id: portfolio.wallet_id },
      ]),
      ...(portfolio.change_24h_pct
        ? [
            claim(sentences[1] as string, [
              {
                type: AIEvidenceType.PORTFOLIO_PERFORMANCE,
                id: context.performance.calculation_id ?? 'performance_24h',
              },
            ]),
          ]
        : []),
    ],
    references: top.map(assetReference),
  }
}

function explainPerformance(context: AiContext): AnswerBody {
  const { performance } = context

  if (performance.status === PerformanceStatus.UNAVAILABLE) {
    return {
      answer:
        'I don\u2019t have enough portfolio history to measure performance for that period yet. Performance is calculated from stored snapshots, and the first ones are still being collected — it is not something I can estimate.',
      claims: [],
      references: [],
    }
  }

  const drivers = performance.drivers.slice(0, 3)
  const lead = `Over the selected period your portfolio went from ${usd(performance.opening?.value_usd ?? null)} to ${usd(performance.closing?.value_usd ?? null)}, ${direction(performance.change_pct)} ${pct(performance.change_pct)} (${usd(performance.change_usd)}).`

  const driverSentences = drivers.map(
    (driver) =>
      `${driver.asset.symbol} contributed ${usd(driver.contribution_usd)} at ${pct(driver.allocation_pct)} of the portfolio.`,
  )

  const concentration = drivers[0]
    ? `Most of the movement traces back to ${drivers[0].asset.symbol}, which is the natural consequence of it being your largest holding rather than an unusual event.`
    : ''

  return {
    answer: [lead, ...driverSentences, concentration].filter(Boolean).join(' '),
    claims: [
      claim(lead, [
        {
          type: AIEvidenceType.PORTFOLIO_PERFORMANCE,
          id: performance.calculation_id ?? 'performance',
        },
      ]),
      ...drivers.map((driver, index) =>
        claim(driverSentences[index] as string, [
          { type: AIEvidenceType.CALCULATION, id: `contribution:${driver.asset.id}` },
          { type: AIEvidenceType.POSITION, id: driver.asset.id },
        ]),
      ),
    ],
    references: drivers.map((driver) => ({
      type: AIReferenceType.ASSET,
      id: driver.asset.id,
      label: driver.asset.symbol,
    })),
  }
}

function explainAllocation(context: AiContext): AnswerBody {
  const positions = visiblePositions(context.portfolio)
  const top = positions.slice(0, 4)
  const largest = top[0]

  if (!largest) {
    return {
      answer: 'This wallet holds no priced assets, so there is no allocation to describe.',
      claims: [],
      references: [],
    }
  }

  const lead = `${largest.asset.symbol} dominates the portfolio at ${pct(largest.allocation_pct)}, worth ${usd(largest.value_usd)}.`
  const breakdown = `The next positions are ${top
    .slice(1)
    .map((position) => `${position.asset.symbol} at ${pct(position.allocation_pct)}`)
    .join(', ')}.`
  const risk = `Concentration is the thing to notice here: a move in ${largest.asset.symbol} moves roughly ${pct(largest.allocation_pct)} of your portfolio with it, so single-asset volatility becomes portfolio volatility.`

  return {
    answer: [lead, top.length > 1 ? breakdown : '', risk].filter(Boolean).join(' '),
    claims: [
      claim(lead, [{ type: AIEvidenceType.POSITION, id: largest.asset.id }]),
      claim(risk, [
        { type: AIEvidenceType.CALCULATION, id: `allocation:${largest.asset.id}` },
      ]),
    ],
    references: top.map(assetReference),
  }
}

function explainTransaction(context: AiContext): AnswerBody {
  const transaction = context.transaction
  if (!transaction) {
    return { answer: 'That transaction could not be found.', claims: [], references: [] }
  }

  const when = new Date(transaction.timestamp).toUTCString()
  const fee =
    transaction.fee_amount && transaction.fee_asset
      ? `The network fee was ${transaction.fee_amount} ${transaction.fee_asset.symbol}${
          transaction.fee_value_usd ? ` (${usd(transaction.fee_value_usd)})` : ''
        }.`
      : ''

  const lead = ((): string => {
    switch (transaction.type) {
      case TransactionType.SWAP:
        return `You swapped ${transaction.amount_out} ${transaction.asset_out?.symbol} for ${transaction.amount_in} ${transaction.asset_in?.symbol}${
          transaction.protocol ? ` through ${transaction.protocol}` : ''
        }, valued at ${usd(transaction.value_out_usd)} at the time.`
      case TransactionType.TRANSFER:
        return transaction.asset_in
          ? `You received ${transaction.amount_in} ${transaction.asset_in.symbol}${
              transaction.value_in_usd ? `, worth ${usd(transaction.value_in_usd)} at the time` : ''
            }.`
          : `You sent ${transaction.amount_out} ${transaction.asset_out?.symbol}${
              transaction.value_out_usd ? `, worth ${usd(transaction.value_out_usd)} at the time` : ''
            }.`
      case TransactionType.APPROVE:
        return `This was an approval: you granted ${transaction.protocol ?? 'a contract'} permission to move a token on your behalf. No funds moved in this transaction itself.`
      case TransactionType.STAKE:
        return `You staked ${transaction.amount_out} ${transaction.asset_out?.symbol}${
          transaction.protocol ? ` with ${transaction.protocol}` : ''
        }.`
      case TransactionType.UNSTAKE:
        return `You unstaked ${transaction.amount_in} ${transaction.asset_in?.symbol}${
          transaction.protocol ? ` from ${transaction.protocol}` : ''
        }.`
      case TransactionType.CLAIM:
        return `You claimed ${transaction.amount_in} ${transaction.asset_in?.symbol} in rewards.`
      case TransactionType.CONTRACT_INTERACTION:
        return `This was a contract interaction${
          transaction.protocol ? ` with ${transaction.protocol}` : ''
        }. It changed contract state without a simple transfer of value.`
      default:
        return 'This transaction could not be classified reliably, so I will not guess at what it did. What is certain is the on-chain record below: the hash, the block, the addresses and the fee.'
    }
  })()

  const failure =
    transaction.status === TransactionStatus.FAILED
      ? ' The transaction failed on-chain, so its intended effect did not happen — but the fee was still paid.'
      : ''

  return {
    answer: `${lead}${failure} It was recorded on ${when}. ${fee}`.trim(),
    claims: [
      claim(lead, [{ type: AIEvidenceType.TRANSACTION, id: transaction.id }]),
    ],
    references: [
      { type: AIReferenceType.TRANSACTION, id: transaction.id, label: shortHash(transaction.tx_hash) },
      ...(transaction.asset_in
        ? [{ type: AIReferenceType.ASSET, id: transaction.asset_in.id, label: transaction.asset_in.symbol }]
        : []),
      ...(transaction.asset_out
        ? [{ type: AIReferenceType.ASSET, id: transaction.asset_out.id, label: transaction.asset_out.symbol }]
        : []),
    ],
  }
}

function explainScenario(context: AiContext): AnswerBody {
  const scenario = context.scenario
  if (!scenario) {
    return {
      answer:
        'Tell me which asset to move and by how much — for example "what if ETH falls 20%" — and I will run it against your actual positions.',
      claims: [],
      references: [],
    }
  }

  const lead = `If ${scenario.asset.symbol} moves ${pct(scenario.change_pct)}, your portfolio goes from ${usd(scenario.baseline.portfolio_value_usd)} to ${usd(scenario.projection.portfolio_value_usd)} — a change of ${usd(scenario.projection.portfolio_change_usd)} (${pct(scenario.projection.portfolio_change_pct)}).`
  const why = `${scenario.asset.symbol} is ${pct(scenario.baseline.asset_allocation_pct)} of your portfolio, so a move in it is diluted by everything else you hold. The position itself changes by ${usd(scenario.projection.asset_impact_usd)}.`
  const caveat =
    'This holds every other position still. Assets tend to move together, so treat it as a sensitivity check on one variable rather than a forecast.'

  return {
    answer: [lead, why, caveat].join(' '),
    claims: [
      claim(lead, [{ type: AIEvidenceType.SCENARIO, id: scenario.calculation_id }]),
      claim(why, [
        { type: AIEvidenceType.POSITION, id: scenario.asset.id },
        { type: AIEvidenceType.CALCULATION, id: scenario.calculation_id },
      ]),
    ],
    references: [
      { type: AIReferenceType.ASSET, id: scenario.asset.id, label: scenario.asset.symbol },
      { type: AIReferenceType.SCENARIO, id: scenario.id, label: null },
    ],
  }
}

/* -------------------------------------------------------------------------- */

function describeDriver(context: AiContext): string {
  const driver = context.performance.drivers[0]
  if (!driver || driver.contribution_usd === null) return ''
  return `${driver.asset.symbol} accounts for the largest share of the 24-hour change at ${usd(driver.contribution_usd)}.`
}

/**
 * Data quality is part of the answer, not a footnote: the model must not
 * present partial or stale figures as exact.
 */
function qualityQualifier(context: AiContext): string {
  const { portfolio, variant } = context

  if (portfolio.data_quality === DataQuality.PARTIAL) {
    const count = portfolio.exclusions.unpriced_positions
    return ` One thing to keep in mind: ${count === 1 ? 'one asset has' : `${count} assets have`} no reliable price, so ${count === 1 ? 'it is' : 'they are'} excluded from these totals.`
  }

  if (
    portfolio.data_freshness === DataFreshness.STALE ||
    portfolio.data_freshness === DataFreshness.VERY_STALE
  ) {
    return ` This is based on portfolio data last updated ${syncAgeMinutes(variant)} minutes ago.`
  }

  return ''
}

function visiblePositions(portfolio: Portfolio): WalletPosition[] {
  return portfolio.positions
    .filter(
      (position) =>
        position.visibility === AssetVisibility.VISIBLE &&
        position.value_usd !== null,
    )
    .sort((a, b) => Number(b.value_usd) - Number(a.value_usd))
}

function assetReference(position: WalletPosition): AIReference {
  return {
    type: AIReferenceType.ASSET,
    id: position.asset.id,
    label: position.asset.symbol,
  }
}

function claim(text: string, evidence: AIClaim['evidence']): AIClaim {
  return { text, evidence }
}

function shortHash(hash: string): string {
  return `${hash.slice(0, 6)}…${hash.slice(-4)}`
}

const usdFormatter = new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'USD',
  maximumFractionDigits: 2,
})

function usd(value: string | null): string {
  if (value === null) return 'an unavailable amount'
  return usdFormatter.format(Number(value))
}

function pct(value: string | null): string {
  if (value === null) return 'an unknown share'
  return `${Math.abs(Number(value)).toFixed(2)}%`
}

function direction(value: string | null): string {
  if (value === null) return 'unchanged by'
  return Number(value) < 0 ? 'down' : 'up'
}
