// Package subscriptionrepo implements subscription persistence.
package subscriptionrepo

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/maxaicrypto/backend/internal/domain/apperr"
	"github.com/maxaicrypto/backend/internal/domain/subscription"
	"github.com/maxaicrypto/backend/internal/generated/sqlc"
	"github.com/maxaicrypto/backend/internal/infrastructure/postgres"
)

// Repository implements subscription.Repository.
type Repository struct {
	pool *postgres.Pool
}

// NewRepository builds a subscription repository.
func NewRepository(pool *postgres.Pool) *Repository {
	return &Repository{pool: pool}
}

// GetByUser implements subscription.Repository.
func (r *Repository) GetByUser(ctx context.Context, userID uuid.UUID) (subscription.Subscription, bool, error) {
	row, err := sqlc.New(r.pool).GetSubscriptionByUser(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return subscription.Subscription{}, false, nil
		}
		return subscription.Subscription{}, false, postgres.MapError(err)
	}
	return mapSubscription(row), true, nil
}

// Upsert implements subscription.Repository.
func (r *Repository) Upsert(ctx context.Context, s subscription.Subscription) (subscription.Subscription, error) {
	_ = ctx
	_ = s
	return subscription.Subscription{}, apperr.ErrNotImplemented
}

func mapSubscription(row sqlc.Subscription) subscription.Subscription {
	return subscription.Subscription{
		ID:               row.ID,
		UserID:           row.UserID,
		Plan:             subscription.Plan(row.Plan),
		Status:           subscription.Status(row.Status),
		CurrentPeriodEnd: row.CurrentPeriodEnd,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
	}
}

var _ subscription.Repository = (*Repository)(nil)
