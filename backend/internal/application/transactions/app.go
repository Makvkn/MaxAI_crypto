package transactions

import (
	"context"

	"github.com/google/uuid"

	apppricing "github.com/maxaicrypto/backend/internal/application/pricing"
	"github.com/maxaicrypto/backend/internal/domain/apperr"
	"github.com/maxaicrypto/backend/internal/domain/asset"
	"github.com/maxaicrypto/backend/internal/domain/price"
	"github.com/maxaicrypto/backend/internal/domain/shared"
	"github.com/maxaicrypto/backend/internal/domain/transaction"
	"github.com/maxaicrypto/backend/internal/domain/wallet"
	"github.com/maxaicrypto/backend/internal/infrastructure/chain/explorer"
)

// App implements Service.
type App struct {
	wallets      wallet.Repository
	syncStates   wallet.SyncStateRepository
	transactions transaction.Repository
	assets       asset.Repository
	pricing      apppricing.Service
}

// NewApp wires the transaction read service.
func NewApp(
	wallets wallet.Repository,
	syncStates wallet.SyncStateRepository,
	transactions transaction.Repository,
	assets asset.Repository,
	pricing apppricing.Service,
) *App {
	return &App{
		wallets:      wallets,
		syncStates:   syncStates,
		transactions: transactions,
		assets:       assets,
		pricing:      pricing,
	}
}

// List implements Service.
func (a *App) List(ctx context.Context, userID, walletID uuid.UUID, filter transaction.Filter, page shared.Cursor, limit int) (shared.Page[View], error) {
	if err := a.requireReadyWallet(ctx, userID, walletID); err != nil {
		return shared.Page[View]{}, err
	}
	if limit < 1 {
		limit = 1
	}

	rows, err := a.transactions.ListByWallet(ctx, walletID, filter, page, limit+1)
	if err != nil {
		return shared.Page[View]{}, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	views, err := a.enrichMany(ctx, rows)
	if err != nil {
		return shared.Page[View]{}, err
	}

	var next shared.Cursor
	if hasMore {
		last := rows[len(rows)-1]
		next = shared.NewCursor(last.Timestamp, last.ID.String())
	}
	return shared.NewPage(views, next, hasMore), nil
}

// Get implements Service.
func (a *App) Get(ctx context.Context, userID, walletID, transactionID uuid.UUID) (View, error) {
	if err := a.requireReadyWallet(ctx, userID, walletID); err != nil {
		return View{}, err
	}

	tx, err := a.transactions.GetByID(ctx, transactionID)
	if err != nil {
		if appErr := apperr.From(err); appErr != nil && appErr.Code == apperr.CodeNotFound {
			return View{}, apperr.New(apperr.CodeTransactionNotFound)
		}
		return View{}, err
	}
	if tx.WalletID != walletID {
		return View{}, apperr.New(apperr.CodeTransactionNotFound)
	}

	views, err := a.enrichMany(ctx, []transaction.Transaction{tx})
	if err != nil {
		return View{}, err
	}
	return views[0], nil
}

func (a *App) requireReadyWallet(ctx context.Context, userID, walletID uuid.UUID) error {
	w, err := a.wallets.GetByID(ctx, walletID)
	if err != nil {
		if appErr := apperr.From(err); appErr != nil && appErr.Code == apperr.CodeNotFound {
			return apperr.New(apperr.CodeWalletNotFound)
		}
		return err
	}
	if w.UserID != userID {
		return apperr.New(apperr.CodeWalletNotFound)
	}

	syncState, err := a.syncStates.Get(ctx, walletID)
	if err != nil {
		return err
	}
	switch syncState.Status {
	case wallet.SyncPending, wallet.SyncSyncing:
		return apperr.New(apperr.CodeWalletNotReady).
			WithDetail("sync_status", string(syncState.Status))
	case wallet.SyncFailed:
		return apperr.New(apperr.CodeWalletSyncFailed)
	}
	return nil
}

func (a *App) enrichMany(ctx context.Context, rows []transaction.Transaction) ([]View, error) {
	if len(rows) == 0 {
		return []View{}, nil
	}

	assetIDs := make([]uuid.UUID, 0, len(rows)*3)
	for _, tx := range rows {
		if tx.AssetInID != nil {
			assetIDs = append(assetIDs, *tx.AssetInID)
		}
		if tx.AssetOutID != nil {
			assetIDs = append(assetIDs, *tx.AssetOutID)
		}
		if tx.FeeAssetID != nil {
			assetIDs = append(assetIDs, *tx.FeeAssetID)
		}
	}

	assets, err := a.assets.GetManyByID(ctx, assetIDs)
	if err != nil {
		return nil, err
	}
	prices, err := a.pricing.GetCurrent(ctx, assetIDs)
	if err != nil {
		return nil, err
	}

	views := make([]View, len(rows))
	for i, tx := range rows {
		views[i] = buildView(tx, assets, prices)
	}
	return views, nil
}

func buildView(tx transaction.Transaction, assets map[uuid.UUID]asset.Asset, prices map[uuid.UUID]price.Price) View {
	view := View{
		Transaction: tx,
		ExplorerURL: explorer.TransactionURL(tx.ChainID, tx.TxHash),
	}
	if tx.AssetInID != nil {
		if ast, ok := assets[*tx.AssetInID]; ok {
			copy := ast
			view.AssetIn = &copy
			view.ValueInUSD = valueUSD(tx.AmountIn, prices[*tx.AssetInID])
		}
	}
	if tx.AssetOutID != nil {
		if ast, ok := assets[*tx.AssetOutID]; ok {
			copy := ast
			view.AssetOut = &copy
			view.ValueOutUSD = valueUSD(tx.AmountOut, prices[*tx.AssetOutID])
		}
	}
	if tx.FeeAssetID != nil {
		if ast, ok := assets[*tx.FeeAssetID]; ok {
			copy := ast
			view.FeeAsset = &copy
			view.FeeValueUSD = valueUSD(tx.FeeAmount, prices[*tx.FeeAssetID])
		}
	}
	return view
}

func valueUSD(amount shared.NullDecimal, quote price.Price) shared.NullDecimal {
	if !amount.Valid || !quote.IsUsable() {
		return shared.Unknown()
	}
	value := amount.Decimal.Value().Mul(quote.ValueUSD.Decimal.Value())
	return shared.Known(shared.NewDecimal(value))
}

var _ Service = (*App)(nil)
