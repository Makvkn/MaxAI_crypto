package wallets

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/maxaicrypto/backend/internal/application/sync"
	"github.com/maxaicrypto/backend/internal/application/usage"
	"github.com/maxaicrypto/backend/internal/domain/apperr"
	"github.com/maxaicrypto/backend/internal/domain/chain"
	"github.com/maxaicrypto/backend/internal/domain/shared"
	"github.com/maxaicrypto/backend/internal/domain/wallet"
	"github.com/maxaicrypto/backend/internal/infrastructure/postgres"
)

// App implements Service.
type App struct {
	wallets      wallet.Repository
	syncStates   syncStateRepo
	chains       chain.Repository
	addresses    chain.AddressValidator
	entitlements usage.EntitlementService
	enqueuer     sync.Enqueuer
	tx           *postgres.TxRunner
	freshness    shared.FreshnessThresholds
}

type syncStateRepo interface {
	wallet.SyncStateRepository
	CreateInitial(ctx context.Context, walletID uuid.UUID) (wallet.SyncState, error)
}

// NewApp wires the wallet application service.
func NewApp(
	wallets wallet.Repository,
	syncStates syncStateRepo,
	chains chain.Repository,
	addresses chain.AddressValidator,
	entitlements usage.EntitlementService,
	enqueuer sync.Enqueuer,
	tx *postgres.TxRunner,
	freshness shared.FreshnessThresholds,
) *App {
	return &App{
		wallets:      wallets,
		syncStates:   syncStates,
		chains:       chains,
		addresses:    addresses,
		entitlements: entitlements,
		enqueuer:     enqueuer,
		tx:           tx,
		freshness:    freshness,
	}
}

// Create implements Service.
func (a *App) Create(ctx context.Context, req CreateRequest) (View, error) {
	canCreate, err := a.entitlements.CanCreateWallet(ctx, req.UserID)
	if err != nil {
		return View{}, err
	}
	if !canCreate {
		return View{}, apperr.New(apperr.CodeWalletLimitReached)
	}

	ch, err := a.chains.GetByID(ctx, req.ChainID)
	if err != nil {
		return View{}, err
	}
	_ = ch

	address, err := a.addresses.Normalize(req.ChainID, req.Address)
	if err != nil {
		return View{}, err
	}

	if label, err := normalizeLabel(req.Label); err != nil {
		return View{}, err
	} else {
		req.Label = label
	}

	if _, found, err := a.wallets.FindByAddress(ctx, req.UserID, req.ChainID, address); err != nil {
		return View{}, err
	} else if found {
		return View{}, apperr.New(apperr.CodeValidation).
			WithMessage("This wallet has already been added.").
			WithDetail("fields", map[string]string{"address": "already exists for this chain"})
	}

	var created View
	err = a.tx.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		w, err := a.wallets.Create(ctx, wallet.Wallet{
			UserID:  req.UserID,
			ChainID: req.ChainID,
			Address: address,
			Label:   req.Label,
			Status:  wallet.StatusActive,
		})
		if err != nil {
			return err
		}
		syncState, err := a.syncStates.CreateInitial(ctx, w.ID)
		if err != nil {
			return err
		}
		created = a.toView(w, syncState)
		return nil
	})
	if err != nil {
		return View{}, err
	}

	if err := a.enqueuer.EnqueueInitialSync(ctx, created.Wallet.ID); err != nil {
		return View{}, err
	}

	return created, nil
}

// Get implements Service.
func (a *App) Get(ctx context.Context, userID, walletID uuid.UUID) (View, error) {
	w, syncState, err := a.loadOwned(ctx, userID, walletID)
	if err != nil {
		return View{}, err
	}
	return a.toView(w, syncState), nil
}

// List implements Service.
func (a *App) List(ctx context.Context, userID uuid.UUID, page shared.Cursor, limit int) (shared.Page[View], error) {
	if limit < 1 {
		limit = 1
	}
	rows, err := a.wallets.ListByUser(ctx, userID, page, limit+1)
	if err != nil {
		return shared.Page[View]{}, err
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	ids := make([]uuid.UUID, len(rows))
	for i, w := range rows {
		ids[i] = w.ID
	}
	syncStates, err := a.syncStates.GetMany(ctx, ids)
	if err != nil {
		return shared.Page[View]{}, err
	}

	views := make([]View, len(rows))
	for i, w := range rows {
		syncState, ok := syncStates[w.ID]
		if !ok {
			return shared.Page[View]{}, apperr.New(apperr.CodeInternal)
		}
		views[i] = a.toView(w, syncState)
	}

	var next shared.Cursor
	if hasMore {
		last := rows[len(rows)-1]
		next = shared.NewCursor(last.CreatedAt, last.ID.String())
	}
	return shared.NewPage(views, next, hasMore), nil
}

// Delete implements Service.
func (a *App) Delete(ctx context.Context, userID, walletID uuid.UUID) error {
	if _, _, err := a.loadOwned(ctx, userID, walletID); err != nil {
		return err
	}
	return a.wallets.SoftDelete(ctx, walletID, time.Now().UTC())
}

// RequestSync implements Service.
func (a *App) RequestSync(ctx context.Context, userID, walletID uuid.UUID) (View, error) {
	w, syncState, err := a.loadOwned(ctx, userID, walletID)
	if err != nil {
		return View{}, err
	}
	if syncState.Status == wallet.SyncSyncing {
		return View{}, apperr.New(apperr.CodeWalletSyncInProgress)
	}
	if err := a.enqueuer.EnqueueSync(ctx, walletID, wallet.TriggerManual); err != nil {
		return View{}, err
	}
	return a.toView(w, syncState), nil
}

func (a *App) loadOwned(ctx context.Context, userID, walletID uuid.UUID) (wallet.Wallet, wallet.SyncState, error) {
	w, err := a.wallets.GetByID(ctx, walletID)
	if err != nil {
		if appErr := apperr.From(err); appErr != nil && appErr.Code == apperr.CodeNotFound {
			return wallet.Wallet{}, wallet.SyncState{}, apperr.New(apperr.CodeWalletNotFound)
		}
		return wallet.Wallet{}, wallet.SyncState{}, err
	}
	if w.UserID != userID {
		return wallet.Wallet{}, wallet.SyncState{}, apperr.New(apperr.CodeWalletNotFound)
	}
	syncState, err := a.syncStates.Get(ctx, walletID)
	if err != nil {
		return wallet.Wallet{}, wallet.SyncState{}, err
	}
	return w, syncState, nil
}

func (a *App) toView(w wallet.Wallet, syncState wallet.SyncState) View {
	now := time.Now().UTC()
	if syncState.LastSyncedAt != nil {
		freshness := a.freshness.ClassifyAt(*syncState.LastSyncedAt, now)
		syncState.DataFreshness = &freshness
	}
	return View{Wallet: w, Sync: syncState}
}

func normalizeLabel(label *string) (*string, error) {
	if label == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*label)
	if trimmed == "" {
		return nil, nil
	}
	if utf8.RuneCountInString(trimmed) > 64 {
		return nil, apperr.New(apperr.CodeValidation).
			WithMessage("The wallet label is too long.").
			WithDetail("fields", map[string]string{"label": "must be at most 64 characters"})
	}
	return &trimmed, nil
}

var _ Service = (*App)(nil)
