package portfolio

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/maxaicrypto/backend/internal/domain/asset"
	"github.com/maxaicrypto/backend/internal/domain/position"
	"github.com/maxaicrypto/backend/internal/domain/price"
	"github.com/maxaicrypto/backend/internal/domain/shared"
	"github.com/maxaicrypto/backend/internal/domain/wallet"
)

type emptyPositionsRepo struct{}

func (emptyPositionsRepo) ListByWallet(context.Context, uuid.UUID) ([]position.WalletPosition, error) {
	return nil, nil
}
func (emptyPositionsRepo) GetByAsset(context.Context, uuid.UUID, uuid.UUID) (position.WalletPosition, bool, error) {
	panic("unexpected")
}
func (emptyPositionsRepo) ReplaceForWallet(context.Context, uuid.UUID, []position.WalletPosition, time.Time) error {
	panic("unexpected")
}

type unusedAssetsRepo struct{}

func (unusedAssetsRepo) GetByID(context.Context, uuid.UUID) (asset.Asset, error) {
	panic("unexpected")
}
func (unusedAssetsRepo) GetManyByID(context.Context, []uuid.UUID) (map[uuid.UUID]asset.Asset, error) {
	panic("unexpected")
}
func (unusedAssetsRepo) FindByIdentity(context.Context, asset.Identity) (asset.Asset, bool, error) {
	panic("unexpected")
}
func (unusedAssetsRepo) Upsert(context.Context, asset.Asset) (asset.Asset, error) {
	panic("unexpected")
}
func (unusedAssetsRepo) SetMarketDataMapping(context.Context, uuid.UUID, *asset.MarketDataProvider, *string) error {
	panic("unexpected")
}
func (unusedAssetsRepo) ListUnmapped(context.Context, int) ([]asset.Asset, error) {
	panic("unexpected")
}

type unusedPricing struct{}

func (unusedPricing) GetCurrent(context.Context, []uuid.UUID) (map[uuid.UUID]price.Price, error) {
	panic("unexpected")
}
func (unusedPricing) GetAt(context.Context, uuid.UUID, time.Time) (price.Price, bool, error) {
	panic("unexpected")
}
func (unusedPricing) Refresh(context.Context, []uuid.UUID) (int, error) {
	panic("unexpected")
}
func (unusedPricing) ResolveMapping(context.Context, asset.Asset) (bool, error) {
	panic("unexpected")
}

func TestCalculatorBuildEmptyWalletIsZero(t *testing.T) {
	t.Parallel()

	calc := NewCalculator(
		unusedAssetsRepo{},
		emptyPositionsRepo{},
		unusedPricing{},
		shared.FreshnessThresholds{
			FreshMax:  5 * time.Minute,
			RecentMax: 15 * time.Minute,
			StaleMax:  60 * time.Minute,
		},
	)
	syncedAt := time.Now().UTC().Add(-2 * time.Minute)
	got, err := calc.Build(context.Background(), wallet.Wallet{ID: uuid.New()}, &syncedAt)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if !got.TotalValueUSD.Valid {
		t.Fatal("expected total_value_usd to be known")
	}
	if !got.TotalValueUSD.Decimal.Value().IsZero() {
		t.Fatalf("total_value_usd = %s, want 0", got.TotalValueUSD.Decimal)
	}
	if got.ValuationStatus != shared.ValuationComplete {
		t.Fatalf("valuation_status = %s, want COMPLETE", got.ValuationStatus)
	}
	if got.DataQuality != shared.DataQualityComplete {
		t.Fatalf("data_quality = %s, want COMPLETE", got.DataQuality)
	}
	if len(got.Positions) != 0 {
		t.Fatalf("positions = %d, want 0", len(got.Positions))
	}
	if !got.Change24hUSD.Valid || !got.Change24hUSD.Decimal.Value().IsZero() {
		t.Fatalf("change_24h_usd = %#v, want known 0", got.Change24hUSD)
	}
	if !got.Change24hPct.Valid || !got.Change24hPct.Decimal.Value().IsZero() {
		t.Fatalf("change_24h_pct = %#v, want known 0", got.Change24hPct)
	}
}
