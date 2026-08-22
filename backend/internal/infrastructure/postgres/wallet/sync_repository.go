// Package walletrepo implements wallet sync state persistence with sqlc.
package walletrepo

import (
	"context"

	"github.com/google/uuid"

	"github.com/maxaicrypto/backend/internal/domain/apperr"
	"github.com/maxaicrypto/backend/internal/domain/wallet"
	"github.com/maxaicrypto/backend/internal/generated/sqlc"
	"github.com/maxaicrypto/backend/internal/infrastructure/postgres"
)

// SyncRepository implements wallet.SyncStateRepository.
type SyncRepository struct {
	pool *postgres.Pool
}

// NewSyncRepository builds a sync state repository.
func NewSyncRepository(pool *postgres.Pool) *SyncRepository {
	return &SyncRepository{pool: pool}
}

func (r *SyncRepository) queries(ctx context.Context) *sqlc.Queries {
	if tx, ok := postgres.TxFrom(ctx); ok {
		return sqlc.New(tx)
	}
	return sqlc.New(r.pool)
}

// Get implements wallet.SyncStateRepository.
func (r *SyncRepository) Get(ctx context.Context, walletID uuid.UUID) (wallet.SyncState, error) {
	row, err := r.queries(ctx).GetWalletSyncState(ctx, walletID)
	if err != nil {
		return wallet.SyncState{}, postgres.MapError(err)
	}
	return mapSyncState(row), nil
}

// GetMany implements wallet.SyncStateRepository.
func (r *SyncRepository) GetMany(ctx context.Context, walletIDs []uuid.UUID) (map[uuid.UUID]wallet.SyncState, error) {
	if len(walletIDs) == 0 {
		return map[uuid.UUID]wallet.SyncState{}, nil
	}
	rows, err := r.queries(ctx).ListWalletSyncStates(ctx, walletIDs)
	if err != nil {
		return nil, postgres.MapError(err)
	}
	states := make(map[uuid.UUID]wallet.SyncState, len(rows))
	for _, row := range rows {
		states[row.WalletID] = mapSyncState(row)
	}
	return states, nil
}

// Save implements wallet.SyncStateRepository.
func (r *SyncRepository) Save(ctx context.Context, state wallet.SyncState) error {
	current, err := r.Get(ctx, state.WalletID)
	if err != nil {
		if !apperr.Is(err, apperr.CodeNotFound) {
			return err
		}
	} else if !current.Status.CanTransitionTo(state.Status) && current.Status != state.Status {
		return apperr.New(apperr.CodeValidation).
			WithMessage("The wallet synchronization state cannot be updated in this way.")
	}

	var stage *string
	if state.Stage != nil {
		s := string(*state.Stage)
		stage = &s
	}
	completed := make([]string, len(state.StagesCompleted))
	for i, s := range state.StagesCompleted {
		completed[i] = string(s)
	}

	_, err = r.queries(ctx).UpdateWalletSyncState(ctx, sqlc.UpdateWalletSyncStateParams{
		WalletID:        state.WalletID,
		Status:          string(state.Status),
		Stage:           stage,
		StagesCompleted: completed,
		StartedAt:       state.StartedAt,
		CompletedAt:     state.CompletedAt,
		LastSyncedAt:    state.LastSyncedAt,
		ErrorCode:       state.ErrorCode,
		ErrorMessage:    state.ErrorMessage,
		SyncJobID:       state.SyncJobID,
	})
	if err != nil {
		return postgres.MapError(err)
	}
	return nil
}

// CreateInitial inserts the PENDING sync state for a new wallet.
func (r *SyncRepository) CreateInitial(ctx context.Context, walletID uuid.UUID) (wallet.SyncState, error) {
	row, err := r.queries(ctx).CreateWalletSyncState(ctx, walletID)
	if err != nil {
		return wallet.SyncState{}, postgres.MapError(err)
	}
	return mapSyncState(row), nil
}

// StartRun implements wallet.SyncStateRepository.
func (r *SyncRepository) StartRun(ctx context.Context, run wallet.SyncRun) (wallet.SyncRun, error) {
	row, err := r.queries(ctx).StartWalletSyncRun(ctx, sqlc.StartWalletSyncRunParams{
		WalletID: run.WalletID,
		JobID:    run.JobID,
		Trigger:  string(run.Trigger),
		Provider: run.Provider,
	})
	if err != nil {
		return wallet.SyncRun{}, postgres.MapError(err)
	}
	return mapSyncRun(row), nil
}

// GetRunByJobID returns the sync run recorded for a background job ID.
func (r *SyncRepository) GetRunByJobID(ctx context.Context, jobID string) (wallet.SyncRun, error) {
	row, err := r.queries(ctx).GetWalletSyncRunByJobID(ctx, jobID)
	if err != nil {
		return wallet.SyncRun{}, postgres.MapError(err)
	}
	return mapSyncRun(row), nil
}

// FinishRun implements wallet.SyncStateRepository.
func (r *SyncRepository) FinishRun(ctx context.Context, run wallet.SyncRun) error {
	if err := r.queries(ctx).FinishWalletSyncRun(ctx, sqlc.FinishWalletSyncRunParams{
		ID:         run.ID,
		Status:     string(run.Status),
		FinishedAt: run.FinishedAt,
		ErrorCode:  run.ErrorCode,
		ErrorText:  run.ErrorText,
	}); err != nil {
		return postgres.MapError(err)
	}
	return nil
}

func mapSyncState(row sqlc.WalletSyncState) wallet.SyncState {
	state := wallet.SyncState{
		WalletID:        row.WalletID,
		Status:          wallet.SyncStatus(row.Status),
		StagesCompleted: make([]wallet.SyncStage, len(row.StagesCompleted)),
		StartedAt:       row.StartedAt,
		CompletedAt:     row.CompletedAt,
		LastSyncedAt:    row.LastSyncedAt,
		ErrorCode:       row.ErrorCode,
		ErrorMessage:    row.ErrorMessage,
		SyncJobID:       row.SyncJobID,
		UpdatedAt:       row.UpdatedAt,
	}
	if row.Stage != nil {
		stage := wallet.SyncStage(*row.Stage)
		state.Stage = &stage
	}
	for i, s := range row.StagesCompleted {
		state.StagesCompleted[i] = wallet.SyncStage(s)
	}
	return state
}

func mapSyncRun(row sqlc.WalletSyncRun) wallet.SyncRun {
	return wallet.SyncRun{
		ID:         row.ID,
		WalletID:   row.WalletID,
		JobID:      row.JobID,
		Trigger:    wallet.SyncTrigger(row.Trigger),
		Provider:   row.Provider,
		Status:     wallet.SyncStatus(row.Status),
		StartedAt:  row.StartedAt,
		FinishedAt: row.FinishedAt,
		ErrorCode:  row.ErrorCode,
		ErrorText:  row.ErrorText,
	}
}

var _ wallet.SyncStateRepository = (*SyncRepository)(nil)
