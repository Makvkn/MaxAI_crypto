// Package userrepo implements user, identity and refresh-session persistence.
package userrepo

import (
	"context"
	"errors"
	"net/netip"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/maxaicrypto/backend/internal/domain/user"
	"github.com/maxaicrypto/backend/internal/generated/sqlc"
	"github.com/maxaicrypto/backend/internal/infrastructure/postgres"
)

// SessionRepository implements user.SessionRepository.
type SessionRepository struct {
	pool *postgres.Pool
	tx   *postgres.TxRunner
}

// NewSessionRepository builds a refresh-session repository.
func NewSessionRepository(pool *postgres.Pool, tx *postgres.TxRunner) *SessionRepository {
	return &SessionRepository{pool: pool, tx: tx}
}

// Create implements user.SessionRepository.
func (r *SessionRepository) Create(ctx context.Context, session user.RefreshSession) (user.RefreshSession, error) {
	row, err := sqlc.New(r.pool).CreateRefreshSession(ctx, sqlc.CreateRefreshSessionParams{
		UserID:    session.UserID,
		TokenHash: session.TokenHash,
		ExpiresAt: session.ExpiresAt,
		UserAgent: session.UserAgent,
		IpAddress: parseIP(session.IPAddress),
	})
	if err != nil {
		return user.RefreshSession{}, postgres.MapError(err)
	}
	return mapSession(
		row.ID, row.UserID, row.TokenHash, row.IssuedAt, row.ExpiresAt,
		row.RevokedAt, row.RotatedTo, row.UserAgent, row.IpAddress, row.LastUsedAt,
	), nil
}

// FindByTokenHash implements user.SessionRepository.
func (r *SessionRepository) FindByTokenHash(ctx context.Context, tokenHash string) (user.RefreshSession, bool, error) {
	row, err := sqlc.New(r.pool).GetRefreshSessionByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user.RefreshSession{}, false, nil
		}
		return user.RefreshSession{}, false, postgres.MapError(err)
	}
	return mapSession(
		row.ID, row.UserID, row.TokenHash, row.IssuedAt, row.ExpiresAt,
		row.RevokedAt, row.RotatedTo, row.UserAgent, row.IpAddress, row.LastUsedAt,
	), true, nil
}

// Rotate implements user.SessionRepository.
func (r *SessionRepository) Rotate(ctx context.Context, currentID uuid.UUID, next user.RefreshSession) (user.RefreshSession, error) {
	var created user.RefreshSession
	err := r.tx.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		q := sqlc.New(tx)
		row, err := q.CreateRefreshSession(ctx, sqlc.CreateRefreshSessionParams{
			UserID:    next.UserID,
			TokenHash: next.TokenHash,
			ExpiresAt: next.ExpiresAt,
			UserAgent: next.UserAgent,
			IpAddress: parseIP(next.IPAddress),
		})
		if err != nil {
			return postgres.MapError(err)
		}
		if err := q.MarkRefreshSessionRotated(ctx, sqlc.MarkRefreshSessionRotatedParams{
			ID:        currentID,
			RotatedTo: uuid.NullUUID{UUID: row.ID, Valid: true},
		}); err != nil {
			return postgres.MapError(err)
		}
		created = mapSession(
			row.ID, row.UserID, row.TokenHash, row.IssuedAt, row.ExpiresAt,
			row.RevokedAt, row.RotatedTo, row.UserAgent, row.IpAddress, row.LastUsedAt,
		)
		return nil
	})
	return created, err
}

// Revoke implements user.SessionRepository.
func (r *SessionRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	if err := sqlc.New(r.pool).RevokeRefreshSession(ctx, id); err != nil {
		return postgres.MapError(err)
	}
	return nil
}

// RevokeAllForUser implements user.SessionRepository.
func (r *SessionRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	if err := sqlc.New(r.pool).RevokeAllRefreshSessionsForUser(ctx, userID); err != nil {
		return postgres.MapError(err)
	}
	return nil
}

func mapSession(
	id, userID uuid.UUID,
	tokenHash string,
	issuedAt, expiresAt time.Time,
	revokedAt *time.Time,
	rotatedTo uuid.NullUUID,
	userAgent *string,
	ipAddress string,
	lastUsedAt *time.Time,
) user.RefreshSession {
	var ip *string
	if ipAddress != "" {
		ip = &ipAddress
	}
	var rotated *uuid.UUID
	if rotatedTo.Valid {
		value := rotatedTo.UUID
		rotated = &value
	}
	return user.RefreshSession{
		ID:         id,
		UserID:     userID,
		TokenHash:  tokenHash,
		IssuedAt:   issuedAt,
		ExpiresAt:  expiresAt,
		RevokedAt:  revokedAt,
		RotatedTo:  rotated,
		UserAgent:  userAgent,
		IPAddress:  ip,
		LastUsedAt: lastUsedAt,
	}
}

func parseIP(raw *string) *netip.Addr {
	if raw == nil || *raw == "" {
		return nil
	}
	addr, err := netip.ParseAddr(*raw)
	if err != nil {
		return nil
	}
	return &addr
}

var _ user.SessionRepository = (*SessionRepository)(nil)
