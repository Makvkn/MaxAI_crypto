/**
 * Guards against stale bootstrap callbacks completing after logout or adopt.
 *
 * Each bootstrap attempt captures the epoch at start; invalidate() bumps the
 * epoch synchronously so in-flight promises cannot overwrite AuthStatus.
 */
export function createBootstrapEpoch() {
  let epoch = 0

  return {
    start(): number {
      epoch += 1
      return epoch
    },
    invalidate(): void {
      epoch += 1
    },
    isCurrent(attempt: number): boolean {
      return attempt === epoch
    },
  }
}

export type BootstrapEpoch = ReturnType<typeof createBootstrapEpoch>
