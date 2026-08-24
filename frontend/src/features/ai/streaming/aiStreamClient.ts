import { apiStreamConversationMessage } from '@/api'
import { isCancelledError, normalizeUnknownError, toErrorBody } from '@/api/errors'
import type { AIStreamEvent, SendMessageRequest } from '@/api/types'

/**
 * Drives one AI message stream.
 *
 * The React layer never iterates the stream itself and never parses SSE: it
 * hands over a conversation, a request and an abort signal, and receives typed
 * domain events. Swapping SSE for another transport is an API-layer change.
 */
export interface AiStreamRunner {
  conversationId: string
  request: SendMessageRequest
  signal: AbortSignal
  onEvent: (event: AIStreamEvent) => void
}

export async function runAiStream({
  conversationId,
  request,
  signal,
  onEvent,
}: AiStreamRunner): Promise<void> {
  try {
    const stream = apiStreamConversationMessage({ conversationId }, request, {
      signal,
    })

    for await (const event of stream) {
      if (signal.aborted) return
      onEvent(event)
    }
  } catch (error) {
    if (signal.aborted || isCancelledError(error)) return
    onEvent({ type: 'error', error: toErrorBody(normalizeUnknownError(error)) })
  }
}
