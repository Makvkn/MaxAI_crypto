// Package scenariorepo implements scenario calculation persistence with sqlc.
package scenariorepo

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/maxaicrypto/backend/internal/domain/scenario"
	"github.com/maxaicrypto/backend/internal/domain/shared"
	"github.com/maxaicrypto/backend/internal/generated/sqlc"
	"github.com/maxaicrypto/backend/internal/infrastructure/postgres"
)

// Repository implements scenario.Repository.
type Repository struct {
	pool *postgres.Pool
}

// NewRepository builds a scenario repository.
func NewRepository(pool *postgres.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) queries(ctx context.Context) *sqlc.Queries {
	if tx, ok := postgres.TxFrom(ctx); ok {
		return sqlc.New(tx)
	}
	return sqlc.New(r.pool)
}

// Create implements scenario.Repository.
func (r *Repository) Create(ctx context.Context, result scenario.Result) (scenario.Result, error) {
	row, err := r.queries(ctx).CreateScenarioCalculation(ctx, sqlc.CreateScenarioCalculationParams{
		UserID:                     result.UserID,
		WalletID:                   result.WalletID,
		Type:                       string(result.Type),
		AssetID:                    result.AssetID,
		ChangePct:                  result.ChangePct.Value(),
		BaselinePortfolioValueUsd:  toSQLNullDecimal(result.Baseline.PortfolioValueUSD),
		BaselineAssetValueUsd:      toSQLNullDecimal(result.Baseline.AssetValueUSD),
		BaselineAssetAllocationPct: toSQLNullDecimal(result.Baseline.AssetAllocationPct),
		ProjectedPortfolioValueUsd: toSQLNullDecimal(result.Projection.PortfolioValueUSD),
		ProjectedAssetValueUsd:     toSQLNullDecimal(result.Projection.AssetValueUSD),
		AssetImpactUsd:             toSQLNullDecimal(result.Projection.AssetImpactUSD),
		PortfolioChangeUsd:         toSQLNullDecimal(result.Projection.PortfolioChangeUSD),
		PortfolioChangePct:         toSQLNullDecimal(result.Projection.PortfolioChangePct),
		DataQuality:                string(result.DataQuality),
		CalculationVersion:         int32(result.CalculationVersion),
	})
	if err != nil {
		return scenario.Result{}, postgres.MapError(err)
	}
	return mapRow(row), nil
}

// GetByID implements scenario.Repository.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (scenario.Result, error) {
	row, err := r.queries(ctx).GetScenarioCalculationByID(ctx, id)
	if err != nil {
		return scenario.Result{}, postgres.MapError(err)
	}
	return mapRow(row), nil
}

func mapRow(row sqlc.ScenarioCalculation) scenario.Result {
	return scenario.Result{
		ID:       row.ID,
		UserID:   row.UserID,
		WalletID: row.WalletID,
		Type:     scenario.Type(row.Type),
		Currency: shared.CurrencyUSD,
		AssetID:  row.AssetID,
		ChangePct: shared.NewDecimal(row.ChangePct),
		Baseline: scenario.Baseline{
			PortfolioValueUSD:  fromSQLNullDecimal(row.BaselinePortfolioValueUsd),
			AssetValueUSD:      fromSQLNullDecimal(row.BaselineAssetValueUsd),
			AssetAllocationPct: fromSQLNullDecimal(row.BaselineAssetAllocationPct),
		},
		Projection: scenario.Projection{
			PortfolioValueUSD:  fromSQLNullDecimal(row.ProjectedPortfolioValueUsd),
			AssetValueUSD:      fromSQLNullDecimal(row.ProjectedAssetValueUsd),
			AssetImpactUSD:     fromSQLNullDecimal(row.AssetImpactUsd),
			PortfolioChangeUSD: fromSQLNullDecimal(row.PortfolioChangeUsd),
			PortfolioChangePct: fromSQLNullDecimal(row.PortfolioChangePct),
		},
		DataQuality:        shared.DataQuality(row.DataQuality),
		CalculationID:      row.ID,
		CalculationVersion: int(row.CalculationVersion),
		CreatedAt:          row.CreatedAt,
	}
}

func toSQLNullDecimal(value shared.NullDecimal) decimal.NullDecimal {
	if !value.Valid {
		return decimal.NullDecimal{}
	}
	return decimal.NullDecimal{Decimal: value.Decimal.Value(), Valid: true}
}

func fromSQLNullDecimal(value decimal.NullDecimal) shared.NullDecimal {
	if !value.Valid {
		return shared.Unknown()
	}
	return shared.Known(shared.NewDecimal(value.Decimal))
}

var _ scenario.Repository = (*Repository)(nil)
