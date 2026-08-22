import type { Timestamp } from '@/api/types'
import { UNKNOWN_VALUE } from '../formatting/money'

/** Date and time presentation. Timestamps arrive as UTC ISO strings. */

const dateTime = new Intl.DateTimeFormat('en-US', {
  month: 'short',
  day: 'numeric',
  hour: '2-digit',
  minute: '2-digit',
})

const dateOnly = new Intl.DateTimeFormat('en-US', {
  month: 'short',
  day: 'numeric',
  year: 'numeric',
})

const timeOnly = new Intl.DateTimeFormat('en-US', {
  hour: '2-digit',
  minute: '2-digit',
})

function parse(value: Timestamp | null | undefined): Date | null {
  if (!value) return null
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? null : date
}

export function formatDateTime(value: Timestamp | null | undefined): string {
  const date = parse(value)
  return date ? dateTime.format(date) : UNKNOWN_VALUE
}

export function formatDate(value: Timestamp | null | undefined): string {
  const date = parse(value)
  return date ? dateOnly.format(date) : UNKNOWN_VALUE
}

export function formatTime(value: Timestamp | null | undefined): string {
  const date = parse(value)
  return date ? timeOnly.format(date) : UNKNOWN_VALUE
}

/** `"42 minutes ago"`. Used heavily by freshness messaging. */
export function formatRelativeTime(
  value: Timestamp | null | undefined,
  now: Date = new Date(),
): string {
  const date = parse(value)
  if (!date) return UNKNOWN_VALUE

  const diffMs = now.getTime() - date.getTime()
  const minutes = Math.round(diffMs / 60_000)

  if (minutes < 1) return 'just now'
  if (minutes === 1) return '1 minute ago'
  if (minutes < 60) return `${minutes} minutes ago`

  const hours = Math.round(minutes / 60)
  if (hours === 1) return '1 hour ago'
  if (hours < 24) return `${hours} hours ago`

  const days = Math.round(hours / 24)
  if (days === 1) return 'yesterday'
  if (days < 30) return `${days} days ago`

  return formatDate(value)
}

/** Whole minutes elapsed, for copy such as "updated 42 minutes ago". */
export function minutesSince(
  value: Timestamp | null | undefined,
  now: Date = new Date(),
): number | null {
  const date = parse(value)
  if (!date) return null
  return Math.max(0, Math.round((now.getTime() - date.getTime()) / 60_000))
}
