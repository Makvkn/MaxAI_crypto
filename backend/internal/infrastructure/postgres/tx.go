package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBTX is the subset of pgx used by generated queries. Both the pool and an
// open transaction satisfy it, which lets a repository method run either
// standalone or inside a caller-controlled transaction.
type DBTX interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// TxRunner executes a function inside a single database transaction. The
// specification requires explicit transaction boundaries for wallet creation,
// snapshot writes, guest upgrade and AI usage accounting (§117).
type TxRunner struct {
	pool *pgxpool.Pool
}

// NewTxRunner builds a runner bound to the given pool.
func NewTxRunner(pool *Pool) *TxRunner {
	return &TxRunner{pool: pool.Pool}
}

// InTx begins a transaction, invokes fn, and commits when fn returns nil.
// Any error or panic rolls the transaction back.
func (r *TxRunner) InTx(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) (err error) {
	tx, beginErr := r.pool.Begin(ctx)
	if beginErr != nil {
		return fmt.Errorf("begin transaction: %w", beginErr)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(context.WithoutCancel(ctx))
			panic(p)
		}
		if err != nil {
			if rbErr := tx.Rollback(context.WithoutCancel(ctx)); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
				err = errors.Join(err, fmt.Errorf("rollback transaction: %w", rbErr))
			}
		}
	}()

	if err = fn(WithTx(ctx, tx), tx); err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
