// Package portfoliorepo implements portfolio snapshot persistence with sqlc.
package portfoliorepo

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/maxaicrypto/backend/internal/domain/portfolio"
	"github.com/maxaicrypto/backend/internal/domain/price"
	"github.com/maxaicrypto/backend/internal/domain/shared"
	"github.com/maxaicrypto/backend/internal/generated/sqlc"
	"github.com/maxaicrypto/backend/internal/infrastructure/postgres"
)

const (
	getClosestValidBeforeSQL = `
SELECT id, wallet_id, captured_at, total_value_usd, status, data_quality,
       calculation_version, sync_run_id, created_at
FROM portfolio_snapshots
WHERE wallet_id = $1
  AND status <> 'UNAVAILABLE'
  AND total_value_usd IS NOT NULL
  AND captured_at <= $2
ORDER BY captured_at DESC
LIMIT 1`

	getFirstValidSQL = `
SELECT id, wallet_id, captured_at, total_value_usd, status, data_quality,
       calculation_version, sync_run_id, created_at
FROM portfolio_snapshots
WHERE wallet_id = $1
  AND status <> 'UNAVAILABLE'
  AND total_value_usd IS NOT NULL
ORDER BY captured_at ASC
LIMIT 1`

	listBetweenSQL = `
SELECT id, wallet_id, captured_at, total_value_usd, status, data_quality,
       calculation_version, sync_run_id, created_at
FROM portfolio_snapshots
WHERE wallet_id = $1
  AND captured_at >= $2
  AND captured_at <= $3
ORDER BY captured_at ASC
LIMIT $4`

	listSnapshotPositionsSQL = `
SELECT snapshot_id, asset_id, balance, price_usd, value_usd, allocation_pct,
       price_timestamp, price_source
FROM portfolio_snapshot_positions
WHERE snapshot_id = $1`
)

// SnapshotRepository implements portfolio.SnapshotRepository.
type SnapshotRepository struct {
	pool *postgres.Pool
	tx   *postgres.TxRunner
}

// NewSnapshotRepository builds a portfolio snapshot repository.
func NewSnapshotRepository(pool *postgres.Pool, tx *postgres.TxRunner) *SnapshotRepository {
	return &SnapshotRepository{pool: pool, tx: tx}
}

func (r *SnapshotRepository) db(ctx context.Context) postgres.DBTX {
	if tx, ok := postgres.TxFrom(ctx); ok {
		return tx
	}
	return r.pool
}

func (r *SnapshotRepository) queries(ctx context.Context) *sqlc.Queries {
	return sqlc.New(r.db(ctx))
}

// Create implements portfolio.SnapshotRepository.
func (r *SnapshotRepository) Create(ctx context.Context, snapshot portfolio.Snapshot, positions []portfolio.SnapshotPosition) (portfolio.Snapshot, error) {
	var created portfolio.Snapshot
	err := r.tx.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		q := sqlc.New(tx)
		row, err := q.CreatePortfolioSnapshot(ctx, sqlc.CreatePortfolioSnapshotParams{
			WalletID:           snapshot.WalletID,
			CapturedAt:         snapshot.CapturedAt,
			TotalValueUsd:      toSQLNullDecimal(snapshot.TotalValueUSD),
			Status:             string(snapshot.Status),
			DataQuality:        string(snapshot.DataQuality),
			CalculationVersion: int32(snapshot.CalculationVersion),
			SyncRunID:          toNullUUID(snapshot.SyncRunID),
		})
		if err != nil {
			return postgres.MapError(err)
		}
		created = mapSnapshot(row)

		for _, position := range positions {
			var priceSource *string
			if position.PriceSource != nil {
				s := string(*position.PriceSource)
				priceSource = &s
			}
			if err := q.InsertSnapshotPosition(ctx, sqlc.InsertSnapshotPositionParams{
				SnapshotID:     created.ID,
				AssetID:        position.AssetID,
				Balance:        position.Balance.Value(),
				PriceUsd:       toSQLNullDecimal(position.PriceUSD),
				ValueUsd:       toSQLNullDecimal(position.ValueUSD),
				AllocationPct:  toSQLNullDecimal(position.AllocationPct),
				PriceTimestamp: position.PriceTimestamp,
				PriceSource:    priceSource,
			}); err != nil {
				return postgres.MapError(err)
			}
		}
		return nil
	})
	return created, err
}

// GetLatestValid implements portfolio.SnapshotRepository.
func (r *SnapshotRepository) GetLatestValid(ctx context.Context, walletID uuid.UUID) (portfolio.Snapshot, bool, error) {
	row, err := r.queries(ctx).GetLatestValidSnapshot(ctx, walletID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return portfolio.Snapshot{}, false, nil
		}
		return portfolio.Snapshot{}, false, postgres.MapError(err)
	}
	return mapSnapshot(row), true, nil
}

// GetBySyncRunID implements portfolio.SnapshotRepository.
func (r *SnapshotRepository) GetBySyncRunID(ctx context.Context, syncRunID uuid.UUID) (portfolio.Snapshot, error) {
	row, err := r.queries(ctx).GetSnapshotBySyncRunID(ctx, uuid.NullUUID{UUID: syncRunID, Valid: true})
	if err != nil {
		return portfolio.Snapshot{}, postgres.MapError(err)
	}
	return mapSnapshot(row), nil
}

// GetClosestValidBefore implements portfolio.SnapshotRepository.
func (r *SnapshotRepository) GetClosestValidBefore(ctx context.Context, walletID uuid.UUID, at time.Time) (portfolio.Snapshot, bool, error) {
	row := r.db(ctx).QueryRow(ctx, getClosestValidBeforeSQL, walletID, at)
	snapshot, err := scanSnapshot(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return portfolio.Snapshot{}, false, nil
		}
		return portfolio.Snapshot{}, false, postgres.MapError(err)
	}
	return snapshot, true, nil
}

// GetFirstValid implements portfolio.SnapshotRepository.
func (r *SnapshotRepository) GetFirstValid(ctx context.Context, walletID uuid.UUID) (portfolio.Snapshot, bool, error) {
	row := r.db(ctx).QueryRow(ctx, getFirstValidSQL, walletID)
	snapshot, err := scanSnapshot(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return portfolio.Snapshot{}, false, nil
		}
		return portfolio.Snapshot{}, false, postgres.MapError(err)
	}
	return snapshot, true, nil
}

// ListBetween implements portfolio.SnapshotRepository.
func (r *SnapshotRepository) ListBetween(ctx context.Context, walletID uuid.UUID, from, to time.Time, limit int) ([]portfolio.Snapshot, error) {
	rows, err := r.db(ctx).Query(ctx, listBetweenSQL, walletID, from, to, limit)
	if err != nil {
		return nil, postgres.MapError(err)
	}
	defer rows.Close()

	snapshots := make([]portfolio.Snapshot, 0, limit)
	for rows.Next() {
		snapshot, err := scanSnapshot(rows)
		if err != nil {
			return nil, postgres.MapError(err)
		}
		snapshots = append(snapshots, snapshot)
	}
	if err := rows.Err(); err != nil {
		return nil, postgres.MapError(err)
	}
	return snapshots, nil
}

// ListPositions implements portfolio.SnapshotRepository.
func (r *SnapshotRepository) ListPositions(ctx context.Context, snapshotID uuid.UUID) ([]portfolio.SnapshotPosition, error) {
	rows, err := r.db(ctx).Query(ctx, listSnapshotPositionsSQL, snapshotID)
	if err != nil {
		return nil, postgres.MapError(err)
	}
	defer rows.Close()

	positions := make([]portfolio.SnapshotPosition, 0)
	for rows.Next() {
		position, err := scanSnapshotPosition(rows)
		if err != nil {
			return nil, err
		}
		positions = append(positions, position)
	}
	if err := rows.Err(); err != nil {
		return nil, postgres.MapError(err)
	}
	return positions, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanSnapshot(row rowScanner) (portfolio.Snapshot, error) {
	var dbRow sqlc.PortfolioSnapshot
	if err := row.Scan(
		&dbRow.ID,
		&dbRow.WalletID,
		&dbRow.CapturedAt,
		&dbRow.TotalValueUsd,
		&dbRow.Status,
		&dbRow.DataQuality,
		&dbRow.CalculationVersion,
		&dbRow.SyncRunID,
		&dbRow.CreatedAt,
	); err != nil {
		return portfolio.Snapshot{}, err
	}
	return mapSnapshot(dbRow), nil
}

func scanSnapshotPosition(row rowScanner) (portfolio.SnapshotPosition, error) {
	var (
		snapshotID     uuid.UUID
		assetID        uuid.UUID
		balance        decimal.Decimal
		priceUSD       decimal.NullDecimal
		valueUSD       decimal.NullDecimal
		allocationPct  decimal.NullDecimal
		priceTimestamp *time.Time
		priceSource    *string
	)
	if err := row.Scan(
		&snapshotID,
		&assetID,
		&balance,
		&priceUSD,
		&valueUSD,
		&allocationPct,
		&priceTimestamp,
		&priceSource,
	); err != nil {
		return portfolio.SnapshotPosition{}, err
	}

	position := portfolio.SnapshotPosition{
		SnapshotID:     snapshotID,
		AssetID:        assetID,
		Balance:        shared.NewDecimal(balance),
		PriceUSD:       fromSQLNullDecimal(priceUSD),
		ValueUSD:       fromSQLNullDecimal(valueUSD),
		AllocationPct:  fromSQLNullDecimal(allocationPct),
		PriceTimestamp: priceTimestamp,
	}
	if priceSource != nil {
		source := price.Source(*priceSource)
		position.PriceSource = &source
	}
	return position, nil
}

func mapSnapshot(row sqlc.PortfolioSnapshot) portfolio.Snapshot {
	var syncRunID *uuid.UUID
	if row.SyncRunID.Valid {
		syncRunID = &row.SyncRunID.UUID
	}
	return portfolio.Snapshot{
		ID:                 row.ID,
		WalletID:           row.WalletID,
		CapturedAt:         row.CapturedAt,
		TotalValueUSD:      fromSQLNullDecimal(row.TotalValueUsd),
		Status:             shared.ValuationStatus(row.Status),
		DataQuality:        shared.DataQuality(row.DataQuality),
		CalculationVersion: int(row.CalculationVersion),
		SyncRunID:          syncRunID,
		CreatedAt:          row.CreatedAt,
	}
}

func toNullUUID(id *uuid.UUID) uuid.NullUUID {
	if id == nil {
		return uuid.NullUUID{}
	}
	return uuid.NullUUID{UUID: *id, Valid: true}
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

var _ portfolio.SnapshotRepository = (*SnapshotRepository)(nil)
