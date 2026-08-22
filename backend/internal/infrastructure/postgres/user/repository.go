// Package userrepo implements user and identity persistence with sqlc.
package userrepo

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/maxaicrypto/backend/internal/domain/user"
	"github.com/maxaicrypto/backend/internal/generated/sqlc"
	"github.com/maxaicrypto/backend/internal/infrastructure/postgres"
)

// Repository implements user.Repository.
type Repository struct {
	pool *postgres.Pool
	tx   *postgres.TxRunner
}

// NewRepository builds a user repository.
func NewRepository(pool *postgres.Pool, tx *postgres.TxRunner) *Repository {
	return &Repository{pool: pool, tx: tx}
}

// CreateGuest implements user.Repository.
func (r *Repository) CreateGuest(ctx context.Context, subject string) (user.User, error) {
	var created user.User
	err := r.tx.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		q := sqlc.New(tx)
		row, err := q.CreateUser(ctx, sqlc.CreateUserParams{
			Kind:        string(user.KindGuest),
			Email:       nil,
			DisplayName: nil,
		})
		if err != nil {
			return postgres.MapError(err)
		}
		if _, err := q.CreateIdentity(ctx, sqlc.CreateIdentityParams{
			UserID:        row.ID,
			Provider:      string(user.ProviderGuest),
			Subject:       subject,
			Email:         nil,
			PasswordHash:  nil,
			EmailVerified: false,
		}); err != nil {
			return postgres.MapError(err)
		}
		if _, err := q.CreateSubscription(ctx, sqlc.CreateSubscriptionParams{
			UserID: row.ID,
			Plan:   "FREE",
			Status: "ACTIVE",
		}); err != nil {
			return postgres.MapError(err)
		}
		created = mapUser(row)
		return nil
	})
	return created, err
}

// CreateRegistered provisions a registered account with its first identity.
func (r *Repository) CreateRegistered(ctx context.Context, u user.User, identity user.Identity) (user.User, error) {
	var created user.User
	err := r.tx.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		q := sqlc.New(tx)
		row, err := q.CreateUser(ctx, sqlc.CreateUserParams{
			Kind:        string(user.KindRegistered),
			Email:       u.Email,
			DisplayName: u.DisplayName,
		})
		if err != nil {
			return postgres.MapError(err)
		}
		if _, err := q.CreateIdentity(ctx, sqlc.CreateIdentityParams{
			UserID:        row.ID,
			Provider:      string(identity.Provider),
			Subject:       identity.Subject,
			Email:         identity.Email,
			PasswordHash:  identity.PasswordHash,
			EmailVerified: identity.EmailVerified,
		}); err != nil {
			return postgres.MapError(err)
		}
		if _, err := q.CreateSubscription(ctx, sqlc.CreateSubscriptionParams{
			UserID: row.ID,
			Plan:   "FREE",
			Status: "ACTIVE",
		}); err != nil {
			return postgres.MapError(err)
		}
		created = mapUser(row)
		return nil
	})
	return created, err
}

// GetByID implements user.Repository.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (user.User, error) {
	row, err := sqlc.New(r.pool).GetUserByID(ctx, id)
	if err != nil {
		return user.User{}, postgres.MapError(err)
	}
	return mapUser(row), nil
}

// FindByIdentity implements user.Repository.
func (r *Repository) FindByIdentity(ctx context.Context, provider user.AuthProvider, subject string) (user.User, bool, error) {
	identity, err := sqlc.New(r.pool).GetIdentityByProviderSubject(ctx, sqlc.GetIdentityByProviderSubjectParams{
		Provider: string(provider),
		Subject:  subject,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user.User{}, false, nil
		}
		return user.User{}, false, postgres.MapError(err)
	}
	u, err := r.GetByID(ctx, identity.UserID)
	if err != nil {
		return user.User{}, false, err
	}
	return u, true, nil
}

// FindIdentity implements user.Repository.
func (r *Repository) FindIdentity(ctx context.Context, provider user.AuthProvider, subject string) (user.Identity, bool, error) {
	row, err := sqlc.New(r.pool).GetIdentityByProviderSubject(ctx, sqlc.GetIdentityByProviderSubjectParams{
		Provider: string(provider),
		Subject:  subject,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user.Identity{}, false, nil
		}
		return user.Identity{}, false, postgres.MapError(err)
	}
	return mapIdentity(row), true, nil
}

// Upgrade implements user.Repository.
func (r *Repository) Upgrade(ctx context.Context, userID uuid.UUID, identity user.Identity, displayName *string) (user.User, error) {
	var upgraded user.User
	err := r.tx.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		q := sqlc.New(tx)
		if _, err := q.CreateIdentity(ctx, sqlc.CreateIdentityParams{
			UserID:        userID,
			Provider:      string(identity.Provider),
			Subject:       identity.Subject,
			Email:         identity.Email,
			PasswordHash:  identity.PasswordHash,
			EmailVerified: identity.EmailVerified,
		}); err != nil {
			return postgres.MapError(err)
		}
		row, err := q.UpgradeUser(ctx, sqlc.UpgradeUserParams{
			ID:          userID,
			Email:       identity.Email,
			DisplayName: displayName,
		})
		if err != nil {
			return postgres.MapError(err)
		}
		upgraded = mapUser(row)
		return nil
	})
	return upgraded, err
}

// ListAuthProviders implements user.Repository.
func (r *Repository) ListAuthProviders(ctx context.Context, userID uuid.UUID) ([]user.AuthProvider, error) {
	rows, err := sqlc.New(r.pool).ListAuthProvidersByUser(ctx, userID)
	if err != nil {
		return nil, postgres.MapError(err)
	}
	providers := make([]user.AuthProvider, 0, len(rows))
	for _, row := range rows {
		providers = append(providers, user.AuthProvider(row))
	}
	return providers, nil
}

func mapUser(row sqlc.User) user.User {
	return user.User{
		ID:          row.ID,
		Kind:        user.Kind(row.Kind),
		Email:       row.Email,
		DisplayName: row.DisplayName,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
		DeletedAt:   row.DeletedAt,
	}
}

func mapIdentity(row sqlc.AuthIdentity) user.Identity {
	return user.Identity{
		ID:            row.ID,
		UserID:        row.UserID,
		Provider:      user.AuthProvider(row.Provider),
		Subject:       row.Subject,
		Email:         row.Email,
		PasswordHash:  row.PasswordHash,
		EmailVerified: row.EmailVerified,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
}

var _ user.Repository = (*Repository)(nil)
