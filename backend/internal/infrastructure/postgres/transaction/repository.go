// Package transactionrepo implements transaction persistence with sqlc.
package transactionrepo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/maxaicrypto/backend/internal/domain/chain"
	"github.com/maxaicrypto/backend/internal/domain/shared"
	"github.com/maxaicrypto/backend/internal/domain/transaction"
	"github.com/maxaicrypto/backend/internal/generated/sqlc"
	"github.com/maxaicrypto/backend/internal/infrastructure/postgres"
)

const listTransactionsByWalletSQL = `
SELECT id, wallet_id, chain_id, tx_hash, log_index, block_number, timestamp, status, type,
       from_address, to_address, asset_in_id, amount_in, asset_out_id, amount_out,
       fee_asset_id, fee_amount, protocol, counterparty, raw_reference, created_at, updated_at
FROM transactions
WHERE wallet_id = $1
  AND ($2::text IS NULL OR type = $2)
  AND (
    $4::timestamptz IS NULL
    OR (timestamp, id) < ($4::timestamptz, $5::uuid)
  )
ORDER BY timestamp DESC, id DESC
LIMIT $3`

// Repository implements transaction.Repository.
type Repository struct {
	pool *postgres.Pool
	tx   *postgres.TxRunner
}

// NewRepository builds a transaction repository.
func NewRepository(pool *postgres.Pool, tx *postgres.TxRunner) *Repository {
	return &Repository{pool: pool, tx: tx}
}

func (r *Repository) db(ctx context.Context) postgres.DBTX {
	if tx, ok := postgres.TxFrom(ctx); ok {
		return tx
	}
	return r.pool
}

func (r *Repository) queries(ctx context.Context) *sqlc.Queries {
	return sqlc.New(r.db(ctx))
}

// ListByWallet implements transaction.Repository.
func (r *Repository) ListByWallet(ctx context.Context, walletID uuid.UUID, filter transaction.Filter, page shared.Cursor, limit int) ([]transaction.Transaction, error) {
	var typeFilter *string
	if filter.Type != nil {
		s := string(*filter.Type)
		typeFilter = &s
	}

	var cursorAt *time.Time
	var cursorID uuid.NullUUID
	if !page.IsZero() {
		cursorIDParsed, err := uuid.Parse(page.TieBreaker)
		if err != nil {
			return nil, err
		}
		at := page.SortKey
		cursorAt = &at
		cursorID = uuid.NullUUID{UUID: cursorIDParsed, Valid: true}
	}

	rows, err := r.db(ctx).Query(ctx, listTransactionsByWalletSQL,
		walletID,
		typeFilter,
		limit,
		cursorAt,
		cursorID,
	)
	if err != nil {
		return nil, postgres.MapError(err)
	}
	defer rows.Close()

	transactions := make([]transaction.Transaction, 0, limit)
	for rows.Next() {
		var row sqlc.Transaction
		if err := rows.Scan(
			&row.ID,
			&row.WalletID,
			&row.ChainID,
			&row.TxHash,
			&row.LogIndex,
			&row.BlockNumber,
			&row.Timestamp,
			&row.Status,
			&row.Type,
			&row.FromAddress,
			&row.ToAddress,
			&row.AssetInID,
			&row.AmountIn,
			&row.AssetOutID,
			&row.AmountOut,
			&row.FeeAssetID,
			&row.FeeAmount,
			&row.Protocol,
			&row.Counterparty,
			&row.RawReference,
			&row.CreatedAt,
			&row.UpdatedAt,
		); err != nil {
			return nil, postgres.MapError(err)
		}
		transactions = append(transactions, mapTransaction(row))
	}
	if err := rows.Err(); err != nil {
		return nil, postgres.MapError(err)
	}
	return transactions, nil
}

// GetByID implements transaction.Repository.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (transaction.Transaction, error) {
	row, err := r.queries(ctx).GetTransactionByID(ctx, id)
	if err != nil {
		return transaction.Transaction{}, postgres.MapError(err)
	}
	return mapTransaction(row), nil
}

// UpsertBatch implements transaction.Repository.
func (r *Repository) UpsertBatch(ctx context.Context, transactions []transaction.Transaction) (int, error) {
	if len(transactions) == 0 {
		return 0, nil
	}

	q := r.queries(ctx)
	count := 0
	for _, tx := range transactions {
		if _, err := q.UpsertTransaction(ctx, toUpsertParams(tx)); err != nil {
			return count, postgres.MapError(err)
		}
		count++
	}
	return count, nil
}

func toUpsertParams(tx transaction.Transaction) sqlc.UpsertTransactionParams {
	return sqlc.UpsertTransactionParams{
		WalletID:     tx.WalletID,
		ChainID:      string(tx.ChainID),
		TxHash:       tx.TxHash,
		LogIndex:     0,
		BlockNumber:  tx.BlockNumber,
		Timestamp:    tx.Timestamp,
		Status:       string(tx.Status),
		Type:         string(tx.Type),
		FromAddress:  tx.FromAddress,
		ToAddress:    tx.ToAddress,
		AssetInID:    toNullUUID(tx.AssetInID),
		AmountIn:     toSQLNullDecimal(tx.AmountIn),
		AssetOutID:   toNullUUID(tx.AssetOutID),
		AmountOut:    toSQLNullDecimal(tx.AmountOut),
		FeeAssetID:   toNullUUID(tx.FeeAssetID),
		FeeAmount:    toSQLNullDecimal(tx.FeeAmount),
		Protocol:     tx.Protocol,
		Counterparty: tx.Counterparty,
		RawReference: tx.RawReference,
	}
}

func mapTransaction(row sqlc.Transaction) transaction.Transaction {
	return transaction.Transaction{
		ID:           row.ID,
		WalletID:     row.WalletID,
		ChainID:      chain.ID(row.ChainID),
		TxHash:       row.TxHash,
		BlockNumber:  row.BlockNumber,
		Timestamp:    row.Timestamp,
		Status:       transaction.Status(row.Status),
		Type:         transaction.Type(row.Type),
		FromAddress:  row.FromAddress,
		ToAddress:    row.ToAddress,
		AssetInID:    fromNullUUID(row.AssetInID),
		AmountIn:     fromSQLNullDecimal(row.AmountIn),
		AssetOutID:   fromNullUUID(row.AssetOutID),
		AmountOut:    fromSQLNullDecimal(row.AmountOut),
		FeeAssetID:   fromNullUUID(row.FeeAssetID),
		FeeAmount:    fromSQLNullDecimal(row.FeeAmount),
		Protocol:     row.Protocol,
		Counterparty: row.Counterparty,
		RawReference: row.RawReference,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

func toNullUUID(id *uuid.UUID) uuid.NullUUID {
	if id == nil {
		return uuid.NullUUID{}
	}
	return uuid.NullUUID{UUID: *id, Valid: true}
}

func fromNullUUID(id uuid.NullUUID) *uuid.UUID {
	if !id.Valid {
		return nil
	}
	value := id.UUID
	return &value
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

var _ transaction.Repository = (*Repository)(nil)
