package portfolio

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	apppricing "github.com/maxaicrypto/backend/internal/application/pricing"
	"github.com/maxaicrypto/backend/internal/domain/apperr"
	"github.com/maxaicrypto/backend/internal/domain/asset"
	"github.com/maxaicrypto/backend/internal/domain/portfolio"
	"github.com/maxaicrypto/backend/internal/domain/position"
	"github.com/maxaicrypto/backend/internal/domain/price"
	"github.com/maxaicrypto/backend/internal/domain/shared"
	"github.com/maxaicrypto/backend/internal/domain/wallet"
)

// Calculator builds a portfolio from positions and prices.
type Calculator struct {
	assets     asset.Repository
	positions  position.Repository
	pricing    apppricing.Service
	visibility asset.VisibilityClassifier
	freshness  shared.FreshnessThresholds
}

// NewCalculator wires the portfolio calculator.
func NewCalculator(
	assets asset.Repository,
	positions position.Repository,
	pricing apppricing.Service,
	freshness shared.FreshnessThresholds,
) *Calculator {
	return &Calculator{
		assets:     assets,
		positions:  positions,
		pricing:    pricing,
		visibility: asset.NewRulesVisibilityClassifier(),
		freshness:  freshness,
	}
}

// Build computes the current portfolio for a wallet.
func (c *Calculator) Build(ctx context.Context, w wallet.Wallet, lastSyncedAt *time.Time) (portfolio.Portfolio, error) {
	positions, err := c.positions.ListByWallet(ctx, w.ID)
	if err != nil {
		return portfolio.Portfolio{}, err
	}
	now := time.Now().UTC()
	if len(positions) == 0 {
		// A finished sync with no holdings is a known empty portfolio: report
		// $0 rather than unavailable. Unknown≠zero still applies when prices
		// are missing for held assets below.
		freshness := shared.FreshnessVeryStale
		if lastSyncedAt != nil {
			freshness = c.freshness.ClassifyAt(*lastSyncedAt, now)
		}
		zero := shared.Known(shared.NewDecimal(decimal.Zero))
		return portfolio.Portfolio{
			WalletID:           w.ID,
			Currency:           shared.CurrencyUSD,
			TotalValueUSD:      zero,
			ValuationStatus:    shared.ValuationComplete,
			DataQuality:        shared.DataQualityComplete,
			DataFreshness:      freshness,
			Change24hUSD:       zero,
			Change24hPct:       zero,
			AsOf:               now,
			LastSyncedAt:       lastSyncedAt,
			CalculationVersion: portfolio.CalculationVersion,
			Positions:          []portfolio.Position{},
		}, nil
	}

	assetIDs := make([]uuid.UUID, len(positions))
	for i, p := range positions {
		assetIDs[i] = p.AssetID
	}
	assets, err := c.assets.GetManyByID(ctx, assetIDs)
	if err != nil {
		return portfolio.Portfolio{}, err
	}
	prices, err := c.pricing.GetCurrent(ctx, assetIDs)
	if err != nil {
		return portfolio.Portfolio{}, err
	}

	result := portfolio.Portfolio{
		WalletID:           w.ID,
		Currency:           shared.CurrencyUSD,
		AsOf:               now,
		LastSyncedAt:       lastSyncedAt,
		CalculationVersion: portfolio.CalculationVersion,
		Positions:          make([]portfolio.Position, 0, len(positions)),
	}
	if lastSyncedAt != nil {
		result.DataFreshness = c.freshness.ClassifyAt(*lastSyncedAt, now)
	} else {
		result.DataFreshness = shared.FreshnessVeryStale
	}

	total := decimal.Zero
	valued := 0
	unpriced := 0
	for _, pos := range positions {
		ast, ok := assets[pos.AssetID]
		if !ok {
			return portfolio.Portfolio{}, apperr.New(apperr.CodeInternal)
		}
		item := portfolio.Position{
			Asset:           ast,
			Balance:         pos.BalanceNormalized,
			BalanceRaw:      pos.BalanceRaw,
			Visibility:      asset.VisibilityVisible,
			UpdatedAt:       pos.UpdatedAt,
			ValuationStatus: shared.ValuationUnavailable,
		}
		if quote, ok := prices[pos.AssetID]; ok && quote.IsUsable() {
			item.Price = &quote
			value := pos.BalanceNormalized.Value().Mul(quote.ValueUSD.Decimal.Value())
			item.ValueUSD = shared.Known(shared.NewDecimal(value))
			item.ValuationStatus = shared.ValuationComplete
			total = total.Add(value)
			valued++
		} else {
			unpriced++
		}
		item.Visibility = c.classifyVisibility(ctx, ast, item.ValueUSD)
		result.Positions = append(result.Positions, item)
	}

	if valued == 0 {
		result.ValuationStatus = shared.ValuationUnavailable
		result.DataQuality = shared.DataQualityUnavailable
		return result, nil
	}
	result.TotalValueUSD = shared.Known(shared.NewDecimal(total))
	if unpriced > 0 {
		result.ValuationStatus = shared.ValuationPartial
		result.DataQuality = shared.DataQualityPartial
		result.Exclusions.UnpricedPositions = unpriced
		result.Notices = append(result.Notices, portfolio.Notice{
			Code:     portfolio.NoticeUnpricedAssetsExcluded,
			Severity: portfolio.SeverityWarning,
			Params:   map[string]string{"count": decimal.NewFromInt(int64(unpriced)).String()},
		})
	} else {
		result.ValuationStatus = shared.ValuationComplete
		result.DataQuality = shared.DataQualityComplete
	}

	for i := range result.Positions {
		if !result.Positions[i].ValueUSD.Valid || total.IsZero() {
			continue
		}
		pct := result.Positions[i].ValueUSD.Decimal.Value().Div(total).Mul(decimal.NewFromInt(100))
		result.Positions[i].AllocationPct = shared.Known(shared.NewDecimal(pct))
	}
	return result, nil
}

func (c *Calculator) classifyVisibility(ctx context.Context, ast asset.Asset, valueUSD shared.NullDecimal) asset.Visibility {
	if c.visibility == nil {
		return asset.VisibilityVisible
	}
	input := asset.VisibilityInput{
		Asset:         ast,
		HasMarketData: ast.HasMarketData(),
	}
	if valueUSD.Valid {
		raw := valueUSD.Decimal.String()
		input.ValueUSD = &raw
	}
	visibility, err := c.visibility.Classify(ctx, input)
	if err != nil {
		return asset.VisibilityUnknown
	}
	return visibility
}

// SnapshotApp implements SnapshotService.
type SnapshotApp struct {
	calculator *Calculator
	snapshots  portfolio.SnapshotRepository
}

// NewSnapshotApp wires snapshot creation.
func NewSnapshotApp(calculator *Calculator, snapshots portfolio.SnapshotRepository) *SnapshotApp {
	return &SnapshotApp{calculator: calculator, snapshots: snapshots}
}

// Create implements SnapshotService.
func (a *SnapshotApp) Create(ctx context.Context, walletID uuid.UUID, syncRunID *uuid.UUID, capturedAt time.Time) (portfolio.Snapshot, error) {
	if syncRunID != nil {
		if existing, err := a.snapshots.GetBySyncRunID(ctx, *syncRunID); err == nil {
			return existing, nil
		} else if !apperr.Is(err, apperr.CodeNotFound) {
			return portfolio.Snapshot{}, err
		}
	}

	w := wallet.Wallet{ID: walletID}
	current, err := a.calculator.Build(ctx, w, &capturedAt)
	if err != nil {
		return portfolio.Snapshot{}, err
	}

	snapshot := portfolio.Snapshot{
		WalletID:           walletID,
		CapturedAt:         capturedAt,
		TotalValueUSD:      current.TotalValueUSD,
		Status:             current.ValuationStatus,
		DataQuality:        current.DataQuality,
		CalculationVersion: portfolio.CalculationVersion,
		SyncRunID:          syncRunID,
	}
	positions := make([]portfolio.SnapshotPosition, 0, len(current.Positions))
	for _, pos := range current.Positions {
		var priceUSD shared.NullDecimal
		var priceTS *time.Time
		var priceSource *price.Source
		if pos.Price != nil {
			priceUSD = pos.Price.ValueUSD
			priceTS = &pos.Price.AsOf
			source := pos.Price.Source
			priceSource = &source
		}
		positions = append(positions, portfolio.SnapshotPosition{
			AssetID:        pos.Asset.ID,
			Balance:        pos.Balance,
			PriceUSD:       priceUSD,
			ValueUSD:       pos.ValueUSD,
			AllocationPct:  pos.AllocationPct,
			PriceTimestamp: priceTS,
			PriceSource:    priceSource,
		})
	}
	return a.snapshots.Create(ctx, snapshot, positions)
}

var _ SnapshotService = (*SnapshotApp)(nil)

// ReadApp serves portfolio read requests.
type ReadApp struct {
	wallets    wallet.Repository
	syncStates wallet.SyncStateRepository
	calculator *Calculator
}

// NewReadApp wires the portfolio read service.
func NewReadApp(
	wallets wallet.Repository,
	syncStates wallet.SyncStateRepository,
	calculator *Calculator,
) *ReadApp {
	return &ReadApp{
		wallets:    wallets,
		syncStates: syncStates,
		calculator: calculator,
	}
}

// Get implements Service.
func (a *ReadApp) Get(ctx context.Context, userID, walletID uuid.UUID) (portfolio.Portfolio, error) {
	w, err := a.wallets.GetByID(ctx, walletID)
	if err != nil {
		if appErr := apperr.From(err); appErr != nil && appErr.Code == apperr.CodeNotFound {
			return portfolio.Portfolio{}, apperr.New(apperr.CodeWalletNotFound)
		}
		return portfolio.Portfolio{}, err
	}
	if w.UserID != userID {
		return portfolio.Portfolio{}, apperr.New(apperr.CodeWalletNotFound)
	}

	syncState, err := a.syncStates.Get(ctx, walletID)
	if err != nil {
		return portfolio.Portfolio{}, err
	}
	switch syncState.Status {
	case wallet.SyncPending, wallet.SyncSyncing:
		return portfolio.Portfolio{}, apperr.New(apperr.CodeWalletNotReady).
			WithDetail("sync_status", string(syncState.Status))
	case wallet.SyncFailed:
		return portfolio.Portfolio{}, apperr.New(apperr.CodeWalletSyncFailed)
	}

	return a.calculator.Build(ctx, w, syncState.LastSyncedAt)
}

var _ Service = (*ReadApp)(nil)
