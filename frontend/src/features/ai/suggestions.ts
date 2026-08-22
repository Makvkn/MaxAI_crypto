import { PerformancePeriod, type Portfolio } from '@/api/types'
import { deltaDirection } from '@/lib/formatting/percent'
import { periodDescriptions } from '@/lib/copy/labels'

/**
 * Starter questions.
 *
 * These are copy, not analysis: the only thing read from the portfolio is the
 * sign of a backend-provided change, so the wording matches what the user is
 * looking at. No figure is derived here.
 */
export function suggestedQuestions(
  portfolio: Portfolio | undefined,
  period: PerformancePeriod,
): string[] {
  const direction = deltaDirection(portfolio?.change_24h_pct)
  const window = periodDescriptions[period]

  const leading =
    direction === 'down'
      ? 'Why is my portfolio down?'
      : direction === 'up'
        ? 'Why is my portfolio up?'
        : 'What is my portfolio doing?'

  return [
    leading,
    'Summarise my portfolio',
    'How is my allocation distributed?',
    `What drove my performance over ${window}?`,
  ]
}
