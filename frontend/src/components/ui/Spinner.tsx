import { cn } from '@/lib/utils/cn'

/** Small inline activity indicator. Full-screen spinners are avoided. */
export function Spinner({ className }: { className?: string }) {
  return (
    <span
      className={cn(
        'inline-block size-4 animate-spin rounded-full border-2 border-current border-t-transparent align-middle',
        className,
      )}
      role="presentation"
    />
  )
}
