import type { RequestOptions } from '../client'
import { isCancelledError, normalizeUnknownError, toErrorBody } from '../errors'
import { http } from '../http'
import { parseSseData, readSseFrames, type SseFrame } from '../sse'
import type {
  AIStreamEvent,
  Conversation,
  ConversationIdPath,
  ConversationListParams,
  ConversationMessage,
  CreateConversationRequest,
  CursorPage,
  MessageListParams,
  SendMessageRequest,
} from '../types'

/**
 * `/api/v1/ai/conversations`
 *
 * Sending a message opens an SSE stream. This module converts SSE frames into
 * typed domain events; unknown event types are skipped so the backend can add
 * events without breaking the client.
 */

export const apiGetConversations = (
  params?: ConversationListParams,
  options?: RequestOptions,
): Promise<CursorPage<Conversation>> =>
  http.get<CursorPage<Conversation>>(
    '/ai/conversations',
    {
      wallet_id: params?.wallet_id,
      limit: params?.limit,
      cursor: params?.cursor,
    },
    options,
  )

export const apiCreateConversation = (
  request: CreateConversationRequest,
  options?: RequestOptions,
): Promise<Conversation> =>
  http.post<Conversation>('/ai/conversations', request, options)

export const apiGetConversationMessages = (
  { conversationId }: ConversationIdPath,
  params?: MessageListParams,
  options?: RequestOptions,
): Promise<CursorPage<ConversationMessage>> =>
  http.get<CursorPage<ConversationMessage>>(
    `/ai/conversations/${encodeURIComponent(conversationId)}/messages`,
    { limit: params?.limit, cursor: params?.cursor },
    options,
  )

export function apiStreamConversationMessage(
  { conversationId }: ConversationIdPath,
  request: SendMessageRequest,
  options?: { signal?: AbortSignal },
): AsyncGenerator<AIStreamEvent> {
  return readConversationMessageStream(conversationId, request, options)
}

async function* readConversationMessageStream(
  conversationId: string,
  request: SendMessageRequest,
  options?: { signal?: AbortSignal },
): AsyncGenerator<AIStreamEvent> {
  try {
    const response = await http.openStream(
      `/ai/conversations/${encodeURIComponent(conversationId)}/messages`,
      request,
      { signal: options?.signal },
    )

    for await (const frame of readSseFrames(response)) {
      const event = parseConversationStreamEvent(frame)
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
function parseConversationStreamEvent(frame: SseFrame): AIStreamEvent | null {
  const payload = parseSseData<Record<string, unknown>>(frame)
  if (!payload) return null

  const type =
    typeof payload.type === 'string' ? payload.type : (frame.event ?? '')

  if (!STREAM_EVENT_TYPES.has(type as AIStreamEvent['type'])) return null
  return { ...payload, type } as AIStreamEvent
}
