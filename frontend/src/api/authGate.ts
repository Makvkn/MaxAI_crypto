/**
 * Coordinates auth bootstrap across React Query and HttpClient.
 *
 * Protected HTTP requests wait here until session validation (including any
 * startup refresh) completes. Reactive 401 refresh stays unchanged.
 */

let ready = false
let bootstrapping = false
const waiters = new Set<() => void>()

function flushWaiters() {
  for (const resolve of waiters) resolve()
  waiters.clear()
}

/** Allows authenticated requests again (guest / signed-out visitors). */
export function markAuthReady(): void {
  ready = true
  bootstrapping = false
  flushWaiters()
}

/** Blocks authenticated requests until bootstrap finishes. */
export function markAuthPending(): void {
  ready = false
}

/** True while initializeSession is running. */
export function isAuthBootstrapping(): boolean {
  return bootstrapping
}

export function isAuthReady(): boolean {
  return ready
}

export function waitForAuthReady(): Promise<void> {
  if (ready) return Promise.resolve()
  return new Promise((resolve) => {
    waiters.add(resolve)
  })
}

/** Runs startup session validation with the HTTP gate lifted. */
export async function runAuthBootstrap<T>(task: () => Promise<T>): Promise<T> {
  bootstrapping = true
  markAuthPending()
  try {
    return await task()
  } finally {
    bootstrapping = false
    markAuthReady()
  }
}

/** Called on logout so the next visitor starts from a clean gate. */
export function resetAuthGate(): void {
  ready = false
  bootstrapping = false
  waiters.clear()
}
