import { DataQuality, ValuationStatus } from '@/api/types'
import { Badge } from '@/components/ui/Badge'
import { Tooltip } from '@/components/ui/Tooltip'
import { dataQualityLabels } from '@/lib/copy/labels'

/**
 * Marks a figure whose completeness is not `COMPLETE`.
 *
 * Placed next to the number it qualifies, so a partial total never reads as an
 * exact one.
 */
export function ValuationStatusNote({
  valuationStatus,
  dataQuality,
  unpricedCount,
}: {
  valuationStatus: ValuationStatus
  dataQuality: DataQuality
  unpricedCount: number
}) {
  if (
    valuationStatus === ValuationStatus.COMPLETE &&
    dataQuality === DataQuality.COMPLETE
  ) {
    return null
  }

  if (valuationStatus === ValuationStatus.PARTIAL) {
    return (
      <Tooltip
        content={
          unpricedCount === 1
            ? 'One held asset has no reliable market price, so it is excluded from this total.'
            : `${unpricedCount} held assets have no reliable market price, so they are excluded from this total.`
        }
      >
        <Badge tone="caution">Partial</Badge>
      </Tooltip>
    )
  }

  if (valuationStatus === ValuationStatus.UNAVAILABLE) {
    return (
      <Tooltip content="Valuation data is unavailable, so no total is shown.">
        <Badge tone="negative">Unavailable</Badge>
      </Tooltip>
    )
  }

  return (
    <Tooltip content="This figure is based on data that is no longer fresh.">
      <Badge tone="caution">{dataQualityLabels[dataQuality]}</Badge>
    </Tooltip>
  )
}
