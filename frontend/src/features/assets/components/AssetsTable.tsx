import {
  AssetVisibility,
  type Portfolio,
  type WalletPosition,
} from '@/api/types'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Card, CardHeader } from '@/components/ui/Card'
import { ChevronDown, Scenario } from '@/components/ui/Icon'
import { Table, TableScroll, Td, Th, Tr } from '@/components/ui/Table'
import { Tooltip } from '@/components/ui/Tooltip'
import { EmptyState } from '@/components/feedback/States'
import { Delta, Money, UnknownValue } from '@/components/finance/Money'
import { AssetMonogram } from './AssetMonogram'
import { formatBalance, formatPrice } from '@/lib/formatting/money'
import { formatPercent } from '@/lib/formatting/percent'
import { visibilityLabels } from '@/lib/copy/labels'
import { useUiStore } from '@/stores/uiStore'
import { cn } from '@/lib/utils/cn'

/**
 * Holdings.
 *
 * Visibility is decided by the backend: dust and spam classification never
 * happens here. Hidden positions are collapsed but always reachable, and an
 * unpriced asset shows its balance with no invented value.
 */
export function AssetsTable({
  portfolio,
  onSimulate,
}: {
  portfolio: Portfolio
  onSimulate?: (position: WalletPosition) => void
}) {
  const expanded = useUiStore((state) => state.hiddenAssetsExpanded)
  const toggleHidden = useUiStore((state) => state.toggleHiddenAssets)

  const visible = portfolio.positions.filter(
    (position) => position.visibility === AssetVisibility.VISIBLE,
  )
  const hidden = portfolio.positions.filter(
    (position) => position.visibility !== AssetVisibility.VISIBLE,
  )

  return (
    <Card>
      <CardHeader
        title="Assets"
        action={
          <span className="text-[12px] text-fg-subtle">
            {portfolio.visible_positions_count} held
          </span>
        }
      />

      {portfolio.positions.length === 0 ? (
        <EmptyState
          title="No assets"
          description="This wallet has no token or native balances on the selected chain."
        />
      ) : visible.length === 0 ? (
        <EmptyState
          title="No visible assets"
          description="This wallet holds no assets that pass the backend's visibility rules."
        />
      ) : (
        <TableScroll aria-label="Assets">
          <Table>
            <thead>
              <tr>
                <Th>Asset</Th>
                <Th align="right">Balance</Th>
                <Th align="right">Price</Th>
                <Th align="right">Value</Th>
                <Th align="right">Allocation</Th>
                <Th align="right">24h</Th>
                {onSimulate ? <Th align="right">Scenario</Th> : null}
              </tr>
            </thead>
            <tbody>
              {visible.map((position) => (
                <AssetRow
                  key={position.asset.id}
                  position={position}
                  onSimulate={onSimulate}
                />
              ))}
            </tbody>
          </Table>
        </TableScroll>
      )}

      {hidden.length > 0 ? (
        <div className="border-t border-line">
          <button
            type="button"
            onClick={toggleHidden}
            aria-expanded={expanded}
            className="flex w-full items-center justify-between gap-3 px-5 py-3.5 text-left text-[13px] text-fg-muted transition-colors hover:bg-surface-raised/60"
          >
            <span>
              Hidden assets ({portfolio.hidden_positions_count})
              <span className="ml-2 text-fg-subtle">
                dust, spam and unrecognised tokens
              </span>
            </span>
            <ChevronDown
              className={cn(
                'size-4 shrink-0 transition-transform',
                expanded && 'rotate-180',
              )}
            />
          </button>

          {expanded ? (
            <TableScroll aria-label="Hidden assets">
              <Table>
                <thead>
                  <tr>
                    <Th>Asset</Th>
                    <Th align="right">Balance</Th>
                    <Th align="right">Price</Th>
                    <Th align="right">Value</Th>
                    <Th align="right">Reason</Th>
                  </tr>
                </thead>
                <tbody>
                  {hidden.map((position) => (
                    <Tr key={position.asset.id}>
                      <Td>
                        <AssetCell position={position} />
                      </Td>
                      <Td align="right" className="text-fg-muted">
                        {formatBalance(position.balance)}
                      </Td>
                      <Td align="right" className="text-fg-muted">
                        <PriceCell position={position} />
                      </Td>
                      <Td align="right">
                        <Money
                          value={position.value_usd}
                          unknownReason="No reliable market price"
                        />
                      </Td>
                      <Td align="right">
                        <Badge tone="muted">
                          {visibilityLabels[position.visibility]}
                        </Badge>
                      </Td>
                    </Tr>
                  ))}
                </tbody>
              </Table>
            </TableScroll>
          ) : null}
        </div>
      ) : null}
    </Card>
  )
}

function AssetRow({
  position,
  onSimulate,
}: {
  position: WalletPosition
  onSimulate?: (position: WalletPosition) => void
}) {
  return (
    <Tr>
      <Td>
        <AssetCell position={position} />
      </Td>
      <Td align="right" className="text-fg-muted">
        {formatBalance(position.balance)}
      </Td>
      <Td align="right" className="text-fg-muted">
        <PriceCell position={position} />
      </Td>
      <Td align="right" className="font-medium">
        <Money
          value={position.value_usd}
          unknownReason="No reliable market price for this asset"
        />
      </Td>
      <Td align="right" className="text-fg-muted">
        {formatPercent(position.allocation_pct)}
      </Td>
      <Td align="right">
        <Delta percent={position.change_24h_pct} size="sm" showIcon={false} />
      </Td>
      {onSimulate ? (
        <Td align="right">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onSimulate(position)}
            iconLeft={<Scenario className="size-3.5" />}
            // A position the backend could not value cannot be projected.
            disabled={position.value_usd === null}
          >
            Simulate
          </Button>
        </Td>
      ) : null}
    </Tr>
  )
}

function AssetCell({ position }: { position: WalletPosition }) {
  return (
    <div className="flex items-center gap-3">
      <AssetMonogram asset={position.asset} size="sm" />
      <div className="min-w-0">
        <p className="truncate text-[13px] font-medium text-fg">
          {position.asset.symbol}
        </p>
        <p className="truncate text-[12px] text-fg-subtle">
          {position.asset.name}
        </p>
      </div>
    </div>
  )
}

/** An unpriced asset shows why there is no price, never a zero. */
function PriceCell({ position }: { position: WalletPosition }) {
  if (position.price === null || position.price.value_usd === null) {
    return (
      <Tooltip content="No reliable market price is available for this asset, so it cannot be valued.">
        <UnknownValue reason="Price unavailable" />
      </Tooltip>
    )
  }
  return <span className="tabular">{formatPrice(position.price.value_usd)}</span>
}
