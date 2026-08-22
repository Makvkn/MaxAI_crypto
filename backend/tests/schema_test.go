//go:build integration

package tests

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"

	"github.com/maxaicrypto/backend/internal/domain/chain"
	"github.com/maxaicrypto/backend/tests/testsupport"
)

// TestChainsAreSeeded verifies the reference data the application depends on:
// every MVP chain exists and owns a native asset (§20, §31).
func TestChainsAreSeeded(t *testing.T) {
	pool := testsupport.Postgres(t)
	ctx := context.Background()

	for _, id := range chain.Supported {
		var nativeAssetID *string
		err := pool.QueryRow(ctx,
			`SELECT native_asset_id::text FROM chains WHERE id = $1`, string(id),
		).Scan(&nativeAssetID)
		if err != nil {
			t.Fatalf("chain %s is not seeded: %v", id, err)
		}
		if nativeAssetID == nil {
			t.Errorf("chain %s has no native asset", id)
		}
	}
}

// TestNumericKeepsExactPrecision guards the rule that money never round-trips
// through a float (§97, §112). A float64 cannot represent this value exactly.
func TestNumericKeepsExactPrecision(t *testing.T) {
	pool := testsupport.Postgres(t)
	ctx := context.Background()

	want := decimal.RequireFromString("12345678901234567890.123456789012345678")

	var got decimal.Decimal
	if err := pool.QueryRow(ctx, `SELECT $1::numeric`, want).Scan(&got); err != nil {
		t.Fatalf("round-trip numeric: %v", err)
	}
	if !got.Equal(want) {
		t.Fatalf("numeric lost precision: want %s, got %s", want, got)
	}
}

// TestWalletAddressIsUniquePerUser guards the idempotency constraint behind
// wallet creation (§16, §60).
func TestWalletAddressIsUniquePerUser(t *testing.T) {
	pool := testsupport.Postgres(t)
	ctx := context.Background()

	var userID string
	if err := pool.QueryRow(ctx,
		`INSERT INTO users (kind) VALUES ('GUEST') RETURNING id::text`,
	).Scan(&userID); err != nil {
		t.Fatalf("create user: %v", err)
	}

	insert := `INSERT INTO wallets (user_id, chain_id, address, status) VALUES ($1, 'ethereum', '0xabc', 'ACTIVE')`
	if _, err := pool.Exec(ctx, insert, userID); err != nil {
		t.Fatalf("create wallet: %v", err)
	}
	if _, err := pool.Exec(ctx, insert, userID); err == nil {
		t.Fatal("expected the duplicate wallet address to be rejected")
	}
}
