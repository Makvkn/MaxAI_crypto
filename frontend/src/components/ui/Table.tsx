import type {
  HTMLAttributes,
  TdHTMLAttributes,
  ThHTMLAttributes,
} from 'react'
import { cn } from '@/lib/utils/cn'

/**
 * Table primitives for dense financial data. Semantic `<table>` markup is used
 * so rows and columns are navigable and announced correctly.
 */
export function TableScroll({
  className,
  ...props
}: HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={cn('-mx-px overflow-x-auto', className)}
      tabIndex={0}
      role="region"
      {...props}
    />
  )
}

export function Table({
  className,
  ...props
}: HTMLAttributes<HTMLTableElement>) {
  return (
    <table
      className={cn('w-full border-collapse text-sm', className)}
      {...props}
    />
  )
}

export function Th({
  className,
  align = 'left',
  ...props
}: ThHTMLAttributes<HTMLTableCellElement> & { align?: 'left' | 'right' }) {
  return (
    <th
      scope="col"
      className={cn(
        'px-4 py-2.5 text-[11px] font-medium tracking-[0.06em] text-fg-subtle uppercase',
        align === 'right' ? 'text-right' : 'text-left',
        className,
      )}
      {...props}
    />
  )
}

export function Td({
  className,
  align = 'left',
  ...props
}: TdHTMLAttributes<HTMLTableCellElement> & { align?: 'left' | 'right' }) {
  return (
    <td
      className={cn(
        'px-4 py-3 align-middle text-fg',
        align === 'right' ? 'text-right' : 'text-left',
        className,
      )}
      {...props}
    />
  )
}

export function Tr({
  className,
  interactive,
  ...props
}: HTMLAttributes<HTMLTableRowElement> & { interactive?: boolean }) {
  return (
    <tr
      className={cn(
        'border-t border-line/70',
        interactive &&
          'cursor-pointer transition-colors hover:bg-surface-raised/60 focus-within:bg-surface-raised/60',
        className,
      )}
      {...props}
    />
  )
}
