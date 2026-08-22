import type { HttpClient } from '../client'
import type { ConversationsApi } from '../contract'
import { isCancelledError, normalizeUnknownError, toErrorBody } from '../errors'
import { parseSseData, readSseFrames, type SseFrame } from '../sse'
import type {
  AIStreamEvent,
  Conversation,
  ConversationMessage,
  CursorPage,
} from '../types'

/**
 * `/api/v1/ai/conversations`
 *
 * Sending a message opens an SSE stream. This module converts SSE frames into
 * typed domain events; unknown event types are skipped so the backend can add
 * events without breaking the client.
 */
export function createConversationsApi(http: HttpClient): ConversationsApi {
  return {
    list: (params, options) =>
      http.get<CursorPage<Conversation>>(
        '/ai/conversations',
        {
          wallet_id: params?.wallet_id,
          limit: params?.limit,
          cursor: params?.cursor,
        },
        options,
      ),

    create: (request, options) =>
      http.post<Conversation>('/ai/conversations', request, options),

    listMessages: (conversationId, params, options) =>
      http.get<CursorPage<ConversationMessage>>(
        `/ai/conversations/${encodeURIComponent(conversationId)}/messages`,
        { limit: params?.limit, cursor: params?.cursor },
        options,
      ),

    streamMessage(conversationId, request, options) {
      return streamMessage(http, conversationId, request, options)
    },
  }
}

async function* streamMessage(
  http: HttpClient,
  conversationId: string,
  request: unknown,
  options?: { signal?: AbortSignal },
): AsyncGenerator<AIStreamEvent> {
  try {
    const response = await http.openStream(
      `/ai/conversations/${encodeURIComponent(conversationId)}/messages`,
      request,
      { signal: options?.signal },
    )

    for await (const frame of readSseFrames(response)) {
      const event = toStreamEvent(frame)
      if (!event) continue
      yield event
      if (event.type === 'completed' || event.type === 'error') return
    }
  } catch (error) {
    // A deliberate cancellation simply ends the iteration; anything else is
    // surfaced as a terminal domain error so the UI has one code path.
    if (isCancelledError(error) || options?.signal?.aborted) return
    yield { type: 'error', error: toErrorBody(normalizeUnknownError(error)) }
  }
}

const STREAM_EVENT_TYPES = new Set<AIStreamEvent['type']>([
  'tool_started',
  'tool_completed',
  'text_delta',
  'completed',
  'error',
])

/**
 * Accepts the event name either from the SSE `event:` field or from a `type`
 * property inside the JSON payload.
 */
function toStreamEvent(frame: SseFrame): AIStreamEvent | null {
  const payload = parseSseData<Record<string, unknown>>(frame)
  if (!payload) return null

  const type =
    typeof payload.type === 'string' ? payload.type : (frame.event ?? '')

  if (!STREAM_EVENT_TYPES.has(type as AIStreamEvent['type'])) return null
  return { ...payload, type } as AIStreamEvent
}
