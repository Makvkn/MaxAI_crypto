package sync

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	appnormalize "github.com/maxaicrypto/backend/internal/application/normalize"
	appportfolio "github.com/maxaicrypto/backend/internal/application/portfolio"
	apppricing "github.com/maxaicrypto/backend/internal/application/pricing"
	"github.com/maxaicrypto/backend/internal/app/config"
	"github.com/maxaicrypto/backend/internal/domain/apperr"
	"github.com/maxaicrypto/backend/internal/domain/asset"
	"github.com/maxaicrypto/backend/internal/domain/position"
	"github.com/maxaicrypto/backend/internal/domain/provider"
	"github.com/maxaicrypto/backend/internal/domain/shared"
	"github.com/maxaicrypto/backend/internal/domain/transaction"
	"github.com/maxaicrypto/backend/internal/domain/wallet"
	"github.com/maxaicrypto/backend/internal/infrastructure/redis"
)

type syncStateRepository interface {
	wallet.SyncStateRepository
	GetRunByJobID(ctx context.Context, jobID string) (wallet.SyncRun, error)
}

// App implements Service.
type App struct {
	wallets      wallet.Repository
	syncStates   syncStateRepository
	positions    position.Repository
	transactions transaction.Repository
	assets       asset.Repository
	snapshots    appportfolio.SnapshotService
	pricing      apppricing.Service
	classifier   transaction.Classifier
	resolver     provider.Resolver
	locker       *redis.Locker
	cfg          config.SyncConfig
}

// NewApp wires the synchronization service.
func NewApp(
	wallets wallet.Repository,
	syncStates syncStateRepository,
	positions position.Repository,
	transactions transaction.Repository,
	assets asset.Repository,
	snapshots appportfolio.SnapshotService,
	pricing apppricing.Service,
	classifier transaction.Classifier,
	resolver provider.Resolver,
	locker *redis.Locker,
	cfg config.SyncConfig,
) *App {
	if classifier == nil {
		classifier = transaction.NewRulesClassifier()
	}
	return &App{
		wallets:      wallets,
		syncStates:   syncStates,
		positions:    positions,
		transactions: transactions,
		assets:       assets,
		snapshots:    snapshots,
		pricing:      pricing,
		classifier:   classifier,
		resolver:     resolver,
		locker:       locker,
		cfg:          cfg,
	}
}

// Run implements Service.
func (a *App) Run(ctx context.Context, req Request) (Result, error) {
	lock, err := a.locker.Acquire(ctx, redis.WalletSyncLockKey(req.WalletID.String()), a.cfg.LockTTL)
	if errors.Is(err, redis.ErrLockHeld) {
		return Result{}, nil
	}
	if err != nil {
		return Result{}, apperr.Wrap(apperr.CodeInternal, err)
	}
	defer lock.Release(ctx)

	w, err := a.wallets.GetByID(ctx, req.WalletID)
	if err != nil {
		return Result{}, err
	}

	run, err := a.ensureRun(ctx, req)
	if err != nil {
		return Result{}, err
	}

	state, err := a.syncStates.Get(ctx, req.WalletID)
	if err != nil {
		return Result{}, err
	}
	if state.Status == wallet.SyncSyncing && state.SyncJobID != nil && *state.SyncJobID != req.JobID {
		return Result{}, nil
	}

	now := time.Now().UTC()
	state = a.beginSync(state, req.JobID, now)
	if err := a.syncStates.Save(ctx, state); err != nil {
		return Result{}, err
	}

	result, runErr := a.execute(ctx, w, &run, state)
	finishedAt := time.Now().UTC()
	run.Status = result.Status
	run.FinishedAt = &finishedAt
	if runErr != nil {
		code := string(apperr.CodeWalletSyncFailed)
		msg := runErr.Error()
		if appErr := apperr.From(runErr); appErr != nil {
			code = string(appErr.Code)
			msg = appErr.Message
		}
		run.ErrorCode = &code
		run.ErrorText = &msg
	}
	if err := a.syncStates.FinishRun(ctx, run); err != nil {
		return result, err
	}
	return result, runErr
}

// ListDue implements Service.
func (a *App) ListDue(ctx context.Context, limit int) ([]uuid.UUID, error) {
	olderThan := time.Now().UTC().Add(-a.cfg.Interval)
	wallets, err := a.wallets.ListDueForSync(ctx, olderThan, limit)
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, len(wallets))
	for i, w := range wallets {
		ids[i] = w.ID
	}
	return ids, nil
}

func (a *App) ensureRun(ctx context.Context, req Request) (wallet.SyncRun, error) {
	run, err := a.syncStates.GetRunByJobID(ctx, req.JobID)
	if err == nil {
		return run, nil
	}
	if appErr := apperr.From(err); appErr == nil || appErr.Code != apperr.CodeNotFound {
		return wallet.SyncRun{}, err
	}
	return a.syncStates.StartRun(ctx, wallet.SyncRun{
		WalletID: req.WalletID,
		JobID:    req.JobID,
		Trigger:  req.Trigger,
	})
}

func (a *App) beginSync(state wallet.SyncState, jobID string, now time.Time) wallet.SyncState {
	if state.Status != wallet.SyncSyncing {
		state.Status = wallet.SyncSyncing
	}
	state.Stage = nil
	state.StagesCompleted = nil
	state.StartedAt = &now
	state.CompletedAt = nil
	state.ErrorCode = nil
	state.ErrorMessage = nil
	state.SyncJobID = &jobID
	return state
}

func (a *App) execute(ctx context.Context, w wallet.Wallet, run *wallet.SyncRun, state wallet.SyncState) (Result, error) {
	var (
		balancesFetched     int
		transactionsFetched int
		assetIDs            []uuid.UUID
		partial             bool
	)

	if err := a.enterStage(ctx, &state, wallet.StageFetchingBalances); err != nil {
		return a.fail(ctx, w.ID, state, err)
	}
	balances, providerName, err := a.fetchBalances(ctx, w)
	if err != nil {
		return a.fail(ctx, w.ID, state, err)
	}
	balancesFetched = len(balances)
	run.Provider = &providerName

	positions := make([]position.WalletPosition, 0, len(balances))
	observedAt := time.Now().UTC()
	for _, balance := range balances {
		ast, err := a.assets.Upsert(ctx, appnormalize.AssetFromBalance(balance))
		if err != nil {
			return a.fail(ctx, w.ID, state, err)
		}
		assetIDs = append(assetIDs, ast.ID)
		positions = append(positions, position.WalletPosition{
			WalletID:          w.ID,
			AssetID:           ast.ID,
			BalanceRaw:        balance.BalanceRaw,
			BalanceNormalized: balance.BalanceNormalized,
			UpdatedAt:         observedAt,
		})
	}
	if err := a.positions.ReplaceForWallet(ctx, w.ID, positions, observedAt); err != nil {
		return a.fail(ctx, w.ID, state, err)
	}

	if err := a.completeStage(ctx, &state, wallet.StageFetchingBalances); err != nil {
		return a.fail(ctx, w.ID, state, err)
	}

	if err := a.enterStage(ctx, &state, wallet.StageFetchingTransactions); err != nil {
		return a.fail(ctx, w.ID, state, err)
	}
	page, err := a.fetchTransactions(ctx, w)
	if err != nil {
		partial = true
	} else {
		txs, err := a.persistTransactions(ctx, w, page.Transactions)
		if err != nil {
			return a.fail(ctx, w.ID, state, err)
		}
		transactionsFetched = txs
	}

	if err := a.completeStage(ctx, &state, wallet.StageFetchingTransactions); err != nil {
		return a.fail(ctx, w.ID, state, err)
	}

	if err := a.enterStage(ctx, &state, wallet.StageNormalizingAssets); err != nil {
		return a.fail(ctx, w.ID, state, err)
	}
	for _, assetID := range assetIDs {
		ast, err := a.assets.GetByID(ctx, assetID)
		if err != nil {
			return a.fail(ctx, w.ID, state, err)
		}
		if _, err := a.pricing.ResolveMapping(ctx, ast); err != nil {
			partial = true
		}
	}
	if err := a.completeStage(ctx, &state, wallet.StageNormalizingAssets); err != nil {
		return a.fail(ctx, w.ID, state, err)
	}

	if err := a.enterStage(ctx, &state, wallet.StageFetchingPrices); err != nil {
		return a.fail(ctx, w.ID, state, err)
	}
	if _, err := a.pricing.Refresh(ctx, assetIDs); err != nil {
		partial = true
	}

	if err := a.completeStage(ctx, &state, wallet.StageFetchingPrices); err != nil {
		return a.fail(ctx, w.ID, state, err)
	}

	if err := a.enterStage(ctx, &state, wallet.StageCalculatingPortfolio); err != nil {
		return a.fail(ctx, w.ID, state, err)
	}
	var snapshotID *uuid.UUID
	snapshot, err := a.snapshots.Create(ctx, w.ID, &run.ID, time.Now().UTC())
	if err != nil {
		partial = true
	} else {
		snapshotID = &snapshot.ID
	}

	if err := a.completeStage(ctx, &state, wallet.StageCalculatingPortfolio); err != nil {
		return a.fail(ctx, w.ID, state, err)
	}

	if err := a.enterStage(ctx, &state, wallet.StagePreparingAnalysis); err != nil {
		return a.fail(ctx, w.ID, state, err)
	}
	if err := a.completeStage(ctx, &state, wallet.StagePreparingAnalysis); err != nil {
		return a.fail(ctx, w.ID, state, err)
	}

	status := wallet.SyncReady
	if partial {
		status = wallet.SyncPartial
	}
	return a.complete(ctx, w.ID, state, Result{
		Status:              status,
		SnapshotID:          snapshotID,
		BalancesFetched:     balancesFetched,
		TransactionsFetched: transactionsFetched,
	}), nil
}

func (a *App) fetchBalances(ctx context.Context, w wallet.Wallet) ([]provider.NormalizedBalance, string, error) {
	providers, err := a.resolver.ResolveBlockchainChain(ctx, w.ChainID, provider.CapabilityBalances)
	if err != nil {
		return nil, "", err
	}
	req := provider.BalanceRequest{ChainID: w.ChainID, Address: w.Address}
	var lastErr error
	for _, p := range providers {
		balances, err := p.GetBalances(ctx, req)
		if err == nil {
			return balances, string(p.Name()), nil
		}
		lastErr = err
	}
	if lastErr != nil {
		return nil, "", lastErr
	}
	return nil, "", apperr.New(apperr.CodeProviderError)
}

func (a *App) fetchTransactions(ctx context.Context, w wallet.Wallet) (provider.TransactionPage, error) {
	providers, err := a.resolver.ResolveBlockchainChain(ctx, w.ChainID, provider.CapabilityTransactions)
	if err != nil {
		return provider.TransactionPage{}, err
	}
	req := provider.TransactionRequest{ChainID: w.ChainID, Address: w.Address, Limit: 100}
	var lastErr error
	for _, p := range providers {
		page, err := p.GetTransactions(ctx, req)
		if err == nil {
			return page, nil
		}
		lastErr = err
	}
	if lastErr != nil {
		return provider.TransactionPage{}, lastErr
	}
	return provider.TransactionPage{}, apperr.New(apperr.CodeProviderError)
}

func (a *App) persistTransactions(ctx context.Context, w wallet.Wallet, normalized []provider.NormalizedTransaction) (int, error) {
	records := make([]transaction.Transaction, 0, len(normalized))
	for _, item := range normalized {
		for i, transfer := range item.Transfers {
			ast, err := a.assets.Upsert(ctx, appnormalize.AssetFromTransfer(w.ChainID, transfer.Metadata, transfer.AssetIdentity))
			if err != nil {
				return 0, err
			}
			var assetInID, assetOutID *uuid.UUID
			var amountIn, amountOut shared.NullDecimal
			if transfer.Direction == provider.DirectionIn {
				assetInID = &ast.ID
				amountIn = shared.Known(transfer.Amount)
			} else {
				assetOutID = &ast.ID
				amountOut = shared.Known(transfer.Amount)
			}
			records = append(records, appnormalize.Transaction(
				w.ID, w.Address, item, appnormalize.TransferLogIndex(i),
				assetInID, amountIn, assetOutID, amountOut, nil, shared.Unknown(),
			))
		}
	}
	for i := range records {
		classified, err := a.classifier.Classify(ctx, records[i])
		if err != nil {
			return 0, err
		}
		records[i].Type = classified
	}
	return a.transactions.UpsertBatch(ctx, records)
}

func (a *App) enterStage(ctx context.Context, state *wallet.SyncState, stage wallet.SyncStage) error {
	state.Stage = &stage
	return a.syncStates.Save(ctx, *state)
}

func (a *App) completeStage(ctx context.Context, state *wallet.SyncState, stage wallet.SyncStage) error {
	state.Stage = nil
	if !containsStage(state.StagesCompleted, stage) {
		state.StagesCompleted = append(state.StagesCompleted, stage)
	}
	return a.syncStates.Save(ctx, *state)
}

func (a *App) fail(ctx context.Context, walletID uuid.UUID, state wallet.SyncState, err error) (Result, error) {
	code := string(apperr.CodeWalletSyncFailed)
	message := err.Error()
	if appErr := apperr.From(err); appErr != nil {
		code = string(appErr.Code)
		message = appErr.Message
	}
	now := time.Now().UTC()
	state.Status = wallet.SyncFailed
	state.CompletedAt = &now
	state.ErrorCode = &code
	state.ErrorMessage = &message
	state.Stage = nil
	_ = a.syncStates.Save(ctx, state)
	return Result{Status: wallet.SyncFailed}, err
}

func (a *App) complete(ctx context.Context, walletID uuid.UUID, state wallet.SyncState, result Result) Result {
	now := time.Now().UTC()
	state.Status = result.Status
	state.CompletedAt = &now
	state.LastSyncedAt = &now
	state.Stage = nil
	state.ErrorCode = nil
	state.ErrorMessage = nil
	_ = a.syncStates.Save(ctx, state)
	return result
}

func containsStage(stages []wallet.SyncStage, stage wallet.SyncStage) bool {
	for _, s := range stages {
		if s == stage {
			return true
		}
	}
	return false
}

var _ Service = (*App)(nil)
