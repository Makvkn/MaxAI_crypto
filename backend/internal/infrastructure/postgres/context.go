package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type txKey struct{}

// WithTx stores an open transaction in ctx so repositories can participate in
// a caller-controlled unit of work.
func WithTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// TxFrom returns the transaction bound to ctx, if any.
func TxFrom(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	return tx, ok
}
