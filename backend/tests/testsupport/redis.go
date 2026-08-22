//go:build integration

package testsupport

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/maxaicrypto/backend/internal/app/config"
	"github.com/maxaicrypto/backend/internal/infrastructure/redis"
)

const defaultTimeout = 10 * time.Second

// Redis connects to the test Redis instance and flushes it, so locks, quota
// counters and cached values never leak between tests.
func Redis(t *testing.T) *redis.Client {
	t.Helper()

	url := os.Getenv("TEST_REDIS_URL")
	if url == "" {
		url = os.Getenv("REDIS_URL")
	}
	if url == "" {
		t.Skip("TEST_REDIS_URL is not set; run `make up` and `make test-integration`")
	}

	ctx := context.Background()
	client, err := redis.NewClient(ctx, config.RedisConfig{
		URL:          url,
		DialTimeout:  defaultTimeout,
		ReadTimeout:  defaultTimeout,
		WriteTimeout: defaultTimeout,
	})
	if err != nil {
		t.Fatalf("open test redis: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	if err := client.FlushDB(ctx).Err(); err != nil {
		t.Fatalf("flush test redis: %v", err)
	}
	return client
}
