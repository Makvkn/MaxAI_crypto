// Package postgres owns the PostgreSQL connection pool and the repository
// implementations. PostgreSQL is the primary source of truth (§6 Principle 6).
package postgres

import (
	"context"
	"fmt"
	"time"

	pgxdecimal "github.com/jackc/pgx-shopspring-decimal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/maxaicrypto/backend/internal/app/config"
)

// Pool wraps pgxpool with the project's configuration and health semantics.
type Pool struct {
	*pgxpool.Pool
}

// NewPool opens the connection pool and verifies connectivity before returning,
// so a misconfigured database fails at startup rather than on first request.
func NewPool(ctx context.Context, cfg config.DatabaseConfig) (*Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse DATABASE_URL: %w", err)
	}

	poolCfg.MaxConns = cfg.MaxConns
	poolCfg.MinConns = cfg.MinConns
	if cfg.MaxConnLifetime > 0 {
		poolCfg.MaxConnLifetime = cfg.MaxConnLifetime
	}
	if cfg.MaxConnIdleTime > 0 {
		poolCfg.MaxConnIdleTime = cfg.MaxConnIdleTime
	}
	poolCfg.ConnConfig.ConnectTimeout = cfg.ConnectTimeout

	// Register the decimal codec so NUMERIC columns round-trip as exact
	// decimals. Without it pgx would fall back to a float representation,
	// which is forbidden for financial values (§111, §112).
	poolCfg.AfterConnect = func(_ context.Context, conn *pgx.Conn) error {
		pgxdecimal.Register(conn.TypeMap())
		return nil
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, cfg.ConnectTimeout)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &Pool{Pool: pool}, nil
}

// Check reports whether the database is reachable. It backs the readiness
// endpoint and must stay cheap.
func (p *Pool) Check(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return p.Ping(ctx)
}
