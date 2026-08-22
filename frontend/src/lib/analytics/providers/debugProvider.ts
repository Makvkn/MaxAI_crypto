import type { AnalyticsProvider } from '../analytics'

/**
 * Development sink: prints the event schema to the console so the funnel can be
 * verified without wiring a vendor. Production builds register a real provider
 * here instead.
 */
export function createDebugProvider(): AnalyticsProvider {
  return {
    name: 'debug',
    track(event, properties) {
      console.info(`[analytics] ${event}`, properties)
    },
    identify(userId, traits) {
      console.info('[analytics] identify', userId, traits ?? {})
    },
    page(path) {
      console.info('[analytics] page', path)
    },
  }
}
