/**
 * Defensive HTTP barrier synced from AuthStatus.
 *
 * Protected requests wait here until auth bootstrap completes. This is not the
 * primary application lifecycle — SessionProvider owns AuthStatus and calls
 * syncAuthGate on every transition.
 */

export type GateAuthStatus =
  | 'bootstrapping'
  | 'authenticated'
  | 'unauthenticated'

let ready = false
let bootstrapping = false
const waiters = new Set<() => void>()

function flushWaiters(): void {
  for (const resolve of waiters) resolve()
  waiters.clear()
}

function closeGate(): void {
  ready = false
}

function openGate(): void {
  ready = true
  flushWaiters()
}

/** Keeps the HTTP gate aligned with the auth lifecycle state machine. */
export function syncAuthGate(status: GateAuthStatus): void {
  switch (status) {
    case 'bootstrapping':
      closeGate()
      break
    case 'authenticated':
    case 'unauthenticated':
      openGate()
      break
  }
}

/** True while initializeSession runs inside runAuthBootstrap. */
export function isAuthBootstrapping(): boolean {
  return bootstrapping
}

/** @internal Tests and diagnostics only — prefer SessionContext.authReady. */
export function isAuthGateOpen(): boolean {
  return ready
}

export function waitForAuthReady(): Promise<void> {
  if (ready) return Promise.resolve()
  return new Promise((resolve) => {
    waiters.add(resolve)
  })
}

/**
 * Lifts the HTTP gate for internal bootstrap requests so initializeSession
 * cannot deadlock on waitForAuthReady.
 */
export async function runAuthBootstrap<T>(task: () => Promise<T>): Promise<T> {
  bootstrapping = true
  try {
    return await task()
  } finally {
    bootstrapping = false
  }
}

/** Clears waiters on logout before the next lifecycle begins. */
export function resetAuthGate(): void {
  ready = false
  bootstrapping = false
  waiters.clear()
}
