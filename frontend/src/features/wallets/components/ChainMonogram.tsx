import type { ChainId } from '@/api/types'
import { chainPresentation } from '@/app/config/chains'
import { cn } from '@/lib/utils/cn'

/**
 * Chain mark derived from configuration.
 *
 * No per-chain branching lives in components: colour and symbol come from the
 * chain presentation registry, so a new chain needs no code here.
 */
export function ChainMonogram({
  chainId,
  size = 'md',
  className,
}: {
  chainId: ChainId
  size?: 'sm' | 'md' | 'lg'
  className?: string
}) {
  const chain = chainPresentation(chainId)

  return (
    <span
      className={cn(
        'inline-grid shrink-0 place-items-center rounded-full border font-medium',
        size === 'sm' && 'size-6 text-[10px]',
        size === 'md' && 'size-8 text-[11px]',
        size === 'lg' && 'size-11 text-[13px]',
        className,
      )}
      style={{
        borderColor: `${chain.accent}40`,
        backgroundColor: `${chain.accent}14`,
        color: chain.accent,
      }}
      aria-hidden="true"
    >
      {chain.nativeSymbol.slice(0, 3)}
    </span>
  )
}
