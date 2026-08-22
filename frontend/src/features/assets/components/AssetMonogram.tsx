import type { Asset } from '@/api/types'
import { chainPresentation } from '@/app/config/chains'
import { cn } from '@/lib/utils/cn'

/**
 * Asset mark.
 *
 * Token icons are not fetched from third parties: the backend may provide
 * `icon_url`, and until then a monogram keeps the list calm and avoids
 * leaking requests to external hosts.
 */
export function AssetMonogram({
  asset,
  size = 'md',
  className,
}: {
  asset: Asset
  size?: 'sm' | 'md'
  className?: string
}) {
  const accent = chainPresentation(asset.chain_id).accent

  if (asset.icon_url) {
    return (
      <img
        src={asset.icon_url}
        alt=""
        className={cn(
          'shrink-0 rounded-full',
          size === 'sm' ? 'size-7' : 'size-9',
          className,
        )}
      />
    )
  }

  return (
    <span
      className={cn(
        'inline-grid shrink-0 place-items-center rounded-full border border-line-strong bg-surface-raised font-medium text-fg-muted',
        size === 'sm' ? 'size-7 text-[10px]' : 'size-9 text-[11px]',
        className,
      )}
      style={
        asset.contract_address === null
          ? { borderColor: `${accent}40`, color: accent }
          : undefined
      }
      aria-hidden="true"
    >
      {asset.symbol.slice(0, 4)}
    </span>
  )
}
