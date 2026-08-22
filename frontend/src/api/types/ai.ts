/**
 * AI contracts.
 *
 * An AI answer is a structured object, never a bare string. Claims carry
 * evidence pointing at backend calculations, and references point at domain
 * entities so answers can become interactive without a contract change.
 *
 * PROVISIONAL CONTRACT LAYER — see `src/api/types/index.ts`.
 */
import type { ApiErrorBody } from './errors'
import type {
  AIEvidenceType,
  AIIntent,
  AIReferenceType,
  AIToolCallStatus,
  AIToolName,
  DataQuality,
  MessageRole,
  MessageStatus,
  SubscriptionPlan,
} from './enums'
import type { CursorParams, DateOnly, Timestamp } from './primitives'

/** Backend fact a claim is anchored to. */
export interface AIEvidence {
  type: AIEvidenceType | (string & {})
  id: string
}

/** A factual statement inside an answer, with its supporting evidence. */
export interface AIClaim {
  text: string
  evidence: AIEvidence[]
}

/** A domain entity the answer talks about. */
export interface AIReference {
  type: AIReferenceType | (string & {})
  id: string
  /** Display label resolved by the backend, e.g. an asset symbol. */
  label: string | null
}

/**
 * The AI response contract.
 *
 * `data_quality` propagates the quality of the underlying facts — an answer
 * built on PARTIAL data must not be presented as exact.
 */
export interface AIResponse {
  answer: string
  intent: AIIntent
  data_quality: DataQuality
  claims: AIClaim[]
  references: AIReference[]
  /**
   * Set when `intent` is `UNSUPPORTED`: a domain reason the capability is out
   * of scope, so the UI can explain instead of pretending.
   */
  unsupported_reason: string | null
}

export interface AIToolCall {
  id: string
  tool: AIToolName | (string & {})
  status: AIToolCallStatus
  started_at: Timestamp
  completed_at: Timestamp | null
}

export interface Conversation {
  id: string
  wallet_id: string
  title: string
  message_count: number
  last_message_preview: string | null
  created_at: Timestamp
  updated_at: Timestamp
}

export interface ConversationMessage {
  id: string
  conversation_id: string
  role: MessageRole
  status: MessageStatus
  /** Raw user text, or the assistant's answer text. */
  content: string
  /** Structured payload; present on completed assistant messages. */
  response: AIResponse | null
  tool_calls: AIToolCall[]
  error: ApiErrorBody | null
  created_at: Timestamp
}

export interface CreateConversationRequest {
  wallet_id: string
  /** Optional first question, so the UI can create-and-ask in one step. */
  title?: string | null
}

export interface SendMessageRequest {
  content: string
  /**
   * Optional entity the question is about (e.g. a transaction for the
   * explainer). The backend decides which tools this implies.
   */
  context?: {
    transaction_id?: string
    scenario_id?: string
  }
}

/* -------------------------------------------------------------------------- */
/* SSE stream                                                                 */
/* -------------------------------------------------------------------------- */

/**
 * Events emitted by `POST /ai/conversations/:id/messages`.
 *
 * Unknown event types are ignored by the stream reader, so the backend can add
 * events without breaking clients.
 */
export type AIStreamEvent =
  | AIStreamToolStartedEvent
  | AIStreamToolCompletedEvent
  | AIStreamTextDeltaEvent
  | AIStreamCompletedEvent
  | AIStreamErrorEvent

export interface AIStreamToolStartedEvent {
  type: 'tool_started'
  tool_call_id: string
  tool: AIToolName | (string & {})
}

export interface AIStreamToolCompletedEvent {
  type: 'tool_completed'
  tool_call_id: string
  tool: AIToolName | (string & {})
  ok: boolean
}

export interface AIStreamTextDeltaEvent {
  type: 'text_delta'
  text: string
}

/** Terminal success event: carries the persisted, structured message. */
export interface AIStreamCompletedEvent {
  type: 'completed'
  message: ConversationMessage
  usage: AIUsage | null
}

/** Terminal failure event carrying a domain error. */
export interface AIStreamErrorEvent {
  type: 'error'
  error: ApiErrorBody
}

/* -------------------------------------------------------------------------- */
/* Usage                                                                      */
/* -------------------------------------------------------------------------- */

/**
 * Daily AI operation budget. Enforced by the backend; the frontend only
 * displays it.
 */
export interface AIUsage {
  date: DateOnly
  used: number
  limit: number
  remaining: number
  resets_at: Timestamp
  plan: SubscriptionPlan
}

export type ConversationListParams = CursorParams & { wallet_id?: string }
export type MessageListParams = CursorParams
