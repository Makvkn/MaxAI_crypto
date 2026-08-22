import type {
  AIResponse,
  AIToolCallStatus,
  AIUsage,
  ApiErrorBody,
  ConversationMessage,
} from '@/api/types'

/**
 * UI-facing shape of an in-flight AI answer.
 *
 * The stream is modelled as state, not as a string being appended to a
 * component: tool activity, the streamed text, the final structured response
 * and terminal errors are all first-class.
 */
export type AiStreamStatus =
  | 'idle'
  | 'streaming'
  | 'completed'
  | 'failed'
  | 'cancelled'

export interface AiToolActivity {
  id: string
  tool: string
  status: AIToolCallStatus
}

export interface AiStreamState {
  status: AiStreamStatus
  /** Text accumulated from `text_delta` events. */
  text: string
  tools: AiToolActivity[]
  /** Persisted message, available once `completed` arrives. */
  message: ConversationMessage | null
  /** Structured payload: claims, references, data quality, intent. */
  response: AIResponse | null
  usage: AIUsage | null
  error: ApiErrorBody | null
}

export const initialAiStreamState: AiStreamState = {
  status: 'idle',
  text: '',
  tools: [],
  message: null,
  response: null,
  usage: null,
  error: null,
}

export type AiStreamAction =
  | { type: 'start' }
  | { type: 'event'; event: import('@/api/types').AIStreamEvent }
  | { type: 'cancelled' }
  | { type: 'reset' }
