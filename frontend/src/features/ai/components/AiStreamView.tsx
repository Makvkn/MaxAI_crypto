import { AIToolCallStatus } from '@/api/types'
import { Check, Warning } from '@/components/ui/Icon'
import { Spinner } from '@/components/ui/Spinner'
import { aiToolLabel } from '@/lib/copy/labels'
import { errorBodyCopy } from '@/lib/errors/messages'
import { cn } from '@/lib/utils/cn'
import type { AiStreamState, AiToolActivity } from '../streaming/aiStreamTypes'
import { AiAnswer, AnswerText } from './AiAnswer'

/**
 * The in-flight answer.
 *
 * While streaming, the accumulated text is shown as it arrives. Once the
 * backend confirms the persisted message, the structured response replaces it —
 * so claims, references and data quality come from the contract rather than
 * being parsed out of prose.
 */
export function AiStreamView({ state }: { state: AiStreamState }) {
  if (state.status === 'idle') return null

  return (
    <div className="space-y-3">
      {state.tools.length > 0 ? <ToolTrail tools={state.tools} /> : null}

      {state.response ? (
        <AiAnswer response={state.response} />
      ) : state.text ? (
        <div>
          <AnswerText text={state.text} />
          {state.status === 'streaming' ? (
            <span
              className="ml-0.5 inline-block h-4 w-[2px] animate-caret bg-accent align-middle"
              aria-hidden="true"
            />
          ) : null}
        </div>
      ) : state.status === 'streaming' ? (
        <p className="flex items-center gap-2 text-[13px] text-fg-subtle">
          <Spinner className="size-3.5 text-accent" />
          Thinking
        </p>
      ) : null}

      {state.status === 'cancelled' ? (
        <p className="text-[12px] text-fg-subtle">
          Stopped. The partial answer above is kept.
        </p>
      ) : null}

      {state.error ? <StreamError state={state} /> : null}

      <span className="sr-only" aria-live="polite">
        {state.status === 'streaming'
          ? 'Answer in progress'
          : state.status === 'completed'
            ? 'Answer complete'
            : ''}
      </span>
    </div>
  )
}

function StreamError({ state }: { state: AiStreamState }) {
  const copy = errorBodyCopy(state.error)

  return (
    <div
      role="alert"
      className="flex items-start gap-2.5 rounded-lg border border-caution/25 bg-caution-quiet/40 px-3 py-2.5"
    >
      <Warning className="mt-0.5 size-4 shrink-0 text-caution" />
      <div>
        <p className="text-[13px] font-medium text-fg">{copy.title}</p>
        <p className="mt-0.5 text-[12px] leading-relaxed text-fg-subtle">
          {copy.description}
        </p>
      </div>
    </div>
  )
}

/** Which backend tools the orchestrator ran, and how they finished. */
export function ToolTrail({
  tools,
  className,
}: {
  tools: AiToolActivity[]
  className?: string
}) {
  return (
    <ul className={cn('flex flex-wrap gap-1.5', className)}>
      {tools.map((tool) => (
        <li
          key={tool.id}
          className={cn(
            'inline-flex items-center gap-1.5 rounded-md border px-2 py-1 text-[11px]',
            tool.status === AIToolCallStatus.FAILED
              ? 'border-caution/30 bg-caution-quiet/40 text-caution'
              : 'border-line-strong bg-surface-raised text-fg-subtle',
          )}
        >
          {tool.status === AIToolCallStatus.RUNNING ? (
            <Spinner className="size-3 text-accent" />
          ) : tool.status === AIToolCallStatus.COMPLETED ? (
            <Check className="size-3 text-positive" />
          ) : (
            <Warning className="size-3" />
          )}
          {aiToolLabel(tool.tool)}
        </li>
      ))}
    </ul>
  )
}
