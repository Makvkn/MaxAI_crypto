// Package pricerepo implements price persistence with sqlc.
package pricerepo

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/maxaicrypto/backend/internal/domain/price"
	"github.com/maxaicrypto/backend/internal/domain/shared"
	"github.com/maxaicrypto/backend/internal/generated/sqlc"
	"github.com/maxaicrypto/backend/internal/infrastructure/postgres"
)

// Repository implements price.Repository.
type Repository struct {
	pool *postgres.Pool
	tx   *postgres.TxRunner
}

// NewRepository builds a price repository.
func NewRepository(pool *postgres.Pool, tx *postgres.TxRunner) *Repository {
	return &Repository{pool: pool, tx: tx}
}

func (r *Repository) queries(ctx context.Context) *sqlc.Queries {
	if tx, ok := postgres.TxFrom(ctx); ok {
		return sqlc.New(tx)
	}
	return sqlc.New(r.pool)
}

// UpsertLatest implements price.Repository.
func (r *Repository) UpsertLatest(ctx context.Context, p price.Price) error {
	if err := r.queries(ctx).UpsertPrice(ctx, sqlc.UpsertPriceParams{
		AssetID:      p.AssetID,
		AsOf:         p.AsOf,
		Currency:     string(p.Currency),
		ValueUsd:     toSQLNullDecimal(p.ValueUSD),
		Status:       string(p.Status),
		Source:       string(p.Source),
		Change24hPct: toSQLNullDecimal(p.Change24h),
	}); err != nil {
		return postgres.MapError(err)
	}
	return nil
}

// GetLatest implements price.Repository.
func (r *Repository) GetLatest(ctx context.Context, assetIDs []uuid.UUID) (map[uuid.UUID]price.Price, error) {
	if len(assetIDs) == 0 {
		return map[uuid.UUID]price.Price{}, nil
	}
	rows, err := r.queries(ctx).GetLatestPrices(ctx, assetIDs)
	if err != nil {
		return nil, postgres.MapError(err)
	}
	prices := make(map[uuid.UUID]price.Price, len(rows))
	for _, row := range rows {
		p := mapPrice(row)
		prices[p.AssetID] = p
	}
	return prices, nil
}

// GetClosest implements price.Repository.
func (r *Repository) GetClosest(ctx context.Context, assetID uuid.UUID, at time.Time) (price.Price, bool, error) {
	row, err := r.queries(ctx).GetClosestPrice(ctx, sqlc.GetClosestPriceParams{
		AssetID: assetID,
		Column2: at,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return price.Price{}, false, nil
		}
		return price.Price{}, false, postgres.MapError(err)
	}
	return mapPrice(row), true, nil
}

func mapPrice(row sqlc.Price) price.Price {
	return price.Price{
		AssetID:   row.AssetID,
		Currency:  shared.Currency(row.Currency),
		ValueUSD:  fromSQLNullDecimal(row.ValueUsd),
		Status:    price.Status(row.Status),
		Source:    price.Source(row.Source),
		AsOf:      row.AsOf,
		Change24h: fromSQLNullDecimal(row.Change24hPct),
		CreatedAt: row.CreatedAt,
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

var _ price.Repository = (*Repository)(nil)
