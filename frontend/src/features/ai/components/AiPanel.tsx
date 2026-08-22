import { useEffect, useRef, useState, type FormEvent } from 'react'
import type { SendMessageRequest } from '@/api/types'
import { Button } from '@/components/ui/Button'
import { IconButton } from '@/components/ui/IconButton'
import { Menu } from '@/components/ui/Menu'
import { Send, Sparkle, Stop } from '@/components/ui/Icon'
import { SkeletonRows } from '@/components/ui/Skeleton'
import { ErrorState } from '@/components/feedback/States'
import { AiMessage } from './AiMessage'
import { AiStreamView } from './AiStreamView'
import { AiUsageMeter } from './AiUsageMeter'
import {
  useAiStream,
  useAiUsage,
  useConversationMessages,
  useConversations,
} from '../hooks/useAiConversation'
import { analytics } from '@/lib/analytics/analytics'
import { formatTime } from '@/lib/dates/format'
import { cn } from '@/lib/utils/cn'

/**
 * Ask AI.
 *
 * Conversation history is server state; only the in-flight stream is local. The
 * panel sends a question and renders whatever structured answer comes back — it
 * has no opinion about the portfolio itself.
 */
export function AiPanel({
  walletId,
  suggestions,
  enabled = true,
  className,
}: {
  walletId: string
  suggestions: string[]
  enabled?: boolean
  className?: string
}) {
  const [conversationId, setConversationId] = useState<string | null>(null)
  const [draft, setDraft] = useState('')
  const scrollRef = useRef<HTMLDivElement>(null)

  const conversations = useConversations(walletId, enabled)
  const usage = useAiUsage(enabled)
  const history = useConversationMessages(conversationId, enabled)
  const { state, send, cancel, reset, isSending } = useAiStream({
    walletId,
    conversationId,
    onConversation: setConversationId,
  })

  const messages = history.messages
  const streamedId = state.message?.id ?? null

  // Once the persisted message is in the cache, the local stream copy is
  // redundant — dropping it avoids showing the same answer twice.
  useEffect(() => {
    if (!streamedId) return
    if (messages.some((message) => message.id === streamedId)) reset()
  }, [messages, reset, streamedId])

  useEffect(() => {
    const node = scrollRef.current
    if (!node) return
    node.scrollTop = node.scrollHeight
  }, [messages.length, state.text, state.status])

  const completedIntent =
    state.status === 'completed' ? (state.response?.intent ?? null) : null

  useEffect(() => {
    if (!completedIntent) return
    analytics.track('ai_insight_viewed', {
      wallet_id: walletId,
      intent: completedIntent,
    })
  }, [completedIntent, walletId])

  const limitReached = usage.data ? usage.data.remaining <= 0 : false
  const canSend = draft.trim() !== '' && !isSending && !limitReached

  const ask = (content: string) => {
    const request: SendMessageRequest = { content }
    setDraft('')
    void send(request)
  }

  const onSubmit = (event: FormEvent) => {
    event.preventDefault()
    if (!canSend) return
    ask(draft.trim())
  }

  const isEmpty =
    messages.length === 0 && state.status === 'idle' && !history.isLoading

  return (
    <section
      aria-label="Ask AI"
      className={cn(
        'flex min-h-0 flex-col rounded-card border border-line bg-surface',
        className,
      )}
    >
      <header className="flex items-center justify-between gap-3 border-b border-line px-5 py-4">
        <div className="flex items-center gap-2">
          <Sparkle className="size-4 text-accent" />
          <h2 className="text-[13px] font-medium tracking-[0.06em] text-fg-muted uppercase">
            Ask AI
          </h2>
        </div>

        <div className="flex items-center gap-2">
          {conversations.items.length > 0 ? (
            <Menu
              label="Conversations"
              align="end"
              trigger={(triggerProps) => (
                <button
                  type="button"
                  {...triggerProps}
                  className="rounded-md px-2 py-1 text-[12px] text-fg-subtle transition-colors hover:text-fg"
                >
                  History
                </button>
              )}
              items={[
                {
                  id: 'new',
                  label: 'New conversation',
                  onSelect: () => {
                    setConversationId(null)
                    reset()
                  },
                },
                ...conversations.items.map((conversation) => ({
                  id: conversation.id,
                  label: conversation.title,
                  description: formatTime(conversation.updated_at),
                  onSelect: () => {
                    setConversationId(conversation.id)
                    reset()
                  },
                })),
              ]}
            />
          ) : null}
        </div>
      </header>

      <div
        ref={scrollRef}
        className="min-h-0 flex-1 space-y-5 overflow-y-auto px-5 py-4"
      >
        {history.isLoading ? (
          <SkeletonRows rows={3} />
        ) : history.error ? (
          <ErrorState error={history.error} onRetry={history.refetch} compact />
        ) : null}

        {isEmpty ? (
          <Intro suggestions={suggestions} onAsk={ask} disabled={limitReached} />
        ) : null}

        {messages.map((message) => (
          <AiMessage key={message.id} message={message} />
        ))}

        {history.hasNextPage ? (
          <Button
            variant="quiet"
            size="sm"
            onClick={history.fetchNextPage}
            loading={history.isFetchingNextPage}
          >
            Load earlier messages
          </Button>
        ) : null}

        <AiStreamView state={state} />
      </div>

      <div className="border-t border-line px-5 py-3">
        {limitReached && usage.data ? (
          <p className="mb-2.5 text-[12px] leading-relaxed text-caution">
            You've reached your daily AI limit. It resets at{' '}
            {formatTime(usage.data.resets_at)} — the dashboard stays fully
            available.
          </p>
        ) : null}

        <form onSubmit={onSubmit} className="flex items-end gap-2">
          <label className="sr-only" htmlFor="ai-question">
            Ask about your portfolio
          </label>
          <textarea
            id="ai-question"
            value={draft}
            onChange={(event) => setDraft(event.target.value)}
            onKeyDown={(event) => {
              if (event.key === 'Enter' && !event.shiftKey) {
                event.preventDefault()
                if (canSend) ask(draft.trim())
              }
            }}
            rows={1}
            placeholder={
              limitReached
                ? 'Daily AI limit reached'
                : 'Ask about your portfolio…'
            }
            disabled={limitReached || !enabled}
            className="max-h-32 min-h-[42px] w-full resize-none rounded-lg border border-line-strong bg-base-elevated px-3.5 py-2.5 text-[13px] text-fg placeholder:text-fg-subtle/70 focus:outline-none focus-visible:border-accent focus-visible:ring-2 focus-visible:ring-accent/25 disabled:opacity-60"
          />

          {isSending ? (
            <IconButton
              label="Stop generating"
              onClick={cancel}
              className="border-line-strong bg-base-elevated"
            >
              <Stop className="size-4" />
            </IconButton>
          ) : (
            <IconButton
              label="Send question"
              type="submit"
              disabled={!canSend}
              className="bg-accent text-accent-fg hover:bg-accent-hover hover:text-accent-fg"
            >
              <Send className="size-4" />
            </IconButton>
          )}
        </form>

        {usage.data ? (
          <div className="mt-2.5">
            <AiUsageMeter usage={usage.data} />
          </div>
        ) : null}
      </div>
    </section>
  )
}

function Intro({
  suggestions,
  onAsk,
  disabled,
}: {
  suggestions: string[]
  onAsk: (question: string) => void
  disabled: boolean
}) {
  return (
    <div>
      <p className="text-[14px] leading-relaxed text-fg-muted">
        Ask what your portfolio is doing and why. Answers are built from your
        actual positions, snapshots and transactions — not from market
        commentary.
      </p>

      <ul className="mt-4 space-y-2">
        {suggestions.map((suggestion) => (
          <li key={suggestion}>
            <button
              type="button"
              onClick={() => onAsk(suggestion)}
              disabled={disabled}
              className="w-full rounded-lg border border-line bg-base-elevated px-3.5 py-2.5 text-left text-[13px] text-fg-muted transition-colors hover:border-line-strong hover:text-fg disabled:opacity-50"
            >
              {suggestion}
            </button>
          </li>
        ))}
      </ul>

      <p className="mt-4 text-[12px] leading-relaxed text-fg-subtle">
        MaxAI analyses, explains and simulates. It does not give buy or sell
        recommendations.
      </p>
    </div>
  )
}
