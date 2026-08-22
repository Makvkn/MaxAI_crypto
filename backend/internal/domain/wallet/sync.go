package wallet

import (
	"time"

	"github.com/google/uuid"

	"github.com/maxaicrypto/backend/internal/domain/shared"
)

// SyncStatus is the synchronization state machine from §18. STALE is not a
// sync status; freshness is tracked separately.
type SyncStatus string

const (
	SyncPending SyncStatus = "PENDING"
	SyncSyncing SyncStatus = "SYNCING"
	SyncReady   SyncStatus = "READY"
	SyncPartial SyncStatus = "PARTIAL"
	SyncFailed  SyncStatus = "FAILED"
)

// CanTransitionTo reports whether the documented state machine allows moving
// from the current status to next:
//
//	PENDING -> SYNCING -> {READY, PARTIAL, FAILED}
//
// A finished wallet re-enters SYNCING on the next scheduled or manual sync.
func (s SyncStatus) CanTransitionTo(next SyncStatus) bool {
	switch s {
	case SyncPending:
		return next == SyncSyncing
	case SyncSyncing:
		return next == SyncReady || next == SyncPartial || next == SyncFailed
	case SyncReady, SyncPartial, SyncFailed:
		return next == SyncSyncing
	default:
		return false
	}
}

// IsTerminal reports whether the status represents a finished attempt.
func (s SyncStatus) IsTerminal() bool {
	return s == SyncReady || s == SyncPartial || s == SyncFailed
}

// SyncStage is a real, observable step of the synchronization pipeline (§19).
// A stage may only be reported once the backend has actually reached it;
// fabricating progress for UI animation is forbidden.
type SyncStage string

const (
	StageFetchingBalances     SyncStage = "FETCHING_BALANCES"
	StageFetchingTransactions SyncStage = "FETCHING_TRANSACTIONS"
	StageNormalizingAssets    SyncStage = "NORMALIZING_ASSETS"
	StageFetchingPrices       SyncStage = "FETCHING_PRICES"
	StageCalculatingPortfolio SyncStage = "CALCULATING_PORTFOLIO"
	StagePreparingAnalysis    SyncStage = "PREPARING_ANALYSIS"
)

// StageOrder is the pipeline order used to derive the completed-stage list.
// The spelling of the last two stages follows the frontend contract; see
// openapi/DECISIONS.md for the divergence from the §19 examples.
var StageOrder = []SyncStage{
	StageFetchingBalances,
	StageFetchingTransactions,
	StageNormalizingAssets,
	StageFetchingPrices,
	StageCalculatingPortfolio,
	StagePreparingAnalysis,
}

// SyncState is the persisted synchronization state of one wallet.
type SyncState struct {
	WalletID uuid.UUID
	Status   SyncStatus
	// Stage is the stage currently in progress, or nil when no sync is running.
	Stage           *SyncStage
	StagesCompleted []SyncStage
	StartedAt       *time.Time
	CompletedAt     *time.Time
	LastSyncedAt    *time.Time
	// DataFreshness is derived from LastSyncedAt at read time; it is stored on
	// the state only for convenience of the API layer.
	DataFreshness *shared.DataFreshness
	// ErrorCode carries the domain-level failure reason. Provider errors are
	// mapped before they reach this field (§28).
	ErrorCode    *string
	ErrorMessage *string
	// SyncJobID correlates the state with worker logs and metrics (§122).
	SyncJobID *string
	UpdatedAt time.Time
}

// SyncRun records one synchronization attempt for observability. Every
// background operation must be traceable to a job, wallet, provider and
// outcome (§122).
type SyncRun struct {
	ID         uuid.UUID
	WalletID   uuid.UUID
	JobID      string
	Trigger    SyncTrigger
	Provider   *string
	Status     SyncStatus
	StartedAt  time.Time
	FinishedAt *time.Time
	ErrorCode  *string
	ErrorText  *string
}

// SyncTrigger records what caused a synchronization to run.
type SyncTrigger string

const (
	TriggerInitial   SyncTrigger = "INITIAL"
	TriggerScheduled SyncTrigger = "SCHEDULED"
	TriggerManual    SyncTrigger = "MANUAL"
)
