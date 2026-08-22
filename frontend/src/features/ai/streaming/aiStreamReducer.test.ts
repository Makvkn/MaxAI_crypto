import { describe, expect, it } from 'vitest'
import {
  AIIntent,
  AIToolCallStatus,
  DataQuality,
  MessageRole,
  MessageStatus,
  type AIStreamEvent,
  type ConversationMessage,
} from '@/api/types'
import { aiStreamReducer } from './aiStreamReducer'
import { initialAiStreamState, type AiStreamState } from './aiStreamTypes'

function reduce(events: AIStreamEvent[], from?: AiStreamState): AiStreamState {
  return events.reduce<AiStreamState>(
    (state, event) => aiStreamReducer(state, { type: 'event', event }),
    from ?? aiStreamReducer(initialAiStreamState, { type: 'start' }),
  )
}

const completedMessage: ConversationMessage = {
  id: 'msg_1',
  conversation_id: 'cnv_1',
  role: MessageRole.ASSISTANT,
  status: MessageStatus.COMPLETED,
  content: 'Your portfolio is down 4.21%.',
  response: {
    answer: 'Your portfolio is down 4.21%.',
    intent: AIIntent.PORTFOLIO_PERFORMANCE,
    data_quality: DataQuality.COMPLETE,
    claims: [{ text: 'ETH drove the decline.', evidence: [] }],
    references: [],
    unsupported_reason: null,
  },
  tool_calls: [],
  error: null,
  created_at: new Date().toISOString(),
}

describe('aiStreamReducer', () => {
  it('accumulates text deltas without losing order', () => {
    const state = reduce([
      { type: 'text_delta', text: 'Your portfolio ' },
      { type: 'text_delta', text: 'is down ' },
      { type: 'text_delta', text: '4.21%.' },
    ])

    expect(state.status).toBe('streaming')
    expect(state.text).toBe('Your portfolio is down 4.21%.')
  })

  it('tracks tool activity through to completion', () => {
    const state = reduce([
      { type: 'tool_started', tool_call_id: 'tc_1', tool: 'get_portfolio' },
      {
        type: 'tool_completed',
        tool_call_id: 'tc_1',
        tool: 'get_portfolio',
        ok: true,
      },
    ])

    expect(state.tools).toHaveLength(1)
    expect(state.tools[0]?.status).toBe(AIToolCallStatus.COMPLETED)
  })

  it('marks a failed tool without failing the whole answer', () => {
    const state = reduce([
      { type: 'tool_started', tool_call_id: 'tc_1', tool: 'get_asset_price' },
      {
        type: 'tool_completed',
        tool_call_id: 'tc_1',
        tool: 'get_asset_price',
        ok: false,
      },
    ])

    expect(state.tools[0]?.status).toBe(AIToolCallStatus.FAILED)
    expect(state.status).toBe('streaming')
  })

  it('adopts the structured response on completion', () => {
    const state = reduce([
      { type: 'text_delta', text: 'partial' },
      { type: 'completed', message: completedMessage, usage: null },
    ])

    expect(state.status).toBe('completed')
    expect(state.message?.id).toBe('msg_1')
    expect(state.response?.intent).toBe(AIIntent.PORTFOLIO_PERFORMANCE)
    expect(state.response?.claims).toHaveLength(1)
  })

  it('records a terminal domain error', () => {
    const state = reduce([
      {
        type: 'error',
        error: { code: 'AI_DAILY_LIMIT_REACHED', message: 'limit' },
      },
    ])

    expect(state.status).toBe('failed')
    expect(state.error?.code).toBe('AI_DAILY_LIMIT_REACHED')
  })

  it('keeps the partial answer when the user cancels', () => {
    const streaming = reduce([{ type: 'text_delta', text: 'Half an answer' }])
    const cancelled = aiStreamReducer(streaming, { type: 'cancelled' })

    expect(cancelled.status).toBe('cancelled')
    expect(cancelled.text).toBe('Half an answer')
  })

  it('resets back to idle', () => {
    const state = aiStreamReducer(
      reduce([{ type: 'text_delta', text: 'text' }]),
      { type: 'reset' },
    )

    expect(state).toEqual(initialAiStreamState)
  })
})
