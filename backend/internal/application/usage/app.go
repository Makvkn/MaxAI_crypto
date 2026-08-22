package usage

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/maxaicrypto/backend/internal/domain/apperr"
	"github.com/maxaicrypto/backend/internal/domain/subscription"
	"github.com/maxaicrypto/backend/internal/domain/usage"
)

// App implements Service.
type App struct {
	repo         usage.Repository
	counter      usage.Counter
	entitlements EntitlementService
}

// NewApp wires the AI usage service.
func NewApp(repo usage.Repository, counter usage.Counter, entitlements EntitlementService) *App {
	return &App{
		repo:         repo,
		counter:      counter,
		entitlements: entitlements,
	}
}

// Today implements Service.
func (a *App) Today(ctx context.Context, userID uuid.UUID) (usage.Daily, error) {
	now := time.Now().UTC()
	day := utcDay(now)

	ent, err := a.entitlements.Entitlements(ctx, userID)
	if err != nil {
		return usage.Daily{}, err
	}

	usedRedis, err := a.counter.Used(ctx, userID, day)
	if err != nil {
		return usage.Daily{}, err
	}
	stored, err := a.repo.GetDaily(ctx, userID, day)
	if err != nil {
		return usage.Daily{}, err
	}

	used := usedRedis
	if stored.Used > used {
		used = stored.Used
	}

	plan := subscription.PlanFree
	if stored.Plan != "" {
		plan = stored.Plan
	}

	return usage.Daily{
		UserID:    userID,
		Date:      day,
		Used:      used,
		Limit:     ent.AIOperationsPerDay,
		Plan:      plan,
		ResetsAt:  nextUTCMidnight(now),
		UpdatedAt: now,
	}, nil
}

// Reserve implements Service.
func (a *App) Reserve(ctx context.Context, userID uuid.UUID, op usage.Operation, idempotencyKey string) (usage.Reservation, error) {
	if idempotencyKey == "" {
		return usage.Reservation{}, apperr.New(apperr.CodeValidation).
			WithMessage("An idempotency key is required.")
	}

	ent, err := a.entitlements.Entitlements(ctx, userID)
	if err != nil {
		return usage.Reservation{}, err
	}
	if !ent.Allows(subscription.FeatureAI) {
		return usage.Reservation{}, apperr.New(apperr.CodeForbidden)
	}

	reservation, ok, err := a.counter.Reserve(ctx, userID, utcDay(time.Now().UTC()), ent.AIOperationsPerDay)
	if err != nil {
		return usage.Reservation{}, err
	}
	if !ok {
		return usage.Reservation{}, apperr.New(apperr.CodeAIDailyLimit)
	}
	reservation.Operation = op
	reservation.IdempotencyKey = idempotencyKey
	return reservation, nil
}

// Commit implements Service.
func (a *App) Commit(ctx context.Context, reservation usage.Reservation) error {
	if err := a.counter.Commit(ctx, reservation); err != nil {
		return err
	}
	day := utcDay(reservation.ReservedAt)
	return a.repo.RecordConsumption(ctx, reservation.UserID, day, reservation.Operation, reservation.IdempotencyKey)
}

// Release implements Service.
func (a *App) Release(ctx context.Context, reservation usage.Reservation) error {
	return a.counter.Release(ctx, reservation)
}

func utcDay(day time.Time) time.Time {
	utc := day.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}

func nextUTCMidnight(now time.Time) time.Time {
	utc := now.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC).Add(24 * time.Hour)
}

var _ Service = (*App)(nil)
