import { useNavigate } from 'react-router-dom'
import { ChevronDown } from '@/components/ui/Icon'
import { Menu } from '@/components/ui/Menu'
import { chainPresentation } from '@/app/config/chains'
import { truncateMiddle } from '@/lib/formatting/address'
import { useWallets } from '../hooks/useWallets'
import { ChainMonogram } from './ChainMonogram'

/**
 * Active wallet selector.
 *
 * The MVP usually holds one wallet, but the switcher is driven by the list
 * endpoint, so multiple wallets need no new UI surface.
 */
export function WalletSwitcher({ activeWalletId }: { activeWalletId: string }) {
  const navigate = useNavigate()
  const walletsQuery = useWallets()
  const wallets = walletsQuery.data ?? []
  const active = wallets.find((wallet) => wallet.id === activeWalletId)
  const chain = active ? chainPresentation(active.chain_id) : null

  return (
    <Menu
      label="Wallets"
      items={[
        ...wallets.map((wallet) => ({
          id: wallet.id,
          label: `${chainPresentation(wallet.chain_id).name} · ${truncateMiddle(wallet.address)}`,
          description: wallet.label ?? undefined,
          selected: wallet.id === activeWalletId,
          onSelect: () => navigate(`/wallets/${wallet.id}`),
        })),
        {
          id: 'add',
          label: 'Analyze another wallet',
          onSelect: () => navigate('/analyze'),
        },
      ]}
      trigger={(triggerProps) => (
        <button
          type="button"
          {...triggerProps}
          className="flex items-center gap-2.5 rounded-lg border border-line bg-surface px-2.5 py-1.5 text-left transition-colors hover:border-line-strong"
        >
          {active ? <ChainMonogram chainId={active.chain_id} size="sm" /> : null}
          <span className="min-w-0">
            <span className="block text-[12px] leading-tight text-fg">
              {chain?.name ?? 'Wallet'}
            </span>
            <span className="block font-mono text-[11px] leading-tight text-fg-subtle">
              {active ? truncateMiddle(active.address) : '—'}
            </span>
          </span>
          <ChevronDown className="size-3.5 shrink-0 text-fg-subtle" />
        </button>
      )}
    />
  )
}
