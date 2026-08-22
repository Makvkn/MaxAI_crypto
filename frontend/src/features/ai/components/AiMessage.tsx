import { MessageRole, MessageStatus, type ConversationMessage } from '@/api/types'
import { Sparkle, Warning } from '@/components/ui/Icon'
import { errorBodyCopy } from '@/lib/errors/messages'
import { formatTime } from '@/lib/dates/format'
import { AiAnswer, AnswerText } from './AiAnswer'
import { ToolTrail } from './AiStreamView'

/** One persisted turn of a conversation. */
export function AiMessage({ message }: { message: ConversationMessage }) {
  if (message.role === MessageRole.USER) {
    return (
      <div className="flex justify-end">
        <div className="max-w-[85%] rounded-xl rounded-br-sm border border-line-strong bg-surface-raised px-3.5 py-2.5">
          <p className="text-[13px] leading-relaxed whitespace-pre-line text-fg">
            {message.content}
          </p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-2.5">
      <div className="flex items-center gap-2 text-[11px] tracking-[0.06em] text-fg-subtle uppercase">
        <Sparkle className="size-3.5 text-accent" />
        MaxAI
        <span className="tracking-normal normal-case">
          · {formatTime(message.created_at)}
        </span>
      </div>

      {message.tool_calls.length > 0 ? (
        <ToolTrail
          tools={message.tool_calls.map((call) => ({
            id: call.id,
            tool: call.tool,
            status: call.status,
          }))}
        />
      ) : null}

      {message.status === MessageStatus.FAILED ? (
        <FailedMessage message={message} />
      ) : message.response ? (
        <AiAnswer response={message.response} />
      ) : (
        <AnswerText text={message.content} />
      )}
    </div>
  )
}

function FailedMessage({ message }: { message: ConversationMessage }) {
  const copy = errorBodyCopy(message.error)

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
