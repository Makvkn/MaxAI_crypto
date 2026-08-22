package postgres

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/maxaicrypto/backend/internal/domain/apperr"
)

// MapError translates PostgreSQL failures into domain errors. Internal details
// stay in the cause and never reach the client (§28).
func MapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return apperr.New(apperr.CodeNotFound)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			if pgErr.ConstraintName == "auth_identities_provider_subject_key" ||
				pgErr.ConstraintName == "users_email_key" {
				return apperr.New(apperr.CodeEmailAlreadyRegistered)
			}
			if pgErr.ConstraintName == "wallets_active_identity_key" {
				return apperr.New(apperr.CodeValidation).
					WithMessage("This wallet has already been added.").
					WithDetail("fields", map[string]string{"address": "already exists for this chain"})
			}
		}
	}

	return apperr.Wrap(apperr.CodeInternal, err)
}
