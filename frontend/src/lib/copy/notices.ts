import { DataNoticeCode, type DataNotice } from '@/api/types'

/**
 * Data-quality notice copy.
 *
 * The backend sends a code plus parameters; the wording lives here so the
 * product voice stays consistent and translatable.
 */
export function noticeMessage(notice: DataNotice): string {
  const count = Number(notice.params?.count ?? 0)
  const minutes = Number(notice.params?.minutes ?? 0)

  switch (notice.code) {
    case DataNoticeCode.UNPRICED_ASSETS_EXCLUDED:
      return count === 1
        ? 'Portfolio value is partially calculated because the price of 1 asset is currently unavailable.'
        : `Portfolio value is partially calculated because the prices of ${count} assets are currently unavailable.`

    case DataNoticeCode.DATA_STALE:
      return `Based on portfolio data last updated ${minutes} minutes ago.`

    case DataNoticeCode.HISTORY_INCOMPLETE:
      return 'Some historical snapshots are missing, so this period is incomplete.'

    case DataNoticeCode.SYNC_PARTIALLY_FAILED:
      return 'The last synchronisation completed only partly, so some data may be missing.'

    case DataNoticeCode.NFTS_EXCLUDED_FROM_VALUATION:
      return 'NFTs are not included in portfolio valuation.'

    case DataNoticeCode.DEFI_POSITIONS_EXCLUDED:
      return 'DeFi positions are not included in portfolio valuation.'

    default:
      return 'Some data in this view is incomplete.'
  }
}
