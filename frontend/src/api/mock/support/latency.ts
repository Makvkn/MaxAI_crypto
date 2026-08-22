import { ApiError } from '../../errors'
import { ApiErrorCode } from '../../types'

/**
 * Simulated network latency.
 *
 * The mock is deliberately not instantaneous: loading, cancellation and
 * skeleton states must be exercised during development.
 */
export function delay(ms: number, signal?: AbortSignal): Promise<void> {
  return new Promise((resolve, reject) => {
    if (signal?.aborted) {
      reject(cancelled())
      return
    }

    const timer = setTimeout(() => {
      signal?.removeEventListener('abort', onAbort)
      resolve()
    }, ms)

    function onAbort() {
      clearTimeout(timer)
      reject(cancelled())
    }

    signal?.addEventListener('abort', onAbort, { once: true })
  })
}

function cancelled(): ApiError {
  return new ApiError({
    code: ApiErrorCode.CANCELLED,
    message: 'Request cancelled.',
  })
}
