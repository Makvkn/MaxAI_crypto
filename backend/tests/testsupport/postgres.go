//go:build integration

package testsupport

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/maxaicrypto/backend/internal/app/config"
	"github.com/maxaicrypto/backend/internal/infrastructure/postgres"
)

var (
	migrateOnce sync.Once
	migrateErr  error
)

// DatabaseURL returns the test database DSN, skipping the test when the harness
// is not configured so `go test ./...` stays runnable without Docker.
func DatabaseURL(t *testing.T) string {
	t.Helper()

	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is not set; run `make up` and `make test-integration`")
	}
	return url
}

// Postgres opens a pool against the test database. Migrations run once per test
// process; each test starts from truncated data tables.
func Postgres(t *testing.T) *postgres.Pool {
	t.Helper()

	ctx := context.Background()
	pool, err := postgres.NewPool(ctx, config.DatabaseConfig{
		URL:            DatabaseURL(t),
		MaxConns:       4,
		MinConns:       1,
		ConnectTimeout: defaultTimeout,
	})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(pool.Close)

	migrateOnce.Do(func() {
		migrateErr = ensureMigrations(ctx, pool)
	})
	if migrateErr != nil {
		t.Fatalf("apply migrations: %v", migrateErr)
	}

	Truncate(t, pool)
	return pool
}

func ensureMigrations(ctx context.Context, pool *postgres.Pool) error {
	var initialized bool
	if err := pool.QueryRow(ctx,
		`SELECT to_regclass('public.chains') IS NOT NULL`,
	).Scan(&initialized); err != nil {
		return err
	}
	if initialized {
		return nil
	}
	return applyMigrations(ctx, pool)
}

func applyMigrations(ctx context.Context, pool *postgres.Pool) error {
	dir, err := migrationsDir()
	if err != nil {
		return err
	}

	entries, err := filepath.Glob(filepath.Join(dir, "*.up.sql"))
	if err != nil {
		return err
	}
	sort.Strings(entries)

	for _, path := range entries {
		statements, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if _, err := pool.Exec(ctx, string(statements)); err != nil {
			return err
		}
	}
	return nil
}

// Truncate empties every data table while keeping seeded reference data, so a
// test never observes rows written by its neighbours.
func Truncate(t *testing.T, pool *postgres.Pool) {
	t.Helper()

	tables := []string{
		"conversation_messages",
		"conversations",
		"ai_usage_operations",
		"ai_usage",
		"scenario_calculations",
		"portfolio_snapshot_positions",
		"portfolio_snapshots",
		"transactions",
		"wallet_positions",
		"wallet_sync_runs",
		"wallet_sync_states",
		"wallets",
		"prices",
		"subscriptions",
		"refresh_sessions",
		"auth_identities",
		"users",
	}

	stmt := "TRUNCATE " + strings.Join(tables, ", ") + " RESTART IDENTITY CASCADE"
	if _, err := pool.Exec(context.Background(), stmt); err != nil {
		t.Fatalf("truncate tables: %v", err)
	}
}

func migrationsDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		candidate := filepath.Join(dir, "migrations")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}
