import { useState } from 'react'
import {
  AIIntent,
  DataQuality,
  type AIClaim,
  type AIResponse,
} from '@/api/types'
import { Badge } from '@/components/ui/Badge'
import { ChevronDown, Info, Question } from '@/components/ui/Icon'
import { evidenceTypeLabel } from '@/lib/copy/labels'
import { cn } from '@/lib/utils/cn'

/**
 * Renders a structured AI answer.
 *
 * The response is an object, not a string: the prose is displayed, the claims
 * keep their evidence, and references stay attached to the answer. Evidence is
 * not interactive in the MVP, but it is rendered from the contract so making it
 * clickable later needs no backend change.
 */
export function AiAnswer({
  response,
  className,
}: {
  response: AIResponse
  className?: string
}) {
  if (response.intent === AIIntent.UNSUPPORTED) {
    return <UnsupportedAnswer response={response} className={className} />
  }

  return (
    <div className={cn('space-y-4', className)}>
      <AnswerText text={response.answer} />

      {response.data_quality !== DataQuality.COMPLETE ? (
        <QualityCaveat quality={response.data_quality} />
      ) : null}

      {response.references.length > 0 ? (
        <ul className="flex flex-wrap gap-1.5">
          {response.references.map((reference) => (
            <li key={`${reference.type}:${reference.id}`}>
              <span className="inline-flex items-center rounded-md border border-line-strong bg-surface-raised px-2 py-1 text-[11px] text-fg-muted">
                {reference.label ?? reference.id}
              </span>
            </li>
          ))}
        </ul>
      ) : null}

      {response.claims.length > 0 ? (
        <ClaimsDisclosure claims={response.claims} />
      ) : null}
    </div>
  )
}

export function AnswerText({
  text,
  className,
}: {
  text: string
  className?: string
}) {
  const paragraphs = text.split(/\n{2,}/).filter((part) => part.trim() !== '')

  return (
    <div className={cn('space-y-3', className)}>
      {paragraphs.map((paragraph, index) => (
        <p
          key={index}
          className="text-[14px] leading-relaxed whitespace-pre-line text-fg-muted"
        >
          {paragraph}
        </p>
      ))}
    </div>
  )
}

function UnsupportedAnswer({
  response,
  className,
}: {
  response: AIResponse
  className?: string
}) {
  return (
    <div className={cn('space-y-3', className)}>
      <div className="flex items-center gap-2">
        <Question className="size-4 text-caution" />
        <p className="text-[13px] font-medium text-fg">
          That is outside what MaxAI Crypto can answer
        </p>
      </div>
      <AnswerText text={response.answer} />
      {response.unsupported_reason ? (
        <p className="text-[12px] text-fg-subtle">
          {response.unsupported_reason}
        </p>
      ) : null}
    </div>
  )
}

function QualityCaveat({ quality }: { quality: DataQuality }) {
  const message =
    quality === DataQuality.PARTIAL
      ? 'This answer is based on partially complete data, so treat the figures as approximate.'
      : quality === DataQuality.STALE
        ? 'This answer is based on data that is no longer fresh.'
        : 'Some of the data behind this answer is unavailable.'

  return (
    <p className="flex items-start gap-2 rounded-lg border border-caution/25 bg-caution-quiet/50 px-3 py-2 text-[12px] leading-relaxed text-fg-muted">
      <Info className="mt-0.5 size-3.5 shrink-0 text-caution" />
      {message}
    </p>
  )
}

/** Claims and their backend evidence, collapsed by default. */
function ClaimsDisclosure({ claims }: { claims: AIClaim[] }) {
  const [open, setOpen] = useState(false)

  return (
    <div className="rounded-lg border border-line bg-base-elevated">
      <button
        type="button"
        onClick={() => setOpen((value) => !value)}
        aria-expanded={open}
        className="flex w-full items-center justify-between gap-3 px-3 py-2.5 text-left text-[12px] text-fg-subtle transition-colors hover:text-fg-muted"
      >
        <span>
          {claims.length} {claims.length === 1 ? 'claim' : 'claims'} backed by
          backend data
        </span>
        <ChevronDown
          className={cn('size-3.5 transition-transform', open && 'rotate-180')}
        />
      </button>

      {open ? (
        <ul className="space-y-3 border-t border-line px-3 py-3">
          {claims.map((claim, index) => (
            <li key={index} className="space-y-1.5">
              <p className="text-[13px] leading-relaxed text-fg-muted">
                {claim.text}
              </p>
              {claim.evidence.length > 0 ? (
                <div className="flex flex-wrap gap-1.5">
                  {claim.evidence.map((evidence) => (
                    <Badge
                      key={`${evidence.type}:${evidence.id}`}
                      tone="neutral"
                      title={evidence.id}
                    >
                      {evidenceTypeLabel(evidence.type)}
                    </Badge>
                  ))}
                </div>
              ) : null}
            </li>
          ))}
        </ul>
      ) : null}
    </div>
  )
}
