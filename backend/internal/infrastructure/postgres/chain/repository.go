// Package chainrepo implements chain persistence with sqlc.
package chainrepo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/google/uuid"

	"github.com/maxaicrypto/backend/internal/domain/apperr"
	"github.com/maxaicrypto/backend/internal/domain/chain"
	"github.com/maxaicrypto/backend/internal/generated/sqlc"
	"github.com/maxaicrypto/backend/internal/infrastructure/postgres"
)

// Repository implements chain.Repository.
type Repository struct {
	pool *postgres.Pool
}

// NewRepository builds a chain repository.
func NewRepository(pool *postgres.Pool) *Repository {
	return &Repository{pool: pool}
}

// List implements chain.Repository.
func (r *Repository) List(ctx context.Context) ([]chain.Chain, error) {
	rows, err := sqlc.New(r.pool).ListChains(ctx)
	if err != nil {
		return nil, postgres.MapError(err)
	}
	chains := make([]chain.Chain, len(rows))
	for i, row := range rows {
		chains[i] = mapChain(row)
	}
	return chains, nil
}

// GetByID implements chain.Repository.
func (r *Repository) GetByID(ctx context.Context, id chain.ID) (chain.Chain, error) {
	row, err := sqlc.New(r.pool).GetChain(ctx, string(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return chain.Chain{}, apperr.New(apperr.CodeUnsupportedChain)
		}
		return chain.Chain{}, postgres.MapError(err)
	}
	ch := mapChain(row)
	if !ch.IsSupported {
		return chain.Chain{}, apperr.New(apperr.CodeUnsupportedChain)
	}
	return ch, nil
}

func mapChain(row sqlc.Chain) chain.Chain {
	var nativeAssetID uuid.UUID
	if row.NativeAssetID.Valid {
		nativeAssetID = row.NativeAssetID.UUID
	}
	return chain.Chain{
		ID:            chain.ID(row.ID),
		Name:          row.Name,
		NativeAssetID: nativeAssetID,
		AddressFormat: chain.AddressFormat(row.AddressFormat),
		IsSupported:   row.IsSupported,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
}

var _ chain.Repository = (*Repository)(nil)
