package portfolio

import (
	"context"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/maxaicrypto/backend/internal/domain/apperr"
	"github.com/maxaicrypto/backend/internal/domain/asset"
	"github.com/maxaicrypto/backend/internal/domain/performance"
	"github.com/maxaicrypto/backend/internal/domain/portfolio"
	"github.com/maxaicrypto/backend/internal/domain/shared"
	"github.com/maxaicrypto/backend/internal/domain/wallet"
)

const performanceSeriesLimit = 500

// PerformanceApp implements PerformanceService.
type PerformanceApp struct {
	wallets    wallet.Repository
	syncStates wallet.SyncStateRepository
	snapshots  portfolio.SnapshotRepository
	assets     asset.Repository
}

// NewPerformanceApp wires snapshot-based performance reads.
func NewPerformanceApp(
	wallets wallet.Repository,
	syncStates wallet.SyncStateRepository,
	snapshots portfolio.SnapshotRepository,
	assets asset.Repository,
) *PerformanceApp {
	return &PerformanceApp{
		wallets:    wallets,
		syncStates: syncStates,
		snapshots:  snapshots,
		assets:     assets,
	}
}

// Get implements PerformanceService.
func (a *PerformanceApp) Get(ctx context.Context, userID, walletID uuid.UUID, period performance.Period) (performance.Performance, error) {
	if err := a.requireReadyWallet(ctx, userID, walletID); err != nil {
		return performance.Performance{}, err
	}

	result := performance.Performance{
		WalletID: walletID,
		Period:   period,
		Status:   performance.StatusUnavailable,
		Currency: shared.CurrencyUSD,
		Series:   []performance.SeriesPoint{},
		Drivers:  []performance.Driver{},
	}

	closing, ok, err := a.snapshots.GetLatestValid(ctx, walletID)
	if err != nil {
		return performance.Performance{}, err
	}
	if !ok || !closing.IsValid() {
		reason := string(apperr.CodePerformanceDataUnavailable)
		result.UnavailableReason = &reason
		result.DataQuality = shared.DataQualityUnavailable
		return result, nil
	}

	opening, ok, err := a.locateOpeningSnapshot(ctx, walletID, period, closing)
	if err != nil {
		return performance.Performance{}, err
	}
	if !ok || !opening.IsValid() {
		reason := string(apperr.CodePerformanceDataUnavailable)
		result.UnavailableReason = &reason
		result.DataQuality = shared.DataQualityUnavailable
		return result, nil
	}

	series, err := a.snapshots.ListBetween(ctx, walletID, opening.CapturedAt, closing.CapturedAt, performanceSeriesLimit)
	if err != nil {
		return performance.Performance{}, err
	}

	openingValue := opening.TotalValueUSD.Decimal.Value()
	if openingValue.IsZero() {
		reason := string(apperr.CodePerformanceDataUnavailable)
		result.UnavailableReason = &reason
		result.DataQuality = shared.DataQualityUnavailable
		return result, nil
	}

	closingValue := closing.TotalValueUSD.Decimal.Value()
	changeUSD := closingValue.Sub(openingValue)
	changePct := changeUSD.Div(openingValue).Mul(decimal.NewFromInt(100))

	result.Opening = snapshotEndpoint(opening)
	result.Closing = snapshotEndpoint(closing)
	result.ChangeUSD = shared.Known(shared.NewDecimal(changeUSD))
	result.ChangePct = shared.Known(shared.NewDecimal(changePct))
	result.Series = mapSeries(series)
	result.DataQuality = mergeDataQuality(opening.DataQuality, closing.DataQuality)
	result.Status = derivePerformanceStatus(opening, closing, result.DataQuality)

	drivers, err := a.buildDrivers(ctx, opening, closing)
	if err != nil {
		return performance.Performance{}, err
	}
	result.Drivers = drivers

	calcID := calculationID(walletID, period, opening.ID, closing.ID)
	version := performance.CalculationVersion
	result.CalculationID = &calcID
	result.CalculationVersion = &version

	return result, nil
}

func (a *PerformanceApp) requireReadyWallet(ctx context.Context, userID, walletID uuid.UUID) error {
	w, err := a.wallets.GetByID(ctx, walletID)
	if err != nil {
		if appErr := apperr.From(err); appErr != nil && appErr.Code == apperr.CodeNotFound {
			return apperr.New(apperr.CodeWalletNotFound)
		}
		return err
	}
	if w.UserID != userID {
		return apperr.New(apperr.CodeWalletNotFound)
	}

	syncState, err := a.syncStates.Get(ctx, walletID)
	if err != nil {
		return err
	}
	switch syncState.Status {
	case wallet.SyncPending, wallet.SyncSyncing:
		return apperr.New(apperr.CodeWalletNotReady).
			WithDetail("sync_status", string(syncState.Status))
	case wallet.SyncFailed:
		return apperr.New(apperr.CodeWalletSyncFailed)
	}
	return nil
}

func (a *PerformanceApp) locateOpeningSnapshot(
	ctx context.Context,
	walletID uuid.UUID,
	period performance.Period,
	closing portfolio.Snapshot,
) (portfolio.Snapshot, bool, error) {
	if period == performance.PeriodAllTime {
		return a.snapshots.GetFirstValid(ctx, walletID)
	}

	lookback, ok := period.Lookback()
	if !ok {
		return portfolio.Snapshot{}, false, nil
	}
	anchor := closing.CapturedAt.Add(-lookback)
	return a.snapshots.GetClosestValidBefore(ctx, walletID, anchor)
}

func snapshotEndpoint(snapshot portfolio.Snapshot) *performance.Endpoint {
	if !snapshot.IsValid() {
		return nil
	}
	return &performance.Endpoint{
		SnapshotID: snapshot.ID,
		CapturedAt: snapshot.CapturedAt,
		ValueUSD:   snapshot.TotalValueUSD.Decimal,
		Status:     snapshot.Status,
	}
}

func mapSeries(snapshots []portfolio.Snapshot) []performance.SeriesPoint {
	series := make([]performance.SeriesPoint, len(snapshots))
	for i, snapshot := range snapshots {
		point := performance.SeriesPoint{
			CapturedAt: snapshot.CapturedAt,
			Status:     snapshot.Status,
		}
		if snapshot.TotalValueUSD.Valid {
			point.TotalValueUSD = snapshot.TotalValueUSD
		} else {
			point.TotalValueUSD = shared.Unknown()
		}
		series[i] = point
	}
	return series
}

func mergeDataQuality(a, b shared.DataQuality) shared.DataQuality {
	if a == shared.DataQualityUnavailable || b == shared.DataQualityUnavailable {
		return shared.DataQualityUnavailable
	}
	if a == shared.DataQualityPartial || b == shared.DataQualityPartial {
		return shared.DataQualityPartial
	}
	return shared.DataQualityComplete
}

func derivePerformanceStatus(opening, closing portfolio.Snapshot, quality shared.DataQuality) performance.Status {
	if quality == shared.DataQualityUnavailable {
		return performance.StatusUnavailable
	}
	if opening.Status == shared.ValuationPartial ||
		closing.Status == shared.ValuationPartial ||
		quality == shared.DataQualityPartial {
		return performance.StatusPartial
	}
	return performance.StatusAvailable
}

func calculationID(walletID uuid.UUID, period performance.Period, openingID, closingID uuid.UUID) uuid.UUID {
	raw := fmt.Sprintf("%s:%s:%s:%s", walletID, period, openingID, closingID)
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(raw))
}

func (a *PerformanceApp) buildDrivers(
	ctx context.Context,
	opening, closing portfolio.Snapshot,
) ([]performance.Driver, error) {
	openPositions, err := a.snapshots.ListPositions(ctx, opening.ID)
	if err != nil {
		return nil, err
	}
	closePositions, err := a.snapshots.ListPositions(ctx, closing.ID)
	if err != nil {
		return nil, err
	}

	openValues := positionValues(openPositions)
	closeValues := positionValues(closePositions)

	assetIDs := make([]uuid.UUID, 0, len(openValues.values)+len(closeValues.values))
	seen := make(map[uuid.UUID]struct{}, len(openValues.values)+len(closeValues.values))
	for assetID := range openValues.values {
		if _, ok := seen[assetID]; !ok {
			seen[assetID] = struct{}{}
			assetIDs = append(assetIDs, assetID)
		}
	}
	for assetID := range closeValues.values {
		if _, ok := seen[assetID]; !ok {
			seen[assetID] = struct{}{}
			assetIDs = append(assetIDs, assetID)
		}
	}
	if len(assetIDs) == 0 {
		return []performance.Driver{}, nil
	}

	assets, err := a.assets.GetManyByID(ctx, assetIDs)
	if err != nil {
		return nil, err
	}

	closingTotal := closing.TotalValueUSD.Decimal.Value()
	drivers := make([]performance.Driver, 0, len(assetIDs))
	for _, assetID := range assetIDs {
		ast, ok := assets[assetID]
		if !ok {
			return nil, apperr.New(apperr.CodeInternal)
		}

		openValue := openValues.values[assetID]
		closeValue := closeValues.values[assetID]
		contribution := closeValue.Sub(openValue)

		driver := performance.Driver{
			Asset: ast,
		}
		if closeAlloc, ok := closeValues.allocation(assetID); ok {
			driver.AllocationPct = closeAlloc
		}
		if !contribution.IsZero() || closeValue.IsPositive() || openValue.IsPositive() {
			driver.ContributionUSD = shared.Known(shared.NewDecimal(contribution))
		}
		if !closingTotal.IsZero() {
			pct := contribution.Div(closingTotal).Mul(decimal.NewFromInt(100))
			driver.ContributionPct = shared.Known(shared.NewDecimal(pct))
		}
		if openValue.IsPositive() {
			changePct := closeValue.Sub(openValue).Div(openValue).Mul(decimal.NewFromInt(100))
			driver.ChangePct = shared.Known(shared.NewDecimal(changePct))
		}

		drivers = append(drivers, driver)
	}

	sort.Slice(drivers, func(i, j int) bool {
		return driverMagnitude(drivers[i]) > driverMagnitude(drivers[j])
	})
	if len(drivers) > 5 {
		drivers = drivers[:5]
	}
	return drivers, nil
}

type valuedPositions struct {
	values      map[uuid.UUID]decimal.Decimal
	allocations map[uuid.UUID]shared.NullDecimal
}

func positionValues(positions []portfolio.SnapshotPosition) valuedPositions {
	result := valuedPositions{
		values:      make(map[uuid.UUID]decimal.Decimal, len(positions)),
		allocations: make(map[uuid.UUID]shared.NullDecimal, len(positions)),
	}
	for _, position := range positions {
		if position.ValueUSD.Valid {
			result.values[position.AssetID] = position.ValueUSD.Decimal.Value()
		} else {
			result.values[position.AssetID] = decimal.Zero
		}
		if position.AllocationPct.Valid {
			result.allocations[position.AssetID] = position.AllocationPct
		}
	}
	return result
}

func (v valuedPositions) allocation(assetID uuid.UUID) (shared.NullDecimal, bool) {
	allocation, ok := v.allocations[assetID]
	return allocation, ok
}

func driverMagnitude(driver performance.Driver) float64 {
	if !driver.ContributionUSD.Valid {
		return 0
	}
	f, _ := driver.ContributionUSD.Decimal.Value().Abs().Float64()
	return f
}

var _ PerformanceService = (*PerformanceApp)(nil)
