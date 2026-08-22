import { env } from '@/app/config/env'
import type { AnalyticsEventMap, AnalyticsEventName } from './events'

/**
 * Vendor-agnostic analytics.
 *
 * Components call `analytics.track(...)`; no component imports a vendor SDK.
 * Swapping the destination is a matter of registering a different provider at
 * bootstrap. Events emitted before a provider is registered are buffered.
 */
export interface AnalyticsProvider {
  readonly name: string
  track<Name extends AnalyticsEventName>(
    event: Name,
    properties: AnalyticsEventMap[Name],
  ): void
  identify(userId: string, traits?: Record<string, unknown>): void
  page(path: string): void
}

interface QueuedEvent {
  event: AnalyticsEventName
  properties: unknown
}

const MAX_QUEUE = 50

function createAnalytics() {
  let provider: AnalyticsProvider | null = null
  const queue: QueuedEvent[] = []
  /** Guards once-per-session events such as `first_ai_question`. */
  const fired = new Set<string>()

  return {
    register(next: AnalyticsProvider | null): void {
      provider = next
      if (!provider) return
      for (const queued of queue.splice(0)) {
        provider.track(
          queued.event as AnalyticsEventName,
          queued.properties as never,
        )
      }
    },

    track<Name extends AnalyticsEventName>(
      event: Name,
      properties: AnalyticsEventMap[Name] = {} as AnalyticsEventMap[Name],
    ): void {
      if (!env.analyticsEnabled && !env.analyticsDebug) return

      if (!provider) {
        if (queue.length < MAX_QUEUE) queue.push({ event, properties })
        return
      }
      provider.track(event, properties)
    },

    /** Tracks an event at most once per browser session. */
    trackOnce<Name extends AnalyticsEventName>(
      key: string,
      event: Name,
      properties: AnalyticsEventMap[Name] = {} as AnalyticsEventMap[Name],
    ): void {
      if (fired.has(key)) return
      fired.add(key)
      this.track(event, properties)
    },

    identify(userId: string, traits?: Record<string, unknown>): void {
      provider?.identify(userId, traits)
    },

    page(path: string): void {
      provider?.page(path)
    },
  }
}

export const analytics = createAnalytics()
