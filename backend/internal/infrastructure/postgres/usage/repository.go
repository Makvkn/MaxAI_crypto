// Package usagerepo implements AI usage persistence with sqlc.
package usagerepo

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/maxaicrypto/backend/internal/domain/subscription"
	"github.com/maxaicrypto/backend/internal/domain/usage"
	"github.com/maxaicrypto/backend/internal/generated/sqlc"
	"github.com/maxaicrypto/backend/internal/infrastructure/postgres"
)

// Repository implements usage.Repository.
type Repository struct {
	pool *postgres.Pool
	tx   *postgres.TxRunner
}

// NewRepository builds an AI usage repository.
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

// GetDaily implements usage.Repository.
func (r *Repository) GetDaily(ctx context.Context, userID uuid.UUID, day time.Time) (usage.Daily, error) {
	row, err := r.queries(ctx).GetAIUsage(ctx, sqlc.GetAIUsageParams{
		UserID:    userID,
		UsageDate: utcDay(day),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return usage.Daily{UserID: userID, Date: utcDay(day)}, nil
		}
		return usage.Daily{}, postgres.MapError(err)
	}
	return mapDaily(row), nil
}

// RecordConsumption implements usage.Repository.
func (r *Repository) RecordConsumption(ctx context.Context, userID uuid.UUID, day time.Time, op usage.Operation, idempotencyKey string) error {
	usageDay := utcDay(day)
	existing, err := r.queries(ctx).GetAIUsageOperationByKey(ctx, idempotencyKey)
	if err == nil {
		_ = existing
		return nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return postgres.MapError(err)
	}

	return r.tx.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		q := sqlc.New(tx)
		current, err := q.GetAIUsage(ctx, sqlc.GetAIUsageParams{
			UserID:    userID,
			UsageDate: usageDay,
		})
		used := 1
		plan := string(subscription.PlanFree)
		if err == nil {
			used = int(current.Used) + 1
			plan = current.Plan
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return postgres.MapError(err)
		}

		if _, err := q.UpsertAIUsage(ctx, sqlc.UpsertAIUsageParams{
			UserID:    userID,
			UsageDate: usageDay,
			Used:      int32(used),
			Plan:      plan,
		}); err != nil {
			return postgres.MapError(err)
		}
		if _, err := q.InsertAIUsageOperation(ctx, sqlc.InsertAIUsageOperationParams{
			UserID:         userID,
			UsageDate:      usageDay,
			Operation:      string(op),
			IdempotencyKey: idempotencyKey,
		}); err != nil {
			return postgres.MapError(err)
		}
		return nil
	})
}

func mapDaily(row sqlc.AiUsage) usage.Daily {
	return usage.Daily{
		UserID:    row.UserID,
		Date:      row.UsageDate,
		Used:      int(row.Used),
		Plan:      subscription.Plan(row.Plan),
		UpdatedAt: row.UpdatedAt,
	}
}

func utcDay(day time.Time) time.Time {
	utc := day.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}

var _ usage.Repository = (*Repository)(nil)
