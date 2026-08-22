import { ApiError, normalizeUnknownError } from './errors'
import { ApiErrorCode } from './types'

/**
 * Transport-level Server-Sent Events support.
 *
 * The AI endpoint streams over `POST`, which `EventSource` cannot do, so the
 * stream is read from the `fetch` body. This module knows nothing about AI: it
 * only turns bytes into SSE frames. Domain event typing lives in
 * `@/features/ai/streaming`.
 */

export interface SseFrame {
  /** SSE `event:` field. `null` when the frame carries no explicit type. */
  event: string | null
  /** Concatenated `data:` lines (newline-joined, per the spec). */
  data: string
  id: string | null
}

/**
 * Incremental SSE frame parser. Separated from the network so it can be tested
 * against partial chunks and both `\n` and `\r\n` line endings.
 */
export function createSseParser() {
  let buffer = ''

  function parseBlock(block: string): SseFrame | null {
    let event: string | null = null
    let id: string | null = null
    const data: string[] = []

    for (const rawLine of block.split('\n')) {
      const line = rawLine.endsWith('\r') ? rawLine.slice(0, -1) : rawLine
      if (line === '' || line.startsWith(':')) continue

      const colon = line.indexOf(':')
      const field = colon === -1 ? line : line.slice(0, colon)
      let value = colon === -1 ? '' : line.slice(colon + 1)
      if (value.startsWith(' ')) value = value.slice(1)

      switch (field) {
        case 'event':
          event = value
          break
        case 'data':
          data.push(value)
          break
        case 'id':
          id = value
          break
        default:
          // `retry` and unknown fields are irrelevant here.
          break
      }
    }

    if (data.length === 0 && event === null) return null
    return { event, data: data.join('\n'), id }
  }

  return {
    /** Feeds a decoded chunk and returns every complete frame it produced. */
    push(chunk: string): SseFrame[] {
      buffer += chunk.replace(/\r\n/g, '\n')
      const frames: SseFrame[] = []
      let boundary = buffer.indexOf('\n\n')

      while (boundary !== -1) {
        const block = buffer.slice(0, boundary)
        buffer = buffer.slice(boundary + 2)
        const frame = parseBlock(block)
        if (frame) frames.push(frame)
        boundary = buffer.indexOf('\n\n')
      }

      return frames
    },
    /** Emits a trailing frame if the stream ended without a blank line. */
    flush(): SseFrame[] {
      if (buffer.trim() === '') {
        buffer = ''
        return []
      }
      const frame = parseBlock(buffer)
      buffer = ''
      return frame ? [frame] : []
    },
  }
}

/** Reads an SSE response body as an async iterable of frames. */
export async function* readSseFrames(
  response: Response,
): AsyncGenerator<SseFrame> {
  if (!response.body) {
    throw new ApiError({
      code: ApiErrorCode.MALFORMED_RESPONSE,
      message: 'The AI stream returned an empty body.',
      status: response.status,
    })
  }

  const reader = response.body.getReader()
  const decoder = new TextDecoder()
  const parser = createSseParser()

  try {
    for (;;) {
      const { done, value } = await reader.read()
      if (done) break
      yield* parser.push(decoder.decode(value, { stream: true }))
    }
    yield* parser.flush()
  } catch (error) {
    throw normalizeUnknownError(error)
  } finally {
    reader.releaseLock()
  }
}

/** Parses a frame's `data` payload as JSON, tolerating the `[DONE]` sentinel. */
export function parseSseData<T>(frame: SseFrame): T | null {
  if (!frame.data || frame.data === '[DONE]') return null
  try {
    return JSON.parse(frame.data) as T
  } catch {
    return null
  }
}
