import type { ChainId } from '@/api/types'
import { ChevronRight } from '@/components/ui/Icon'
import { SUPPORTED_CHAINS } from '@/app/config/chains'
import { ChainMonogram } from '@/features/wallets/components/ChainMonogram'
import { cn } from '@/lib/utils/cn'

/**
 * Network selection.
 *
 * The chain is always chosen explicitly — the MVP deliberately does not detect
 * it from the address. The list is rendered from configuration, so a new chain
 * appears here without touching this component.
 */
export function ChainSelectStep({
  selected,
  onSelect,
}: {
  selected: ChainId | null
  onSelect: (chainId: ChainId) => void
}) {
  return (
    <div>
      <h1 className="text-2xl font-medium tracking-tight text-fg">
        Which network is your wallet on?
      </h1>
      <p className="mt-2 text-sm leading-relaxed text-fg-muted">
        Select the blockchain you want to analyse. You can add another wallet
        later.
      </p>

      <ul className="mt-8 grid gap-2.5 sm:grid-cols-2">
        {SUPPORTED_CHAINS.map((chain) => {
          const isSelected = chain.id === selected
          return (
            <li key={chain.id}>
              <button
                type="button"
                onClick={() => onSelect(chain.id)}
                aria-pressed={isSelected}
                className={cn(
                  'group flex w-full items-center gap-3.5 rounded-card border px-4 py-3.5 text-left transition-colors',
                  isSelected
                    ? 'border-accent/60 bg-accent-quiet/40'
                    : 'border-line bg-surface hover:border-line-strong hover:bg-surface-hover',
                )}
              >
                <ChainMonogram chainId={chain.id} />
                <span className="min-w-0 flex-1">
                  <span className="block text-sm font-medium text-fg">
                    {chain.name}
                  </span>
                  <span className="block truncate text-[12px] text-fg-subtle">
                    {chain.summary}
                  </span>
                </span>
                <ChevronRight className="size-4 shrink-0 text-fg-subtle transition-transform group-hover:translate-x-0.5" />
              </button>
            </li>
          )
        })}
      </ul>
    </div>
  )
}
