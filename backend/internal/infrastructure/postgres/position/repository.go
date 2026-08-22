// Package positionrepo implements wallet position persistence with sqlc.
package positionrepo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/maxaicrypto/backend/internal/domain/position"
	"github.com/maxaicrypto/backend/internal/domain/shared"
	"github.com/maxaicrypto/backend/internal/generated/sqlc"
	"github.com/maxaicrypto/backend/internal/infrastructure/postgres"
)

// Repository implements position.Repository.
type Repository struct {
	pool *postgres.Pool
	tx   *postgres.TxRunner
}

// NewRepository builds a position repository.
func NewRepository(pool *postgres.Pool, tx *postgres.TxRunner) *Repository {
	return &Repository{pool: pool, tx: tx}
}

func (r *Repository) queries(ctx context.Context) *sqlc.Queries {
	if tx, ok := postgres.TxFrom(ctx); ok {
		return sqlc.New(tx)
	}
	return sqlc.New(r.pool)
}

// ListByWallet implements position.Repository.
func (r *Repository) ListByWallet(ctx context.Context, walletID uuid.UUID) ([]position.WalletPosition, error) {
	rows, err := r.queries(ctx).ListPositionsByWallet(ctx, walletID)
	if err != nil {
		return nil, postgres.MapError(err)
	}
	positions := make([]position.WalletPosition, len(rows))
	for i, row := range rows {
		p, err := mapPosition(row)
		if err != nil {
			return nil, err
		}
		positions[i] = p
	}
	return positions, nil
}

// GetByAsset implements position.Repository.
func (r *Repository) GetByAsset(ctx context.Context, walletID, assetID uuid.UUID) (position.WalletPosition, bool, error) {
	row, err := r.queries(ctx).GetPositionByAsset(ctx, sqlc.GetPositionByAssetParams{
		WalletID: walletID,
		AssetID:  assetID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return position.WalletPosition{}, false, nil
		}
		return position.WalletPosition{}, false, postgres.MapError(err)
	}
	p, err := mapPosition(row)
	if err != nil {
		return position.WalletPosition{}, false, err
	}
	return p, true, nil
}

// ReplaceForWallet implements position.Repository.
func (r *Repository) ReplaceForWallet(ctx context.Context, walletID uuid.UUID, positions []position.WalletPosition, observedAt time.Time) error {
	return r.tx.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		q := sqlc.New(tx)
		if err := q.DeletePositionsByWallet(ctx, walletID); err != nil {
			return postgres.MapError(err)
		}
		for _, p := range positions {
			balanceRaw, err := decimal.NewFromString(p.BalanceRaw)
			if err != nil {
				return fmt.Errorf("parse balance_raw for asset %s: %w", p.AssetID, err)
			}
			updatedAt := p.UpdatedAt
			if updatedAt.IsZero() {
				updatedAt = observedAt
			}
			if err := q.InsertPosition(ctx, sqlc.InsertPositionParams{
				WalletID:          walletID,
				AssetID:           p.AssetID,
				BalanceRaw:        balanceRaw,
				BalanceNormalized: p.BalanceNormalized.Value(),
				UpdatedAt:         updatedAt,
			}); err != nil {
				return postgres.MapError(err)
			}
		}
		return nil
	})
}

func mapPosition(row sqlc.WalletPosition) (position.WalletPosition, error) {
	balanceNormalized, err := shared.ParseDecimal(row.BalanceNormalized.String())
	if err != nil {
		return position.WalletPosition{}, fmt.Errorf("parse balance_normalized: %w", err)
	}
	return position.WalletPosition{
		WalletID:          row.WalletID,
		AssetID:           row.AssetID,
		BalanceRaw:        row.BalanceRaw.StringFixed(0),
		BalanceNormalized: balanceNormalized,
		UpdatedAt:         row.UpdatedAt,
	}, nil
}

var _ position.Repository = (*Repository)(nil)
