import { useCallback, useEffect, useReducer, useRef, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  apiCreateConversation,
  apiGetAiUsage,
  apiGetConversationMessages,
  apiGetConversations,
} from '@/api'
import type {
  AIUsage,
  Conversation,
  ConversationMessage,
  SendMessageRequest,
} from '@/api/types'
import { AIIntent } from '@/api/types'
import { analytics } from '@/lib/analytics/analytics'
import { queryKeys } from '@/lib/query/queryKeys'
import { useCursorInfiniteQuery } from '@/lib/query/useCursorInfiniteQuery'
import { useProtectedQueryEnabled } from '@/features/auth/useProtectedQueryEnabled'
import { runAiStream } from '../streaming/aiStreamClient'
import { aiStreamReducer } from '../streaming/aiStreamReducer'
import {
  initialAiStreamState,
  type AiStreamState,
} from '../streaming/aiStreamTypes'

/**
 * AI conversation server state and streaming.
 *
 * Conversations and their messages are server state (TanStack Query). Only the
 * in-flight stream lives in local reducer state, and it is folded into the
 * query cache once the backend confirms the persisted message — the whole
 * conversation is never refetched per token.
 */

export function useConversations(walletId: string, enabled = true) {
  const protectedEnabled = useProtectedQueryEnabled(enabled)

  return useCursorInfiniteQuery<Conversation>({
    queryKey: queryKeys.conversations(walletId),
    enabled: protectedEnabled,
    fetchPage: ({ cursor, signal }) =>
      apiGetConversations({ wallet_id: walletId, cursor, limit: 20 }, { signal }),
  })
}

/** Messages arrive newest-first; the UI renders them chronologically. */
export function useConversationMessages(
  conversationId: string | null,
  enabled = true,
) {
  const protectedEnabled = useProtectedQueryEnabled(
    Boolean(conversationId) && enabled,
  )

  const query = useCursorInfiniteQuery<ConversationMessage>({
    queryKey: queryKeys.messages(conversationId ?? 'none'),
    enabled: protectedEnabled,
    staleTime: 10_000,
    fetchPage: ({ cursor, signal }) =>
      apiGetConversationMessages(
        { conversationId: conversationId as string },
        { cursor, limit: 30 },
        { signal },
      ),
  })

  return {
    ...query,
    /** Oldest first, for display. */
    messages: [...query.items].reverse(),
  }
}

export function useCreateConversation(walletId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (title?: string) =>
      apiCreateConversation({ wallet_id: walletId, title }),
    onSuccess: () => {
      void queryClient.invalidateQueries({
        queryKey: queryKeys.conversations(walletId),
      })
    },
  })
}

export function useAiUsage(enabled = true) {
  const protectedEnabled = useProtectedQueryEnabled(enabled)

  return useQuery({
    queryKey: queryKeys.aiUsage(),
    queryFn: ({ signal }) => apiGetAiUsage({ signal }),
    enabled: protectedEnabled,
    staleTime: 30_000,
  })
}

interface SendOptions {
  walletId: string
  conversationId: string | null
  /** Called with the conversation used, so callers can adopt a new one. */
  onConversation?: (conversationId: string) => void
}

/**
 * Sends a message and exposes the stream as state.
 *
 * Cancellation aborts the request; the partial answer stays on screen because
 * the user asked to stop rather than to undo.
 */
export function useAiStream({
  walletId,
  conversationId,
  onConversation,
}: SendOptions) {
  const queryClient = useQueryClient()
  const [state, dispatch] = useReducer(aiStreamReducer, initialAiStreamState)
  const [isSending, setIsSending] = useState(false)
  const abortRef = useRef<AbortController | null>(null)
  const questionCountRef = useRef(0)

  useEffect(() => () => abortRef.current?.abort(), [])

  const cancel = useCallback(() => {
    abortRef.current?.abort()
    abortRef.current = null
    dispatch({ type: 'cancelled' })
    setIsSending(false)
  }, [])

  const reset = useCallback(() => dispatch({ type: 'reset' }), [])

  const send = useCallback(
    async (request: SendMessageRequest) => {
      abortRef.current?.abort()
      const controller = new AbortController()
      abortRef.current = controller

      setIsSending(true)
      dispatch({ type: 'start' })

      let targetConversationId = conversationId
      let usage: AIUsage | null = null

      try {
        if (!targetConversationId) {
          const conversation = await apiCreateConversation({
            wallet_id: walletId,
            title: request.content.slice(0, 60),
          })
          targetConversationId = conversation.id
          onConversation?.(conversation.id)
          void queryClient.invalidateQueries({
            queryKey: queryKeys.conversations(walletId),
          })
        }

        questionCountRef.current += 1
        const questionIndex = questionCountRef.current
        analytics.track('ai_question_asked', {
          wallet_id: walletId,
          question_index: questionIndex,
        })
        if (questionIndex === 1) {
          analytics.trackOnce(`first_ai_question:${walletId}`, 'first_ai_question', {
            wallet_id: walletId,
          })
        } else if (questionIndex === 2) {
          analytics.trackOnce(
            `second_ai_question:${walletId}`,
            'second_ai_question',
            { wallet_id: walletId },
          )
        }

        await runAiStream({
          conversationId: targetConversationId,
          request,
          signal: controller.signal,
          onEvent: (event) => {
            dispatch({ type: 'event', event })

            if (event.type === 'completed') {
              usage = event.usage
              if (event.message.response?.intent === AIIntent.UNSUPPORTED) {
                analytics.track('ai_answer_unsupported', {
                  wallet_id: walletId,
                  reason: event.message.response.unsupported_reason,
                })
              }
            }
            if (
              event.type === 'error' &&
              event.error.code === 'AI_DAILY_LIMIT_REACHED'
            ) {
              analytics.track('ai_limit_reached', {
                limit: Number(event.error.details?.limit ?? 0),
              })
            }
          },
        })
      } finally {
        setIsSending(false)
        abortRef.current = null

        if (targetConversationId) {
          // Reconcile once, at the end of the stream — not per token.
          void queryClient.invalidateQueries({
            queryKey: queryKeys.messages(targetConversationId),
          })
          void queryClient.invalidateQueries({
            queryKey: queryKeys.conversations(walletId),
          })
        }
        if (usage) {
          queryClient.setQueryData(queryKeys.aiUsage(), usage)
        } else {
          void queryClient.invalidateQueries({ queryKey: queryKeys.aiUsage() })
        }
      }
    },
    [conversationId, onConversation, queryClient, walletId],
  )

  return { state: state as AiStreamState, send, cancel, reset, isSending }
}
