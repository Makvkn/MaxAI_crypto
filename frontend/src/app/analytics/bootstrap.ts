import { env } from '@/app/config/env'
import { analytics } from '@/lib/analytics/analytics'
import { createDebugProvider } from '@/lib/analytics/providers/debugProvider'

/**
 * Registers the analytics destination.
 *
 * The only place in the application that knows which provider is in use. A real
 * vendor is registered here without touching a single component.
 */
export function bootstrapAnalytics(): void {
  if (env.analyticsDebug || (env.analyticsEnabled && env.isDev)) {
    analytics.register(createDebugProvider())
  }
}
