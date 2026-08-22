// Package assetrepo implements asset persistence with sqlc.
package assetrepo

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/maxaicrypto/backend/internal/domain/asset"
	"github.com/maxaicrypto/backend/internal/domain/chain"
	"github.com/maxaicrypto/backend/internal/generated/sqlc"
	"github.com/maxaicrypto/backend/internal/infrastructure/postgres"
)

// Repository implements asset.Repository.
type Repository struct {
	pool *postgres.Pool
	tx   *postgres.TxRunner
}

// NewRepository builds an asset repository.
func NewRepository(pool *postgres.Pool, tx *postgres.TxRunner) *Repository {
	return &Repository{pool: pool, tx: tx}
}

func (r *Repository) queries(ctx context.Context) *sqlc.Queries {
	if tx, ok := postgres.TxFrom(ctx); ok {
		return sqlc.New(tx)
	}
	return sqlc.New(r.pool)
}

// GetByID implements asset.Repository.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (asset.Asset, error) {
	row, err := r.queries(ctx).GetAssetByID(ctx, id)
	if err != nil {
		return asset.Asset{}, postgres.MapError(err)
	}
	return mapAsset(row), nil
}

// GetManyByID implements asset.Repository.
func (r *Repository) GetManyByID(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]asset.Asset, error) {
	if len(ids) == 0 {
		return map[uuid.UUID]asset.Asset{}, nil
	}
	rows, err := r.queries(ctx).GetAssetsByIDs(ctx, ids)
	if err != nil {
		return nil, postgres.MapError(err)
	}
	assets := make(map[uuid.UUID]asset.Asset, len(rows))
	for _, row := range rows {
		a := mapAsset(row)
		assets[a.ID] = a
	}
	return assets, nil
}

// FindByIdentity implements asset.Repository.
func (r *Repository) FindByIdentity(ctx context.Context, identity asset.Identity) (asset.Asset, bool, error) {
	row, err := r.queries(ctx).FindAssetByIdentity(ctx, sqlc.FindAssetByIdentityParams{
		ChainID:         string(identity.ChainID),
		ContractAddress: identity.ContractAddress,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return asset.Asset{}, false, nil
		}
		return asset.Asset{}, false, postgres.MapError(err)
	}
	return mapAsset(row), true, nil
}

// Upsert implements asset.Repository.
func (r *Repository) Upsert(ctx context.Context, a asset.Asset) (asset.Asset, error) {
	id := a.ID
	if id == uuid.Nil {
		id = uuid.New()
	}

	var provider *string
	if a.MarketDataProvider != nil {
		s := string(*a.MarketDataProvider)
		provider = &s
	}

	row, err := r.queries(ctx).UpsertAsset(ctx, sqlc.UpsertAssetParams{
		ID:                 id,
		ChainID:            string(a.ChainID),
		ContractAddress:    a.ContractAddress,
		Symbol:             a.Symbol,
		Name:               a.Name,
		Decimals:           int32(a.Decimals),
		AssetType:          string(a.Type),
		IconUrl:            a.IconURL,
		MarketDataProvider: provider,
		MarketDataID:       a.MarketDataID,
	})
	if err != nil {
		return asset.Asset{}, postgres.MapError(err)
	}
	return mapAsset(row), nil
}

// SetMarketDataMapping implements asset.Repository.
func (r *Repository) SetMarketDataMapping(ctx context.Context, id uuid.UUID, provider *asset.MarketDataProvider, marketDataID *string) error {
	var providerName *string
	if provider != nil {
		s := string(*provider)
		providerName = &s
	}
	if err := r.queries(ctx).SetAssetMarketDataMapping(ctx, sqlc.SetAssetMarketDataMappingParams{
		ID:                 id,
		MarketDataProvider: providerName,
		MarketDataID:       marketDataID,
	}); err != nil {
		return postgres.MapError(err)
	}
	return nil
}

// ListUnmapped implements asset.Repository.
func (r *Repository) ListUnmapped(ctx context.Context, limit int) ([]asset.Asset, error) {
	rows, err := r.queries(ctx).ListUnmappedAssets(ctx, int32(limit))
	if err != nil {
		return nil, postgres.MapError(err)
	}
	assets := make([]asset.Asset, len(rows))
	for i, row := range rows {
		assets[i] = mapAsset(row)
	}
	return assets, nil
}

func mapAsset(row sqlc.Asset) asset.Asset {
	a := asset.Asset{
		ID:           row.ID,
		ChainID:      chain.ID(row.ChainID),
		ContractAddress: row.ContractAddress,
		Symbol:       row.Symbol,
		Name:         row.Name,
		Decimals:     int(row.Decimals),
		Type:         asset.Type(row.AssetType),
		IconURL:      row.IconUrl,
		MarketDataID: row.MarketDataID,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
	if row.MarketDataProvider != nil {
		provider := asset.MarketDataProvider(*row.MarketDataProvider)
		a.MarketDataProvider = &provider
	}
	return a
}

var _ asset.Repository = (*Repository)(nil)
