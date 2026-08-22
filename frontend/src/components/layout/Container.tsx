import type { HTMLAttributes } from 'react'
import { cn } from '@/lib/utils/cn'

export function Container({
  className,
  size = 'default',
  ...props
}: HTMLAttributes<HTMLDivElement> & { size?: 'default' | 'wide' | 'narrow' }) {
  return (
    <div
      className={cn(
        'mx-auto w-full px-5 sm:px-8',
        size === 'narrow' && 'max-w-3xl',
        size === 'default' && 'max-w-6xl',
        size === 'wide' && 'max-w-[1440px]',
        className,
      )}
      {...props}
    />
  )
}
