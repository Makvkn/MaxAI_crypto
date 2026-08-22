import { AIToolCallStatus } from '@/api/types'
import {
  initialAiStreamState,
  type AiStreamAction,
  type AiStreamState,
} from './aiStreamTypes'

/**
 * Pure reduction of AI stream events into renderable state.
 *
 * Keeping this a pure function means streaming behaviour — deltas, tool
 * activity, completion, failure, cancellation — is unit-testable without a
 * network or a component.
 */
export function aiStreamReducer(
  state: AiStreamState,
  action: AiStreamAction,
): AiStreamState {
  switch (action.type) {
    case 'start':
      return { ...initialAiStreamState, status: 'streaming' }

    case 'reset':
      return initialAiStreamState

    case 'cancelled':
      // The partial text is kept: the user asked to stop, not to erase.
      return state.status === 'streaming'
        ? { ...state, status: 'cancelled' }
        : state

    case 'event': {
      const event = action.event

      switch (event.type) {
        case 'tool_started':
          return {
            ...state,
            status: 'streaming',
            tools: [
              ...state.tools,
              {
                id: event.tool_call_id,
                tool: event.tool,
                status: AIToolCallStatus.RUNNING,
              },
            ],
          }

        case 'tool_completed':
          return {
            ...state,
            tools: state.tools.map((tool) =>
              tool.id === event.tool_call_id
                ? {
                    ...tool,
                    status: event.ok
                      ? AIToolCallStatus.COMPLETED
                      : AIToolCallStatus.FAILED,
                  }
                : tool,
            ),
          }

        case 'text_delta':
          return { ...state, status: 'streaming', text: state.text + event.text }

        case 'completed':
          return {
            ...state,
            status: 'completed',
            message: event.message,
            response: event.message.response,
            // The persisted answer is authoritative over accumulated deltas.
            text: event.message.response?.answer ?? state.text,
            usage: event.usage,
            tools: state.tools.map((tool) =>
              tool.status === AIToolCallStatus.RUNNING
                ? { ...tool, status: AIToolCallStatus.COMPLETED }
                : tool,
            ),
          }

        case 'error':
          return { ...state, status: 'failed', error: event.error }

        default:
          return state
      }
    }

    default:
      return state
  }
}
