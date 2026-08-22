package config

import (
	"strings"
	"testing"
	"time"
)

func setMinimalEnv(t *testing.T) {
	t.Helper()
	t.Setenv("DATABASE_URL", "postgres://maxai:maxai@localhost:5432/maxai?sslmode=disable")
	t.Setenv("REDIS_URL", "redis://localhost:6379/0")
	t.Setenv("AUTH_JWT_SECRET", strings.Repeat("s", 32))
}

func TestLoadAppliesDocumentedDefaults(t *testing.T) {
	setMinimalEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned an error: %v", err)
	}

	if cfg.Sync.Interval != 15*time.Minute {
		t.Errorf("sync interval = %v, want 15m (spec §62)", cfg.Sync.Interval)
	}
	if cfg.AI.DailyLimit != 10 {
		t.Errorf("AI daily limit = %d, want 10 (spec §86)", cfg.AI.DailyLimit)
	}
	want := []time.Duration{30 * time.Second, time.Minute, 5 * time.Minute}
	if len(cfg.Provider.BackoffSchedule) != len(want) {
		t.Fatalf("backoff schedule = %v, want %v (spec §29)", cfg.Provider.BackoffSchedule, want)
	}
	for i, d := range want {
		if cfg.Provider.BackoffSchedule[i] != d {
			t.Errorf("backoff[%d] = %v, want %v", i, cfg.Provider.BackoffSchedule[i], d)
		}
	}
}

func TestLoadReportsEveryProblemAtOnce(t *testing.T) {
	t.Setenv("AUTH_JWT_SECRET", "too-short")
	t.Setenv("HTTP_PORT", "not-a-number")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() succeeded, want an error")
	}
	for _, key := range []string{"DATABASE_URL", "REDIS_URL", "AUTH_JWT_SECRET", "HTTP_PORT"} {
		if !strings.Contains(err.Error(), key) {
			t.Errorf("error does not mention %s: %v", key, err)
		}
	}
}

func TestLoadRejectsNonIncreasingFreshnessThresholds(t *testing.T) {
	setMinimalEnv(t)
	t.Setenv("FRESHNESS_FRESH_MAX", "20m")
	t.Setenv("FRESHNESS_RECENT_MAX", "15m")

	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted overlapping freshness thresholds, want an error")
	}
}

func TestProductionRequiresProviderSecrets(t *testing.T) {
	setMinimalEnv(t)
	t.Setenv("APP_ENV", "production")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() succeeded without provider secrets in production")
	}
	if !strings.Contains(err.Error(), "OPENAI_API_KEY") {
		t.Errorf("error does not mention OPENAI_API_KEY: %v", err)
	}
}
