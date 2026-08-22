import type { ButtonHTMLAttributes, ReactNode } from 'react'
import { Link, type LinkProps } from 'react-router-dom'
import { cn } from '@/lib/utils/cn'
import { Spinner } from './Spinner'

type Variant = 'primary' | 'secondary' | 'ghost' | 'quiet' | 'danger'
type Size = 'sm' | 'md' | 'lg'

const BASE =
  'relative inline-flex items-center justify-center gap-2 rounded-lg font-medium transition-colors duration-150 select-none disabled:cursor-not-allowed disabled:opacity-45 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent'

const VARIANTS: Record<Variant, string> = {
  primary:
    'bg-accent text-accent-fg hover:bg-accent-hover active:bg-accent shadow-[0_1px_0_0_rgba(255,255,255,0.12)_inset]',
  secondary:
    'border border-line-strong bg-surface-raised text-fg hover:bg-surface-hover hover:border-line-strong',
  ghost: 'text-fg-muted hover:text-fg hover:bg-surface-raised',
  quiet: 'border border-line bg-transparent text-fg-muted hover:text-fg hover:border-line-strong',
  danger: 'border border-negative/40 bg-negative/10 text-negative hover:bg-negative/15',
}

const SIZES: Record<Size, string> = {
  sm: 'h-8 px-3 text-[13px]',
  md: 'h-10 px-4 text-sm',
  lg: 'h-12 px-6 text-[15px]',
}

interface CommonProps {
  variant?: Variant
  size?: Size
  loading?: boolean
  iconLeft?: ReactNode
  iconRight?: ReactNode
  fullWidth?: boolean
}

export type ButtonProps = CommonProps &
  ButtonHTMLAttributes<HTMLButtonElement> & { children?: ReactNode }

export function Button({
  variant = 'primary',
  size = 'md',
  loading = false,
  iconLeft,
  iconRight,
  fullWidth,
  className,
  children,
  disabled,
  ...props
}: ButtonProps) {
  return (
    <button
      className={cn(
        BASE,
        VARIANTS[variant],
        SIZES[size],
        fullWidth && 'w-full',
        className,
      )}
      disabled={disabled || loading}
      aria-busy={loading || undefined}
      {...props}
    >
      {loading ? <Spinner className="size-4" /> : iconLeft}
      {children}
      {!loading && iconRight}
    </button>
  )
}

export type ButtonLinkProps = CommonProps &
  LinkProps & { children?: ReactNode }

/** Same visual language as `Button`, but a real link for real navigation. */
export function ButtonLink({
  variant = 'primary',
  size = 'md',
  iconLeft,
  iconRight,
  fullWidth,
  className,
  children,
  ...props
}: ButtonLinkProps) {
  return (
    <Link
      className={cn(
        BASE,
        VARIANTS[variant],
        SIZES[size],
        fullWidth && 'w-full',
        className,
      )}
      {...props}
    >
      {iconLeft}
      {children}
      {iconRight}
    </Link>
  )
}
