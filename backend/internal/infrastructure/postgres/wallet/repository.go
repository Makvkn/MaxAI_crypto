// Package walletrepo implements wallet persistence with sqlc.
package walletrepo

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/maxaicrypto/backend/internal/domain/chain"
	"github.com/maxaicrypto/backend/internal/domain/shared"
	"github.com/maxaicrypto/backend/internal/domain/wallet"
	"github.com/maxaicrypto/backend/internal/generated/sqlc"
	"github.com/maxaicrypto/backend/internal/infrastructure/postgres"
)

// Repository implements wallet.Repository.
type Repository struct {
	pool *postgres.Pool
	tx   *postgres.TxRunner
}

// NewRepository builds a wallet repository.
func NewRepository(pool *postgres.Pool, tx *postgres.TxRunner) *Repository {
	return &Repository{pool: pool, tx: tx}
}

func (r *Repository) queries(ctx context.Context) *sqlc.Queries {
	if tx, ok := postgres.TxFrom(ctx); ok {
		return sqlc.New(tx)
	}
	return sqlc.New(r.pool)
}

// Create implements wallet.Repository.
func (r *Repository) Create(ctx context.Context, w wallet.Wallet) (wallet.Wallet, error) {
	row, err := r.queries(ctx).CreateWallet(ctx, sqlc.CreateWalletParams{
		UserID:  w.UserID,
		ChainID: string(w.ChainID),
		Address: w.Address,
		Label:   w.Label,
	})
	if err != nil {
		return wallet.Wallet{}, postgres.MapError(err)
	}
	return mapWallet(row), nil
}

// GetByID implements wallet.Repository.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (wallet.Wallet, error) {
	row, err := r.queries(ctx).GetWalletByID(ctx, id)
	if err != nil {
		return wallet.Wallet{}, postgres.MapError(err)
	}
	return mapWallet(row), nil
}

// ListByUser implements wallet.Repository.
func (r *Repository) ListByUser(ctx context.Context, userID uuid.UUID, page shared.Cursor, limit int) ([]wallet.Wallet, error) {
	params := sqlc.ListWalletsByUserParams{
		UserID: userID,
		Limit:  int32(limit),
	}
	if !page.IsZero() {
		cursorID, err := uuid.Parse(page.TieBreaker)
		if err != nil {
			return nil, err
		}
		params.CursorAt = &page.SortKey
		params.CursorID = uuid.NullUUID{UUID: cursorID, Valid: true}
	}

	rows, err := r.queries(ctx).ListWalletsByUser(ctx, params)
	if err != nil {
		return nil, postgres.MapError(err)
	}
	wallets := make([]wallet.Wallet, len(rows))
	for i, row := range rows {
		wallets[i] = mapWallet(row)
	}
	return wallets, nil
}

// FindByAddress implements wallet.Repository.
func (r *Repository) FindByAddress(ctx context.Context, userID uuid.UUID, chainID chain.ID, address string) (wallet.Wallet, bool, error) {
	row, err := r.queries(ctx).FindWalletByUserChainAddress(ctx, sqlc.FindWalletByUserChainAddressParams{
		UserID:  userID,
		ChainID: string(chainID),
		Address: address,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return wallet.Wallet{}, false, nil
		}
		return wallet.Wallet{}, false, postgres.MapError(err)
	}
	return mapWallet(row), true, nil
}

// CountByUser implements wallet.Repository.
func (r *Repository) CountByUser(ctx context.Context, userID uuid.UUID) (int, error) {
	count, err := r.queries(ctx).CountWalletsByUser(ctx, userID)
	if err != nil {
		return 0, postgres.MapError(err)
	}
	return int(count), nil
}

// SoftDelete implements wallet.Repository.
func (r *Repository) SoftDelete(ctx context.Context, id uuid.UUID, at time.Time) error {
	if err := r.queries(ctx).SoftDeleteWallet(ctx, sqlc.SoftDeleteWalletParams{
		ID:        id,
		DeletedAt: &at,
	}); err != nil {
		return postgres.MapError(err)
	}
	return nil
}

// UpdateStatus implements wallet.Repository.
func (r *Repository) UpdateStatus(ctx context.Context, id uuid.UUID, status wallet.Status) error {
	if err := r.queries(ctx).UpdateWalletStatus(ctx, sqlc.UpdateWalletStatusParams{
		ID:     id,
		Status: string(status),
	}); err != nil {
		return postgres.MapError(err)
	}
	return nil
}

// ListDueForSync implements wallet.Repository.
func (r *Repository) ListDueForSync(ctx context.Context, olderThan time.Time, limit int) ([]wallet.Wallet, error) {
	rows, err := r.queries(ctx).ListWalletsDueForSync(ctx, sqlc.ListWalletsDueForSyncParams{
		LastSyncedAt: &olderThan,
		Limit:        int32(limit),
	})
	if err != nil {
		return nil, postgres.MapError(err)
	}
	wallets := make([]wallet.Wallet, len(rows))
	for i, row := range rows {
		wallets[i] = mapWallet(row)
	}
	return wallets, nil
}

func mapWallet(row sqlc.Wallet) wallet.Wallet {
	return wallet.Wallet{
		ID:        row.ID,
		UserID:    row.UserID,
		ChainID:   chain.ID(row.ChainID),
		Address:   row.Address,
		Label:     row.Label,
		Status:    wallet.Status(row.Status),
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
		DeletedAt: row.DeletedAt,
	}
}

var _ wallet.Repository = (*Repository)(nil)
